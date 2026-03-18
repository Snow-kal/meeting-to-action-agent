package agents

import (
	"testing"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

func TestReviewerReview(t *testing.T) {
	base := time.Date(2026, 3, 18, 9, 0, 0, 0, time.Local)
	reviewer := NewReviewerAgent()
	tasks := []domain.Task{
		{ID: "TASK-001", Title: "补充自动化测试"},
	}

	reviewed, issues := reviewer.Review(tasks, base)
	if len(reviewed) != 1 {
		t.Fatalf("want 1 reviewed task, got %d", len(reviewed))
	}
	task := reviewed[0]
	if task.Owner != "待指派" {
		t.Fatalf("want default owner 待指派, got %s", task.Owner)
	}
	if task.DueDate == nil || task.DueDate.Format("2006-01-02") != "2026-03-25" {
		t.Fatalf("want due 2026-03-25, got %+v", task.DueDate)
	}
	if len(issues) == 0 {
		t.Fatalf("want review issues")
	}
}
