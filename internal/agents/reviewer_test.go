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

	reviewed, issues, conflicts, followUp := reviewer.Review(tasks, base)
	if len(reviewed) != 1 {
		t.Fatalf("want 1 reviewed task, got %d", len(reviewed))
	}
	task := reviewed[0]
	if task.Owner != "暂无" {
		t.Fatalf("want default owner 暂无, got %s", task.Owner)
	}
	if task.DueDate == nil || task.DueDate.Format("2006-01-02") != "2026-03-25" {
		t.Fatalf("want due 2026-03-25, got %+v", task.DueDate)
	}
	if len(issues) == 0 {
		t.Fatalf("want review issues")
	}
	if len(conflicts) != 0 {
		t.Fatalf("did not expect conflicts, got %+v", conflicts)
	}
	if len(followUp) == 0 {
		t.Fatalf("want follow up questions")
	}
}

func TestReviewerDetectOwnerOverloadConflict(t *testing.T) {
	base := time.Date(2026, 3, 18, 9, 0, 0, 0, time.Local)
	reviewer := NewReviewerAgent()
	due := time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local)
	tasks := []domain.Task{
		{ID: "TASK-001", Title: "整理发布说明", Owner: "A", DueDate: &due, AcceptanceCriteria: "已提交"},
		{ID: "TASK-002", Title: "整理回滚方案", Owner: "A", DueDate: &due, AcceptanceCriteria: "已提交"},
		{ID: "TASK-003", Title: "整理测试记录", Owner: "A", DueDate: &due, AcceptanceCriteria: "已提交"},
	}

	_, _, conflicts, followUp := reviewer.Review(tasks, base)
	if len(conflicts) == 0 {
		t.Fatalf("expected owner overload conflict")
	}
	if len(followUp) == 0 {
		t.Fatalf("expected follow up from conflict")
	}
}
