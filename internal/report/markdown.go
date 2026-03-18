package report

import (
	"fmt"
	"strings"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

func BuildMarkdown(result *domain.PipelineResult) string {
	if result == nil {
		return "# Meeting To Action Report\n\n结果为空。\n"
	}

	var b strings.Builder
	b.WriteString("# Meeting To Action Report\n\n")
	b.WriteString(fmt.Sprintf("- 会议日期: `%s`\n", result.Record.MeetingDate.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("- 决策数: `%d`\n", len(result.Decisions)))
	b.WriteString(fmt.Sprintf("- 任务数: `%d`\n", len(result.Tasks)))
	b.WriteString(fmt.Sprintf("- Review 项: `%d`\n", len(result.Issues)))
	b.WriteString(fmt.Sprintf("- 同步记录: `%d`\n\n", len(result.Synced)))

	b.WriteString("## Decisions\n\n")
	if len(result.Decisions) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, d := range result.Decisions {
			b.WriteString(fmt.Sprintf("- `%s` %s\n", d.ID, d.Text))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Tasks\n\n")
	if len(result.Tasks) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, t := range result.Tasks {
			due := "-"
			if t.DueDate != nil {
				due = t.DueDate.Format("2006-01-02")
			}
			deps := "-"
			if len(t.Dependencies) > 0 {
				deps = strings.Join(t.Dependencies, ", ")
			}
			b.WriteString(fmt.Sprintf("- `%s` %s | owner: `%s` | due: `%s` | deps: `%s`\n",
				t.ID, t.Title, t.Owner, due, deps))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Review Issues\n\n")
	if len(result.Issues) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, i := range result.Issues {
			b.WriteString(fmt.Sprintf("- `%s` [%s] %s\n", i.TaskID, i.Type, i.Message))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Sync Results\n\n")
	if len(result.Synced) == 0 {
		b.WriteString("- (none)\n")
	} else {
		for _, s := range result.Synced {
			line := fmt.Sprintf("- `%s` -> `%s` status=`%s`", s.TaskID, s.Target, s.Status)
			if s.RemoteID != "" {
				line += fmt.Sprintf(" remote=`%s`", s.RemoteID)
			}
			if s.Error != "" {
				line += fmt.Sprintf(" error=`%s`", s.Error)
			}
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}
