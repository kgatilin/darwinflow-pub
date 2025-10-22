package infra

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// DefaultRPCTimeout is the default timeout for RPC calls.
const DefaultRPCTimeout = 30 * time.Second

// RPCClient manages communication with an external plugin process via JSON-RPC.
// It handles subprocess lifecycle, request/response correlation, and event streaming.
type RPCClient struct {
	// executablePath is the path to the plugin executable
	executablePath string

	// args are the command-line arguments to pass to the plugin
	args []string

	// cmd is the subprocess handle
	cmd *exec.Cmd

	// stdin is the pipe to the subprocess stdin
	stdin io.WriteCloser

	// stdout is the pipe from the subprocess stdout
	stdout io.ReadCloser

	// stderr is the pipe from the subprocess stderr
	stderr io.ReadCloser

	// pendingRequests maps request IDs to response channels
	pendingRequests map[interface{}]*rpcPendingRequest
	requestsMu      sync.RWMutex

	// nextRequestID is an atomic counter for generating request IDs
	nextRequestID atomic.Uint64

	// eventChan receives events from the plugin (if event streaming is active)
	eventChan chan<- pluginsdk.Event
	eventMu   sync.Mutex

	// done signals shutdown
	done chan struct{}

	// err stores any fatal error that terminated the client
	err error
	errMu sync.RWMutex

	// ctx is the client lifecycle context
	ctx    context.Context
	cancel context.CancelFunc
}

// rpcPendingRequest tracks a pending RPC request awaiting response.
type rpcPendingRequest struct {
	responseChan chan *pluginsdk.RPCResponse
	timeout      <-chan time.Time
}

// NewRPCClient creates a new RPC client for the given plugin executable.
// The client is not started until Start() is called.
func NewRPCClient(executablePath string, args ...string) *RPCClient {
	return &RPCClient{
		executablePath:  executablePath,
		args:            args,
		pendingRequests: make(map[interface{}]*rpcPendingRequest),
		done:            make(chan struct{}),
	}
}

// Start starts the plugin subprocess and begins reading responses.
// It returns an error if the subprocess fails to start.
func (c *RPCClient) Start(ctx context.Context) error {
	// Create cancelable context for client lifetime
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Create subprocess
	c.cmd = exec.CommandContext(c.ctx, c.executablePath, c.args...)

	// Setup pipes
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = stdout

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = stderr

	// Start subprocess
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin process: %w", err)
	}

	// Start background goroutines
	go c.readLoop()
	go c.stderrLoop()
	go c.monitorProcess()

	return nil
}

// Stop gracefully stops the plugin subprocess.
// It cancels the context, closes pipes, and waits for the process to exit.
func (c *RPCClient) Stop() error {
	// Cancel context to signal shutdown
	if c.cancel != nil {
		c.cancel()
	}

	// Close stdin to signal plugin to exit
	if c.stdin != nil {
		c.stdin.Close()
	}

	// Wait for done channel to close (signaling readLoop exited)
	select {
	case <-c.done:
		// readLoop exited, process is done
		return nil
	case <-time.After(5 * time.Second):
		// Timeout - force kill
		if c.cmd != nil && c.cmd.Process != nil {
			if err := c.cmd.Process.Kill(); err != nil {
				// Ignore "process already finished" error
				if err.Error() != "os: process already finished" {
					return fmt.Errorf("failed to kill plugin process: %w", err)
				}
			}
		}
		return nil
	}
}

// Call makes a synchronous RPC call to the plugin.
// It sends a request and waits for the response, respecting the context timeout.
func (c *RPCClient) Call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	// Check if client is alive
	if err := c.getError(); err != nil {
		return nil, fmt.Errorf("rpc client is not running: %w", err)
	}

	// Marshal params
	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Generate request ID (use string for JSON-RPC compatibility)
	requestID := fmt.Sprintf("%d", c.nextRequestID.Add(1))

	// Create request
	req := &pluginsdk.RPCRequest{
		JSONRPC: "2.0",
		ID:      requestID,
		Method:  method,
		Params:  paramsJSON,
	}

	// Create response channel
	responseChan := make(chan *pluginsdk.RPCResponse, 1)

	// Determine timeout - use context deadline if available, otherwise default
	var timeoutChan <-chan time.Time
	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		// No context deadline, use default timeout
		timeoutChan = time.After(DefaultRPCTimeout)
	}

	// Register pending request
	c.requestsMu.Lock()
	c.pendingRequests[requestID] = &rpcPendingRequest{
		responseChan: responseChan,
		timeout:      timeoutChan,
	}
	c.requestsMu.Unlock()

	// Send request
	if err := c.sendRequest(req); err != nil {
		c.requestsMu.Lock()
		delete(c.pendingRequests, requestID)
		c.requestsMu.Unlock()
		return nil, err
	}

	// Wait for response, timeout, or cancellation
	if hasDeadline {
		// Context has deadline - only wait for ctx.Done or response
		select {
		case resp := <-responseChan:
			if resp.Error != nil {
				return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
			}
			return resp.Result, nil
		case <-ctx.Done():
			c.requestsMu.Lock()
			delete(c.pendingRequests, requestID)
			c.requestsMu.Unlock()
			return nil, ctx.Err()
		case <-c.done:
			return nil, fmt.Errorf("rpc client stopped: %w", c.getError())
		}
	} else {
		// No context deadline - use default timeout
		select {
		case resp := <-responseChan:
			if resp.Error != nil {
				return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
			}
			return resp.Result, nil
		case <-timeoutChan:
			c.requestsMu.Lock()
			delete(c.pendingRequests, requestID)
			c.requestsMu.Unlock()
			return nil, fmt.Errorf("rpc call timed out after %v", DefaultRPCTimeout)
		case <-ctx.Done():
			c.requestsMu.Lock()
			delete(c.pendingRequests, requestID)
			c.requestsMu.Unlock()
			return nil, ctx.Err()
		case <-c.done:
			return nil, fmt.Errorf("rpc client stopped: %w", c.getError())
		}
	}
}

