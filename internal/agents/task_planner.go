package agents

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/timeutil"
)

var (
	reAtOwner     = regexp.MustCompile(`@([\p{Han}A-Za-z0-9_-]{1,20})`)
	reOwnerLabel  = regexp.MustCompile(`(?:负责人|owner)[:：=]\s*([\p{Han}A-Za-z0-9_-]{1,20})`)
	reOwnerAction = regexp.MustCompile(`([\p{Han}A-Za-z0-9_-]{1,20})\s*负责`)
	reDepends     = regexp.MustCompile(`(?:依赖|depends on)[:：\s]*([A-Z]+-\d+)`)
)

type TaskPlannerAgent struct {
	taskKeywords       []string
	taskActionKeywords []string
}

func NewTaskPlannerAgent() *TaskPlannerAgent {
	return &TaskPlannerAgent{
		taskKeywords:       []string{"行动项", "任务", "todo", "action", "跟进"},
		taskActionKeywords: []string{"负责", "提交", "修复", "准备", "整理", "评审", "落实", "编写", "联调", "回归", "上线", "发布", "冻结", "跟进"},
	}
}

func (a *TaskPlannerAgent) Plan(record domain.MeetingRecord, decisions []domain.Decision) []domain.Task {
	tasks := make([]domain.Task, 0)
	taskIndex := 1

	for _, line := range record.Lines {
		if !a.isTaskLine(line) {
			continue
		}
		task := a.buildTaskFromLine(fmt.Sprintf("TASK-%03d", taskIndex), line, record.MeetingDate)
		tasks = append(tasks, task)
		taskIndex++
	}

	if len(tasks) == 0 {
		for _, decision := range decisions {
			task := a.buildTaskFromDecision(fmt.Sprintf("TASK-%03d", taskIndex), decision, record.MeetingDate)
			tasks = append(tasks, task)
			taskIndex++
		}
	}

	return tasks
}

func (a *TaskPlannerAgent) isTaskLine(line string) bool {
	lower := strings.ToLower(line)
	if strings.HasPrefix(line, "备注") || strings.HasPrefix(line, "说明") || strings.HasPrefix(line, "讨论") || strings.HasPrefix(line, "风险") || strings.HasPrefix(line, "问题") {
		return false
	}
	if strings.HasPrefix(line, "决策") || strings.HasPrefix(line, "决定") || strings.HasPrefix(line, "结论") {
		return false
	}
	for _, kw := range a.taskKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	if extractOwner(line) != "" {
		return true
	}
	for _, kw := range a.taskActionKeywords {
		if strings.Contains(line, kw) {
			return true
		}
	}
	return false
}

func (a *TaskPlannerAgent) buildTaskFromLine(id, line string, baseDate time.Time) domain.Task {
	owner := extractOwner(line)
	dueDate, _ := timeutil.ExtractDueDate(line, baseDate)
	deps := extractDependencies(line)
	title := cleanupTaskTitle(line)

	return domain.Task{
		ID:                 id,
		Title:              title,
		Description:        line,
		Owner:              owner,
		DueDate:            dueDate,
		Dependencies:       deps,
		SourceText:         line,
		AcceptanceCriteria: buildAcceptanceCriteria(title, line),
		Confidence:         taskConfidence(owner, dueDate != nil),
	}
}

func (a *TaskPlannerAgent) buildTaskFromDecision(id string, decision domain.Decision, baseDate time.Time) domain.Task {
	owner := extractOwner(decision.Text)
	dueDate, _ := timeutil.ExtractDueDate(decision.Text, baseDate)
	title := "落实决策: " + decision.Text
	title = truncateByRunes(title, 64)

	return domain.Task{
		ID:                 id,
		Title:              title,
		Description:        decision.Text,
		Owner:              owner,
		DueDate:            dueDate,
		SourceDecisionID:   decision.ID,
		Dependencies:       extractDependencies(decision.Text),
		SourceText:         firstNonEmpty(decision.SourceText, decision.Text),
		AcceptanceCriteria: buildAcceptanceCriteria(title, decision.Text),
		Confidence:         maxFloat(decision.Confidence, taskConfidence(owner, dueDate != nil)),
	}
}

func extractOwner(text string) string {
	if m := reAtOwner.FindStringSubmatch(text); len(m) == 2 {
		return m[1]
	}
	if m := reOwnerLabel.FindStringSubmatch(text); len(m) == 2 {
		return m[1]
	}
	if m := reOwnerAction.FindStringSubmatch(text); len(m) == 2 {
		return m[1]
	}
	reOwnerLead := regexp.MustCompile(`由\s*([\p{Han}A-Za-z0-9_-]{1,20})\s*(?:牵头|跟进|推进)`)
	if m := reOwnerLead.FindStringSubmatch(text); len(m) == 2 {
		return m[1]
	}
	return ""
}

func extractDependencies(text string) []string {
	matches := reDepends.FindAllStringSubmatch(text, -1)
	deps := make([]string, 0, len(matches))
	seen := make(map[string]struct{})
	for _, m := range matches {
		if len(m) != 2 {
			continue
		}
		dep := strings.TrimSpace(m[1])
		if dep == "" {
			continue
		}
		if _, ok := seen[dep]; ok {
			continue
		}
		seen[dep] = struct{}{}
		deps = append(deps, dep)
	}
	return deps
}

func cleanupTaskTitle(line string) string {
	cleaned := strings.TrimSpace(line)
	prefixes := []string{"行动项：", "行动项:", "任务：", "任务:", "TODO:", "todo:", "Action:", "action:"}
	for _, p := range prefixes {
		cleaned = strings.TrimPrefix(cleaned, p)
	}
	cleaned = strings.TrimSpace(cleaned)
	return truncateByRunes(cleaned, 64)
}

func truncateByRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes])
}

func buildAcceptanceCriteria(title, description string) string {
	text := title + " " + description
	switch {
	case strings.Contains(text, "提交"):
		return "相关成果已提交并可查阅。"
	case strings.Contains(text, "修复"):
		return "问题已修复并通过验证。"
	case strings.Contains(text, "评审"):
		return "评审已完成并形成明确结论。"
	case strings.Contains(text, "准备"), strings.Contains(text, "整理"), strings.Contains(text, "编写"):
		return "相关材料已产出并可供审核。"
	case strings.Contains(text, "回归"):
		return "回归测试已执行完成且结果可复核。"
	case strings.Contains(text, "发布"), strings.Contains(text, "上线"), strings.Contains(text, "冻结"):
		return "发布动作完成且状态可验证。"
	default:
		return "任务结果已交付并可验证。"
	}
}

func taskConfidence(owner string, hasDue bool) float64 {
	switch {
	case owner != "" && hasDue:
		return 0.91
	case owner != "" || hasDue:
		return 0.82
	default:
		return 0.68
	}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
