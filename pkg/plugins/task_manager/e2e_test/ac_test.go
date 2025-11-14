package task_manager_e2e_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ACTestSuite tests acceptance criteria commands end-to-end
type ACTestSuite struct {
	E2ETestSuite
}

func TestACSuite(t *testing.T) {
	suite.Run(t, new(ACTestSuite))
}

// TestACAdd tests adding acceptance criteria to a task
func (s *ACTestSuite) TestACAdd() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add acceptance criterion
	acOutput, err := s.run("ac", "add", taskID, "--description", "User can create a task", "--testing-instructions", "1. Run command\n2. Verify output")
	s.requireSuccess(acOutput, err, "failed to add acceptance criterion")
	acID := s.parseID(acOutput, "ac")
	s.NotEmpty(acID, "AC ID should be extracted")
}

// TestACAddWithoutTask tests that AC add fails without valid task
func (s *ACTestSuite) TestACAddWithoutTask() {
	acOutput, err := s.run("ac", "add", "TM-task-invalid", "--description", "Test AC", "--testing-instructions", "Test")
	s.requireError(err, "AC add with invalid task should fail")
	s.NotEmpty(acOutput, "error message should be provided")
}

// TestACList tests listing acceptance criteria for a task
func (s *ACTestSuite) TestACList() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add multiple acceptance criteria
	for i := 1; i <= 2; i++ {
		acOutput, err := s.run("ac", "add", taskID, "--description", "AC "+string(rune('0'+byte(i))), "--testing-instructions", "Test steps")
		s.requireSuccess(acOutput, err, "failed to add acceptance criterion")
	}

	// List ACs for task
	listOutput, err := s.run("ac", "list", taskID)
	s.requireSuccess(listOutput, err, "failed to list acceptance criteria")
	s.Contains(listOutput, "AC 1", "first AC should appear in list")
	s.Contains(listOutput, "AC 2", "second AC should appear in list")
}

// TestACShow tests showing AC details
func (s *ACTestSuite) TestACShow() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Test description", "--testing-instructions", "1. Test step 1\n2. Test step 2")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Show AC details
	showOutput, err := s.run("ac", "show", acID)
	s.requireSuccess(showOutput, err, "failed to show AC")
	s.Contains(showOutput, acID, "AC ID should appear in show output")
	s.Contains(showOutput, "Test description", "AC description should appear in show output")
}

// TestACUpdate tests updating AC properties
func (s *ACTestSuite) TestACUpdate() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Original description", "--testing-instructions", "Original steps")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Update AC
	updateOutput, err := s.run("ac", "update", acID, "--description", "Updated description")
	s.requireSuccess(updateOutput, err, "failed to update AC")

	// Verify update
	showOutput, err := s.run("ac", "show", acID)
	s.requireSuccess(showOutput, err, "failed to show updated AC")
	s.Contains(showOutput, "Updated description", "updated description should appear in show output")
}

// TestACVerify tests verifying an acceptance criterion
func (s *ACTestSuite) TestACVerify() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Test AC", "--testing-instructions", "Test steps")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Verify AC
	verifyOutput, err := s.run("ac", "verify", acID)
	s.requireSuccess(verifyOutput, err, "failed to verify AC")

	// Check verified status
	showOutput, err := s.run("ac", "show", acID)
	s.requireSuccess(showOutput, err, "failed to show AC after verification")
	s.Contains(showOutput, "verified", "AC should show verified status")
}

// TestACFail tests marking an AC as failed with feedback
func (s *ACTestSuite) TestACFail() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Test AC", "--testing-instructions", "Test steps")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Fail AC with feedback
	failOutput, err := s.run("ac", "fail", acID, "--feedback", "This AC did not meet requirements")
	s.requireSuccess(failOutput, err, "failed to mark AC as failed")

	// Check failed status
	showOutput, err := s.run("ac", "show", acID)
	s.requireSuccess(showOutput, err, "failed to show AC after failure")
	s.Contains(showOutput, "failed", "AC should show failed status")
}

// TestACListIteration tests listing ACs by iteration
func (s *ACTestSuite) TestACListIteration() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Create iteration
	iterOutput, err := s.run("iteration", "create", "--name", "Test Iteration", "--goal", "Complete tasks", "--deliverable", "Working features")
	s.requireSuccess(iterOutput, err, "failed to create iteration")

	// Add task to iteration
	addTaskOutput, err := s.run("iteration", "add-task", "1", taskID)
	s.requireSuccess(addTaskOutput, err, "failed to add task to iteration")

	// Add AC to task
	acOutput, err := s.run("ac", "add", taskID, "--description", "Iteration AC", "--testing-instructions", "Test steps")
	s.requireSuccess(acOutput, err, "failed to add AC")

	// List ACs for iteration
	listOutput, err := s.run("ac", "list-iteration", "1")
	s.requireSuccess(listOutput, err, "failed to list ACs for iteration")
	s.Contains(listOutput, "Iteration AC", "iteration AC should appear in list")
}

