package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/agents"
	"github.com/Snow-kal/meeting-to-action-agent/internal/syncer"
)

func TestOrchestratorRun(t *testing.T) {
	orch := NewOrchestrator(
		agents.NewRecorderAgent(),
		agents.NewDecisionAgent(),
		agents.NewTaskPlannerAgent(),
		agents.NewReviewerAgent(),
		syncer.NewJiraClientFromEnv(true),
		syncer.NewNotionClientFromEnv(true),
	)

	base := time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local)
	input := `
	会议结论：决定本周完成移动端灰度。
	行动项：@李雷 明天完成登录流程联调，依赖 TASK-100
	`

	result, err := orch.Run(context.Background(), input, Options{
		MeetingDate: base,
		SyncTarget:  SyncBoth,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(result.Decisions) == 0 {
		t.Fatalf("expected decisions")
	}
	if len(result.Tasks) == 0 {
		t.Fatalf("expected tasks")
	}
	if len(result.Synced) != len(result.Tasks)*2 {
		t.Fatalf("expected sync results for both targets")
	}
}