// SetEventChannel sets the channel to receive events from the plugin.
// This must be called before starting event streaming.
func (c *RPCClient) SetEventChannel(eventChan chan<- pluginsdk.Event) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	c.eventChan = eventChan
}

// sendRequest sends an RPC request to the plugin via stdin.
func (c *RPCClient) sendRequest(req *pluginsdk.RPCRequest) error {
	// Marshal request
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write request with newline
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		c.setError(fmt.Errorf("failed to write request: %w", err))
		return err
	}

	return nil
}

// readLoop reads responses and events from the plugin's stdout.
// It runs in a background goroutine until the client is stopped.
func (c *RPCClient) readLoop() {
	defer close(c.done)

	scanner := bufio.NewScanner(c.stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB initial, 1MB max

	for scanner.Scan() {
		line := scanner.Bytes()

		// Try to parse as event first (events have "event" field)
		var eventCheck struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal(line, &eventCheck); err == nil && eventCheck.Event == "event" {
			c.handleEvent(line)
			continue
		}

		// Otherwise, parse as RPC response
		var resp pluginsdk.RPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Invalid JSON - log and continue
			continue
		}

		c.handleResponse(&resp)
	}

	if err := scanner.Err(); err != nil {
		c.setError(fmt.Errorf("stdout read error: %w", err))
	}
}

// stderrLoop reads stderr output from the plugin for logging/debugging.
func (c *RPCClient) stderrLoop() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		// For now, we just consume stderr to prevent blocking.
		// In a full implementation, this would be forwarded to a logger.
		_ = scanner.Text()
	}
}

// monitorProcess monitors the subprocess and detects crashes.
func (c *RPCClient) monitorProcess() {
	err := c.cmd.Wait()
	if err != nil && c.ctx.Err() == nil {
		// Process crashed (not due to cancellation)
		c.setError(fmt.Errorf("plugin process crashed: %w", err))
	}
}

// handleResponse routes an RPC response to the appropriate pending request.
func (c *RPCClient) handleResponse(resp *pluginsdk.RPCResponse) {
	c.requestsMu.Lock()
	pending, ok := c.pendingRequests[resp.ID]
	if ok {
		delete(c.pendingRequests, resp.ID)
	}
	c.requestsMu.Unlock()

	if ok {
		select {
		case pending.responseChan <- resp:
		case <-pending.timeout:
			// Response arrived too late (after timeout)
		}
	}
}

// handleEvent forwards an event to the event channel.
func (c *RPCClient) handleEvent(data []byte) {
	c.eventMu.Lock()
	eventChan := c.eventChan
	c.eventMu.Unlock()

	if eventChan == nil {
		return // No event channel registered
	}

	var rpcEvent pluginsdk.RPCEvent
	if err := json.Unmarshal(data, &rpcEvent); err != nil {
		return // Invalid event
	}

	// Convert RPC event to SDK event
	timestamp, _ := time.Parse(time.RFC3339, rpcEvent.Timestamp)
	event := pluginsdk.Event{
		Type:      rpcEvent.Type,
		Source:    rpcEvent.Source,
		Timestamp: timestamp,
		Payload:   rpcEvent.Payload,
		Metadata:  rpcEvent.Metadata,
		Version:   rpcEvent.Version,
	}

	// Send to event channel (non-blocking)
	select {
	case eventChan <- event:
	default:
		// Event channel full - drop event
	}
}

// setError stores a fatal error that terminates the client.
func (c *RPCClient) setError(err error) {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	if c.err == nil {
		c.err = err
	}
}

// getError retrieves the stored fatal error, if any.
func (c *RPCClient) getError() error {
	c.errMu.RLock()
	defer c.errMu.RUnlock()
	return c.err
}
