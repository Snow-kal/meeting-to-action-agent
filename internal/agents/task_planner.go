package agents

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/timeutil"
)

var (
	reAtOwner     = regexp.MustCompile(`@([\p{Han}A-Za-z0-9_-]{2,20})`)
	reOwnerLabel  = regexp.MustCompile(`(?:负责人|owner)[:：=]\s*([\p{Han}A-Za-z0-9_-]{2,20})`)
	reOwnerAction = regexp.MustCompile(`([\p{Han}A-Za-z0-9_-]{2,20})\s*负责`)
	reDepends     = regexp.MustCompile(`(?:依赖|depends on)[:：\s]*([A-Z]+-\d+)`)
)

type TaskPlannerAgent struct {
	taskKeywords []string
}

func NewTaskPlannerAgent() *TaskPlannerAgent {
	return &TaskPlannerAgent{
		taskKeywords: []string{"行动项", "任务", "todo", "action", "跟进", "完成"},
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
	for _, kw := range a.taskKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
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
		ID:           id,
		Title:        title,
		Description:  line,
		Owner:        owner,
		DueDate:      dueDate,
		Dependencies: deps,
	}
}

func (a *TaskPlannerAgent) buildTaskFromDecision(id string, decision domain.Decision, baseDate time.Time) domain.Task {
	owner := extractOwner(decision.Text)
	dueDate, _ := timeutil.ExtractDueDate(decision.Text, baseDate)
	title := "落实决策: " + decision.Text
	if len(title) > 64 {
		title = title[:64]
	}

	return domain.Task{
		ID:               id,
		Title:            title,
		Description:      decision.Text,
		Owner:            owner,
		DueDate:          dueDate,
		SourceDecisionID: decision.ID,
		Dependencies:     extractDependencies(decision.Text),
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
	if len(cleaned) > 64 {
		return cleaned[:64]
	}
	return cleaned
}
