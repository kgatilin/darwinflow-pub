package cli

// GetStatusIcon returns the icon for a given status string
// Used by CLI output formatting (roadmap full view, etc.)
func GetStatusIcon(status string) string {
	switch status {
	case "done", "complete":
		return "✓"
	case "review":
		return "👁"
	case "in-progress":
		return "→"
	case "blocked":
		return "✗"
	case "waiting":
		return "⏸"
	default:
		return "○"
	}
}
