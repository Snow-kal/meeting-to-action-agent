package agents

import (
	"testing"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

func TestOwnerAgentResolve(t *testing.T) {
	agent := NewOwnerAgent()
	tasks := []domain.Task{
		{
			ID:         "TASK-001",
			Title:      "补充回归用例",
			SourceText: "行动项：A 负责补充回归用例",
		},
	}

	resolved := agent.Resolve(tasks, nil)
	if len(resolved) != 1 || resolved[0].Owner != "A" {
		t.Fatalf("expected owner A, got %+v", resolved)
	}
}

func TestDeadlineAgentResolve(t *testing.T) {
	agent := NewDeadlineAgent()
	base := time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local)
	tasks := []domain.Task{
		{
			ID:         "TASK-001",
			Title:      "准备发布材料",
			SourceText: "行动项：A 负责准备发布材料，明天中午前提交",
		},
		{
			ID:          "TASK-002",
			Title:       "上线准备",
			Description: "完成上线准备",
		},
	}

	resolved := agent.Resolve(tasks, nil, base)
	if resolved[0].DueDate == nil || resolved[0].DueDate.Format("2006-01-02") != "2026-03-19" {
		t.Fatalf("expected parsed due date 2026-03-19, got %+v", resolved[0].DueDate)
	}
	if resolved[1].DueDate == nil || !resolved[1].DueDateInferred {
		t.Fatalf("expected inferred due date, got %+v", resolved[1].DueDate)
	}
}
