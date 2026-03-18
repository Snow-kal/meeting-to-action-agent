package agents

import (
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

func TestTaskPlannerPlan(t *testing.T) {
	base := time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local)
	record := domain.MeetingRecord{
		MeetingDate: base,
		Lines: []string{
			"行动项：@张三 在 2026-03-20 前完成 API 接入，依赖 TASK-007",
		},
	}

	planner := NewTaskPlannerAgent()
	tasks := planner.Plan(record, nil)
	if len(tasks) != 1 {
		t.Fatalf("want 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Owner != "张三" {
		t.Fatalf("want owner 张三, got %s", task.Owner)
	}
	if task.DueDate == nil || task.DueDate.Format("2006-01-02") != "2026-03-20" {
		t.Fatalf("want due date 2026-03-20, got %+v", task.DueDate)
	}
	if len(task.Dependencies) != 1 || task.Dependencies[0] != "TASK-007" {
		t.Fatalf("unexpected dependencies: %+v", task.Dependencies)
	}
}

func TestTaskPlannerSingleLetterOwner(t *testing.T) {
	base := time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local)
	record := domain.MeetingRecord{
		MeetingDate: base,
		Lines: []string{
			"行动项：A 负责回归测试，明天完成",
		},
	}

	planner := NewTaskPlannerAgent()
	tasks := planner.Plan(record, nil)
	if len(tasks) != 1 {
		t.Fatalf("want 1 task, got %d", len(tasks))
	}
	if tasks[0].Owner != "A" {
		t.Fatalf("want owner A, got %s", tasks[0].Owner)
	}
}

func TestTaskTitleTruncationKeepsUTF8Valid(t *testing.T) {
	longChinese := "行动项：这是一个很长很长的中文任务标题用于验证截断不会把UTF8字符切坏导致乱码和替换符号出现"
	title := cleanupTaskTitle(longChinese)
	if !utf8.ValidString(title) {
		t.Fatalf("title should be valid utf8, got %q", title)
	}

	planner := NewTaskPlannerAgent()
	task := planner.buildTaskFromDecision("TASK-001", domain.Decision{
		ID:   "DEC-001",
		Text: "这是一个非常非常长的决策内容用于验证按字符截断不会导致乱码和替换符号出现",
	}, time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local))
	if !utf8.ValidString(task.Title) {
		t.Fatalf("decision-based title should be valid utf8, got %q", task.Title)
	}
}
