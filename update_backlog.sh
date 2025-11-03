#!/bin/bash

# Update backlog rendering to be always visible (not just in full view)
# And add track information to backlog tasks

cd /workspace/pkg/plugins/task_manager

# Make the changes using perl for multi-line matching
perl -i -pe 'BEGIN{undef $/;} s/\t\/\/ Backlog \(only in full view\)\n\tif m\.showFullRoadmap && len\(m\.backlogTasks\) > 0 \{/\t\/\/ Backlog \(always visible when tasks exist)\n\tif len\(m.backlogTasks\) > 0 {/g' tui_models.go

# Add track info display
perl -i -0777 -pe 's/(backlogSectionLine = countLines\(s\)\n\t\ts \+= "\\n" \+ sectionHeaderStyle\.Render\(fmt\.Sprintf\("Backlog \(%d\):", len\(m\.backlogTasks\)\)\) \+ "\\n"\n\t\tfor idx, task := range m\.backlogTasks \{\n)/\1\t\t\t\/\/ Get track information for display\n\t\t\ttrackInfo := ""\n\t\t\tif task.TrackID != "" {\n\t\t\t\ttrack, err := m.repository.GetTrack\(m.ctx, task.TrackID\)\n\t\t\t\tif err == nil && track != nil {\n\t\t\t\t\ttrackInfo = fmt.Sprintf\(" [%s]", track.ID\)\n\t\t\t\t}\n\t\t\t}\n\n/g' tui_models.go

# Update task line format to include trackInfo
perl -i -pe 's/taskLine = fmt\.Sprintf\("→ %s %s", task\.ID, task\.Title\)/taskLine = fmt.Sprintf("→ %s %s%s", task.ID, task.Title, trackInfo)/g' tui_models.go
perl -i -pe 's/taskLine = fmt\.Sprintf\("  %s %s", task\.ID, task\.Title\)/taskLine = fmt.Sprintf("  %s %s%s", task.ID, task.Title, trackInfo)/g' tui_models.go

echo "Changes applied successfully"
