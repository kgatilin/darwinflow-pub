#!/usr/bin/env python3

import re

# Read the file
with open('/workspace/pkg/plugins/task_manager/tui_models.go', 'r') as f:
    content = f.read()

# Replace the backlog rendering section
old_pattern = r'''		for idx, task := range m\.backlogTasks \{
			var taskLine string
			if m\.selectedItemType == SelectBacklog && idx == m\.selectedBacklogIdx \{
				taskLine = fmt\.Sprintf\("→ %s %s", task\.ID, task\.Title\)
				s \+= selectedTrackStyle\.Render\(taskLine\) \+ "\\n"
			\} else \{
				taskLine = fmt\.Sprintf\("  %s %s", task\.ID, task\.Title\)
				s \+= trackItemStyle\.Render\(taskLine\) \+ "\\n"
			\}
		\}'''

new_replacement = '''		for idx, task := range m.backlogTasks {
			// Get track information for display
			trackInfo := ""
			if task.TrackID != "" {
				track, err := m.repository.GetTrack(m.ctx, task.TrackID)
				if err == nil && track != nil {
					trackInfo = fmt.Sprintf(" [%s]", track.ID)
				}
			}

			var taskLine string
			if m.selectedItemType == SelectBacklog && idx == m.selectedBacklogIdx {
				taskLine = fmt.Sprintf("→ %s %s%s", task.ID, task.Title, trackInfo)
				s += selectedTrackStyle.Render(taskLine) + "\\n"
			} else {
				taskLine = fmt.Sprintf("  %s %s%s", task.ID, task.Title, trackInfo)
				s += trackItemStyle.Render(taskLine) + "\\n"
			}
		}'''

# Apply replacement
content = re.sub(old_pattern, new_replacement, content, flags=re.MULTILINE)

# Write back
with open('/workspace/pkg/plugins/task_manager/tui_models.go', 'w') as f:
    f.write(content)

print("Updated TUI models successfully")
