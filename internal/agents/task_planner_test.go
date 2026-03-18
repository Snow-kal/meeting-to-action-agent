package agents

import (
	"testing"
	"time"

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
