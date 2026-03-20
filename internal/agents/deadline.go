package agents

import (
	"strings"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/timeutil"
)

type DeadlineAgent struct{}

func NewDeadlineAgent() *DeadlineAgent {
	return &DeadlineAgent{}
}

func (a *DeadlineAgent) Resolve(tasks []domain.Task, decisions []domain.Decision, meetingDate time.Time) []domain.Task {
	decisionDueHints := make(map[string]string, len(decisions))
	for _, decision := range decisions {
		if decision.DueHint != "" {
			decisionDueHints[decision.ID] = decision.DueHint
		}
	}

	resolved := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		current := task
		if current.DueDate == nil {
			if due, ok := timeutil.ExtractDueDate(current.SourceText, meetingDate); ok {
				current.DueDate = due
			} else if due, ok := timeutil.ExtractDueDate(current.Description, meetingDate); ok {
				current.DueDate = due
			} else if hint, ok := decisionDueHints[current.SourceDecisionID]; ok {
				if due, parsed := timeutil.ExtractDueDate(hint, meetingDate); parsed {
					current.DueDate = due
				}
			}
		}
		if current.DueDate == nil {
			inferred := inferDueDate(current, meetingDate)
			current.DueDate = &inferred
			current.DueDateInferred = true
		}
		resolved = append(resolved, current)
	}

	return resolved
}

func inferDueDate(task domain.Task, meetingDate time.Time) time.Time {
	base := timeutil.NormalizeDate(meetingDate)
	text := strings.ToLower(task.Title + " " + task.Description)
	switch {
	case strings.Contains(text, "整理"), strings.Contains(text, "提交"), strings.Contains(text, "确认"):
		return base.AddDate(0, 0, 3)
	case strings.Contains(text, "评审"), strings.Contains(text, "回归"), strings.Contains(text, "测试"):
		return base.AddDate(0, 0, 5)
	case strings.Contains(text, "上线"), strings.Contains(text, "发布"), strings.Contains(text, "冻结"), strings.Contains(text, "开发"), strings.Contains(text, "修复"):
		return base.AddDate(0, 0, 7)
	default:
		return base.AddDate(0, 0, 5)
	}
}