// TestACFailed tests listing failed acceptance criteria
func (s *ACTestSuite) TestACFailed() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Test AC", "--testing-instructions", "Test steps")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Fail the AC
	failOutput, err := s.run("ac", "fail", acID, "--feedback", "AC failed testing")
	s.requireSuccess(failOutput, err, "failed to mark AC as failed")

	// List failed ACs
	failedOutput, err := s.run("ac", "failed")
	s.requireSuccess(failedOutput, err, "failed to list failed ACs")
	// Should contain information about failed ACs
	s.NotEmpty(failedOutput, "failed ACs list should not be empty")
}

// TestACFailedByTask tests listing failed ACs filtered by task
func (s *ACTestSuite) TestACFailedByTask() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create two tasks
	task1Output, err := s.run("task", "create", "--track", trackID, "--title", "Task 1", "--rank", "100")
	s.requireSuccess(task1Output, err, "failed to create task 1")
	task1ID := s.parseID(task1Output, "task")

	task2Output, err := s.run("task", "create", "--track", trackID, "--title", "Task 2", "--rank", "200")
	s.requireSuccess(task2Output, err, "failed to create task 2")
	task2ID := s.parseID(task2Output, "task")

	// Add failing AC to task 1
	ac1Output, err := s.run("ac", "add", task1ID, "--description", "AC for Task 1", "--testing-instructions", "Test")
	s.requireSuccess(ac1Output, err, "failed to add AC to task 1")
	ac1ID := s.parseID(ac1Output, "ac")

	// Fail AC 1
	_, err = s.run("ac", "fail", ac1ID, "--feedback", "Failed")
	s.requireSuccess("", err, "failed to mark AC as failed")

	// Add passing AC to task 2
	ac2Output, err := s.run("ac", "add", task2ID, "--description", "AC for Task 2", "--testing-instructions", "Test")
	s.requireSuccess(ac2Output, err, "failed to add AC to task 2")

	// List failed ACs for task 1
	failedOutput, err := s.run("ac", "failed", "--task", task1ID)
	s.requireSuccess(failedOutput, err, "failed to list failed ACs for task")
	s.NotEmpty(failedOutput, "failed ACs list should not be empty")
}

// TestACDelete tests deleting an acceptance criterion
func (s *ACTestSuite) TestACDelete() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Test AC", "--testing-instructions", "Test steps")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Delete AC
	deleteOutput, err := s.run("ac", "delete", acID, "--force")
	s.requireSuccess(deleteOutput, err, "failed to delete AC")

	// Verify AC is deleted
	showOutput, err := s.run("ac", "show", acID)
	s.requireError(err, "AC should not be found after deletion")
	s.NotEmpty(showOutput, "error message should be provided")
}

// TestACMultiple tests creating and managing multiple ACs
func (s *ACTestSuite) TestACMultiple() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add multiple ACs
	acIDs := []string{}
	for i := 1; i <= 3; i++ {
		acOutput, err := s.run("ac", "add", taskID, "--description", "AC "+string(rune('0'+byte(i))), "--testing-instructions", "Test steps")
		s.requireSuccess(acOutput, err, "failed to add AC")
		acID := s.parseID(acOutput, "ac")
		s.NotEmpty(acID, "AC ID should be extracted")
		acIDs = append(acIDs, acID)
	}

	// List all ACs for task
	listOutput, err := s.run("ac", "list", taskID)
	s.requireSuccess(listOutput, err, "failed to list ACs")

	// Verify all ACs appear in list
	for i, acID := range acIDs {
		s.Contains(listOutput, "AC "+string(rune('0'+byte(i+1))), "AC "+acID+" should appear in list")
	}
}

// TestACVerifyAndFail tests transitioning AC between verify and fail states
func (s *ACTestSuite) TestACVerifyAndFail() {
	// Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "Test AC", "--testing-instructions", "Test steps")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Verify AC
	verifyOutput, err := s.run("ac", "verify", acID)
	s.requireSuccess(verifyOutput, err, "failed to verify AC")

	// Check verified status
	showOutput, err := s.run("ac", "show", acID)
	s.requireSuccess(showOutput, err, "failed to show verified AC")
	s.Contains(showOutput, "verified", "AC should show verified status")

	// Fail the AC (transitioning back to failed)
	failOutput, err := s.run("ac", "fail", acID, "--feedback", "Needs rework")
	s.requireSuccess(failOutput, err, "failed to mark AC as failed")

	// Check failed status
	showOutput, err = s.run("ac", "show", acID)
	s.requireSuccess(showOutput, err, "failed to show failed AC")
	s.Contains(showOutput, "failed", "AC should show failed status after transition")
}

