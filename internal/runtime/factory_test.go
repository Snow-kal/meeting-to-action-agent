package runtime

import (
	"testing"

	"github.com/Snow-kal/meeting-to-action-agent/internal/pipeline"
)

func TestNewOrchestratorHybridRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := NewOrchestrator(FactoryOptions{
		DryRun:     true,
		MaxRetries: 1,
		LLMMode:    pipeline.LLMHybrid,
	})
	if err == nil {
		t.Fatalf("expected error when OPENAI_API_KEY is missing")
	}
}
