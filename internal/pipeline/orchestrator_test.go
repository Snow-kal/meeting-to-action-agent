package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/agents"
	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/syncer"
)

func TestOrchestratorRun(t *testing.T) {
	orch := NewOrchestrator(
		agents.NewRecorderAgent(),
		agents.NewDecisionAgent(),
		agents.NewTaskPlannerAgent(),
		agents.NewReviewerAgent(),
		nil,
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

type slowSyncer struct {
	delay time.Duration
}

type mockExtractor struct{}

func (m mockExtractor) Extract(ctx context.Context, rawText string, meetingDate time.Time) ([]domain.Decision, []domain.Task, error) {
	return []domain.Decision{
			{ID: "LLM-1", Text: "确定推进支付重构"},
		},
		[]domain.Task{
			{
				ID:          "LLM-TASK-1",
				Title:       "支付重构设计评审",
				Description: "组织评审会议并输出结论",
				Owner:       "王五",
			},
		}, nil
}

func (s slowSyncer) SyncTasks(ctx context.Context, tasks []domain.Task) ([]domain.SyncResult, error) {
	timer := time.NewTimer(s.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
	}
	out := make([]domain.SyncResult, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, domain.SyncResult{
			TaskID: task.ID,
			Target: "mock",
			Status: "synced",
		})
	}
	return out, nil
}

func TestOrchestratorSyncTimeout(t *testing.T) {
	orch := NewOrchestrator(
		agents.NewRecorderAgent(),
		agents.NewDecisionAgent(),
		agents.NewTaskPlannerAgent(),
		agents.NewReviewerAgent(),
		nil,
		slowSyncer{delay: 200 * time.Millisecond},
		slowSyncer{delay: 200 * time.Millisecond},
	)

	_, err := orch.Run(context.Background(), "行动项：@李雷 明天提交", Options{
		MeetingDate: time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local),
		SyncTarget:  SyncJira,
		SyncTimeout: 10 * time.Millisecond,
	})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want deadline exceeded, got %v", err)
	}
}

func TestOrchestratorHybridLLM(t *testing.T) {
	orch := NewOrchestrator(
		agents.NewRecorderAgent(),
		agents.NewDecisionAgent(),
		agents.NewTaskPlannerAgent(),
		agents.NewReviewerAgent(),
		mockExtractor{},
		syncer.NewJiraClientFromEnv(true),
		syncer.NewNotionClientFromEnv(true),
	)

	result, err := orch.Run(context.Background(), "行动项：@张三 明天补充支付联调", Options{
		MeetingDate: time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local),
		SyncTarget:  SyncNone,
		LLMMode:     LLMHybrid,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(result.Decisions) == 0 {
		t.Fatalf("expected decisions from llm merge")
	}
	found := false
	for _, task := range result.Tasks {
		if task.Title == "支付重构设计评审" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected llm task merged")
	}
}
