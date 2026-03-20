package agents

import (
	"fmt"
	"strings"
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
		defaultOwner: "暂无",
		defaultDueIn: 7,
	}
}

func (a *ReviewerAgent) Review(tasks []domain.Task, meetingDate time.Time) ([]domain.Task, []domain.ReviewIssue, []domain.Conflict, []string) {
	reviewed := make([]domain.Task, 0, len(tasks))
	issues := make([]domain.ReviewIssue, 0)
	conflicts := make([]domain.Conflict, 0)
	followUp := make([]string, 0)
	base := timeutil.NormalizeDate(meetingDate)

	for _, t := range tasks {
		task := t
		task.RiskFlags = uniqueStrings(task.RiskFlags)
		if task.Owner == "" {
			task.Owner = a.defaultOwner
			task.OwnerInferred = true
			issues = append(issues, domain.ReviewIssue{
				TaskID:    task.ID,
				Type:      "missing_owner",
				Message:   "未识别到负责人，已自动标记为暂无",
				Severity:  "high",
				Inference: true,
			})
			task.RiskFlags = append(task.RiskFlags, "missing_owner")
			followUp = append(followUp, fmt.Sprintf("%s 的负责人尚未明确，请确认由谁负责。", task.ID))
		}
		if task.DueDate == nil {
			defaultDue := base.AddDate(0, 0, a.defaultDueIn)
			task.DueDate = &defaultDue
			task.DueDateInferred = true
			issues = append(issues, domain.ReviewIssue{
				TaskID:    task.ID,
				Type:      "missing_due_date",
				Message:   fmt.Sprintf("未识别到截止时间，已自动补齐为会议后 %d 天", a.defaultDueIn),
				Severity:  "high",
				Inference: true,
			})
			task.RiskFlags = append(task.RiskFlags, "missing_due_date")
			followUp = append(followUp, fmt.Sprintf("%s 的截止时间尚未明确，请确认计划完成日期。", task.ID))
		}
		if len(task.Dependencies) == 0 {
			issues = append(issues, domain.ReviewIssue{
				TaskID:   task.ID,
				Type:     "missing_dependency",
				Message:  "未声明依赖关系（可按实际需要补充）",
				Severity: "medium",
			})
			task.RiskFlags = append(task.RiskFlags, "missing_dependency")
		}
		if strings.TrimSpace(task.AcceptanceCriteria) == "" {
			issues = append(issues, domain.ReviewIssue{
				TaskID:   task.ID,
				Type:     "missing_acceptance_criteria",
				Message:  "未明确验收标准，建议补充可验证的完成条件",
				Severity: "medium",
			})
			task.RiskFlags = append(task.RiskFlags, "missing_acceptance_criteria")
			followUp = append(followUp, fmt.Sprintf("%s 缺少验收标准，请确认完成后如何验收。", task.ID))
		}
		if isLargeScope(task) {
			task.RiskFlags = append(task.RiskFlags, "large_scope")
			issues = append(issues, domain.ReviewIssue{
				TaskID:   task.ID,
				Type:     "large_scope",
				Message:  "任务范围可能过大，建议继续拆解为更小的可执行项",
				Severity: "medium",
			})
		}
		if task.OwnerInferred {
			task.RiskFlags = append(task.RiskFlags, "inferred_owner")
		}
		if task.DueDateInferred {
			task.RiskFlags = append(task.RiskFlags, "inferred_due_date")
		}
		task.RiskFlags = uniqueStrings(task.RiskFlags)
		reviewed = append(reviewed, task)
	}

	conflicts = append(conflicts, detectOwnerOverload(reviewed)...)
	conflicts = append(conflicts, detectDeadlineConflicts(reviewed, base)...)
	for _, conflict := range conflicts {
		switch conflict.Type {
		case "owner_overload":
			followUp = append(followUp, "部分负责人在相近时间窗口内任务过多，请确认是否需要重新分配资源。")
		case "unreasonable_deadline":
			followUp = append(followUp, "存在可能不合理的截止时间，请确认优先级与执行顺序。")
		}
	}

	return reviewed, issues, conflicts, uniqueStrings(followUp)
}

func isLargeScope(task domain.Task) bool {
	text := task.Title + " " + task.Description
	return len([]rune(text)) > 40 || strings.Contains(text, "以及") || strings.Contains(text, "并") || strings.Contains(text, "协同")
}

func detectOwnerOverload(tasks []domain.Task) []domain.Conflict {
	type bucket struct {
		taskIDs []string
		count   int
	}
	index := make(map[string]bucket)
	for _, task := range tasks {
		if task.Owner == "" || task.Owner == "暂无" || task.DueDate == nil {
			continue
		}
		key := task.Owner + "|" + task.DueDate.Format("2006-01-02")
		item := index[key]
		item.count++
		item.taskIDs = append(item.taskIDs, task.ID)
		index[key] = item
	}

	out := make([]domain.Conflict, 0)
	for key, item := range index {
		if item.count < 3 {
			continue
		}
		parts := strings.Split(key, "|")
		out = append(out, domain.Conflict{
			Type:     "owner_overload",
			TaskIDs:  item.taskIDs,
			Message:  fmt.Sprintf("负责人 %s 在 %s 附近存在 %d 个任务，可能过载。", parts[0], parts[1], item.count),
			Severity: "medium",
		})
	}
	return out
}

func detectDeadlineConflicts(tasks []domain.Task, meetingDate time.Time) []domain.Conflict {
	out := make([]domain.Conflict, 0)
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}
		if task.DueDate.Before(meetingDate) {
			out = append(out, domain.Conflict{
				Type:     "unreasonable_deadline",
				TaskIDs:  []string{task.ID},
				Message:  fmt.Sprintf("%s 的截止时间早于会议日期，请确认是否录入错误。", task.ID),
				Severity: "high",
			})
			continue
		}
		if isLargeScope(task) && task.DueDate.Sub(meetingDate) <= 24*time.Hour {
			out = append(out, domain.Conflict{
				Type:     "unreasonable_deadline",
				TaskIDs:  []string{task.ID},
				Message:  fmt.Sprintf("%s 任务范围较大但截止时间过近，存在执行风险。", task.ID),
				Severity: "medium",
			})
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
