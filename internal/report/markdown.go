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
	if strings.TrimSpace(result.MeetingSummary) != "" {
		b.WriteString("## Meeting Summary\n\n")
		b.WriteString(result.MeetingSummary + "\n\n")
	}
	b.WriteString(fmt.Sprintf("- 会议日期: `%s`\n", result.Record.MeetingDate.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("- 决策数: `%d`\n", len(result.Decisions)))
	b.WriteString(fmt.Sprintf("- 任务数: `%d`\n", len(result.Tasks)))
	b.WriteString(fmt.Sprintf("- Review 项: `%d`\n", len(result.Issues)))
	b.WriteString(fmt.Sprintf("- 冲突数: `%d`\n", len(result.Conflicts)))
	b.WriteString(fmt.Sprintf("- 同步记录: `%d`\n\n", len(result.Synced)))

	b.WriteString("## Decisions\n\n")
	if len(result.Decisions) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, d := range result.Decisions {
			b.WriteString(fmt.Sprintf("- `%s` %s", d.ID, d.Text))
			if strings.TrimSpace(d.SourceText) != "" {
				b.WriteString(fmt.Sprintf(" | source: `%s`", d.SourceText))
			}
			b.WriteString("\n")
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
			line := fmt.Sprintf("- `%s` %s | owner: `%s` | due: `%s` | deps: `%s`",
				t.ID, t.Title, t.Owner, due, deps)
			if strings.TrimSpace(t.AcceptanceCriteria) != "" {
				line += fmt.Sprintf(" | acceptance: `%s`", t.AcceptanceCriteria)
			}
			if len(t.RiskFlags) > 0 {
				line += fmt.Sprintf(" | risks: `%s`", strings.Join(t.RiskFlags, ", "))
			}
			b.WriteString(line + "\n")
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

	b.WriteString("## Conflicts\n\n")
	if len(result.Conflicts) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, conflict := range result.Conflicts {
			b.WriteString(fmt.Sprintf("- [%s] %s\n", conflict.Type, conflict.Message))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Follow Up Questions\n\n")
	if len(result.FollowUpQuestions) == 0 {
		b.WriteString("- (none)\n\n")
	} else {
		for _, question := range result.FollowUpQuestions {
			b.WriteString("- " + question + "\n")
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
