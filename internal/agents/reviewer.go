package agents

import (
	"fmt"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/timeutil"
)

type ReviewerAgent struct {
	defaultOwner string
	defaultDueIn int
}

func NewReviewerAgent() *ReviewerAgent {
	return &ReviewerAgent{
		defaultOwner: "待指派",
		defaultDueIn: 7,
	}
}

func (a *ReviewerAgent) Review(tasks []domain.Task, meetingDate time.Time) ([]domain.Task, []domain.ReviewIssue) {
	reviewed := make([]domain.Task, 0, len(tasks))
	issues := make([]domain.ReviewIssue, 0)
	base := timeutil.NormalizeDate(meetingDate)

	for _, t := range tasks {
		task := t
		if task.Owner == "" {
			task.Owner = a.defaultOwner
			issues = append(issues, domain.ReviewIssue{
				TaskID:  task.ID,
				Type:    "missing_owner",
				Message: "未识别到负责人，已自动填充为待指派",
			})
		}
		if task.DueDate == nil {
			defaultDue := base.AddDate(0, 0, a.defaultDueIn)
			task.DueDate = &defaultDue
			issues = append(issues, domain.ReviewIssue{
				TaskID:  task.ID,
				Type:    "missing_due_date",
				Message: fmt.Sprintf("未识别到截止时间，已自动补齐为会议后 %d 天", a.defaultDueIn),
			})
		}
		if len(task.Dependencies) == 0 {
			issues = append(issues, domain.ReviewIssue{
				TaskID:  task.ID,
				Type:    "missing_dependency",
				Message: "未声明依赖关系（可按实际需要补充）",
			})
		}
		reviewed = append(reviewed, task)
	}

	return reviewed, issues
}