// TestACVerificationEnforcement tests that tasks cannot be marked done without verified/skipped ACs (Iteration 36, Phase 3)
func (s *ACTestSuite) TestACVerificationEnforcement() {
	// Step 1: Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Step 2: Create task in "todo" status
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Step 3: Add acceptance criteria (creates in "not_started" status)
	ac1Output, err := s.run("ac", "add", taskID, "--description", "AC 1 - Feature works", "--testing-instructions", "Test feature")
	s.requireSuccess(ac1Output, err, "failed to add AC 1")
	ac1ID := s.parseID(ac1Output, "ac")

	ac2Output, err := s.run("ac", "add", taskID, "--description", "AC 2 - Tests pass", "--testing-instructions", "Run tests")
	s.requireSuccess(ac2Output, err, "failed to add AC 2")
	ac2ID := s.parseID(ac2Output, "ac")

	// Step 4: Try to mark task as "done" with pending ACs - should FAIL
	taskUpdateOutput, err := s.run("task", "update", taskID, "--status", "done")
	s.requireError(err, "task update to 'done' should fail with unverified ACs")
	s.Contains(taskUpdateOutput, "unverified acceptance criteria", "error message should mention unverified ACs")
	s.Contains(taskUpdateOutput, ac1ID, "error should list AC1 ID")
	s.Contains(taskUpdateOutput, ac2ID, "error should list AC2 ID")

	// Step 5: Verify one AC
	verifyOutput, err := s.run("ac", "verify", ac1ID)
	s.requireSuccess(verifyOutput, err, "failed to verify AC 1")

	// Step 6: Try to mark task as "done" with one verified AC - should still FAIL (AC2 pending)
	taskUpdateOutput2, err := s.run("task", "update", taskID, "--status", "done")
	s.requireError(err, "task update to 'done' should still fail with one pending AC")
	s.Contains(taskUpdateOutput2, "unverified acceptance criteria", "error message should mention unverified ACs")
	s.Contains(taskUpdateOutput2, ac2ID, "error should list AC2 ID (still pending)")

	// Step 7: Skip the second AC
	skipOutput, err := s.run("ac", "skip", ac2ID, "--reason", "Not applicable for this task")
	s.requireSuccess(skipOutput, err, "failed to skip AC 2")

	// Step 8: Now mark task as "done" - should SUCCEED (all ACs verified or skipped)
	taskUpdateOutput3, err := s.run("task", "update", taskID, "--status", "done")
	s.requireSuccess(taskUpdateOutput3, err, "task update to 'done' should succeed with all ACs verified/skipped")

	// Step 9: Verify task is now in "done" status
	taskShowOutput, err := s.run("task", "show", taskID)
	s.requireSuccess(taskShowOutput, err, "failed to show task")
	s.Contains(taskShowOutput, "done", "task should be in 'done' status")
}

// TestACVerificationEnforcement_FailedACs tests that failed ACs also block task completion
func (s *ACTestSuite) TestACVerificationEnforcement_FailedACs() {
	// Step 1: Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Step 2: Create task
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Step 3: Add AC
	acOutput, err := s.run("ac", "add", taskID, "--description", "AC - Feature works", "--testing-instructions", "Test feature")
	s.requireSuccess(acOutput, err, "failed to add AC")
	acID := s.parseID(acOutput, "ac")

	// Step 4: Mark AC as failed
	failOutput, err := s.run("ac", "fail", acID, "--feedback", "Feature doesn't work as expected")
	s.requireSuccess(failOutput, err, "failed to mark AC as failed")

	// Step 5: Try to mark task as "done" with failed AC - should FAIL
	taskUpdateOutput, err := s.run("task", "update", taskID, "--status", "done")
	s.requireError(err, "task update to 'done' should fail with failed AC")
	s.Contains(taskUpdateOutput, "unverified acceptance criteria", "error message should mention unverified ACs")
	s.Contains(taskUpdateOutput, acID, "error should list failed AC ID")

	// Step 6: Verify AC
	verifyOutput, err := s.run("ac", "verify", acID)
	s.requireSuccess(verifyOutput, err, "failed to verify AC")

	// Step 7: Now mark task as "done" - should SUCCEED
	taskUpdateOutput2, err := s.run("task", "update", taskID, "--status", "done")
	s.requireSuccess(taskUpdateOutput2, err, "task update to 'done' should succeed with verified AC")
}

// TestACVerificationEnforcement_NoACs tests that tasks without ACs can be marked done
func (s *ACTestSuite) TestACVerificationEnforcement_NoACs() {
	// Step 1: Create track
	trackOutput, err := s.run("track", "create", "--title", "Test Track", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track")
	trackID := s.parseID(trackOutput, "track")

	// Step 2: Create task (no ACs added)
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Test Task", "--rank", "100")
	s.requireSuccess(taskOutput, err, "failed to create task")
	taskID := s.parseID(taskOutput, "task")

	// Step 3: Mark task as "done" - should SUCCEED (no ACs to verify)
	taskUpdateOutput, err := s.run("task", "update", taskID, "--status", "done")
	s.requireSuccess(taskUpdateOutput, err, "task update to 'done' should succeed without ACs")

	// Step 4: Verify task is in "done" status
	taskShowOutput, err := s.run("task", "show", taskID)
	s.requireSuccess(taskShowOutput, err, "failed to show task")
	s.Contains(taskShowOutput, "done", "task should be in 'done' status")
}
