package runtime

import (
	"fmt"

	"github.com/Snow-kal/meeting-to-action-agent/internal/agents"
	"github.com/Snow-kal/meeting-to-action-agent/internal/config"
	"github.com/Snow-kal/meeting-to-action-agent/internal/llm"
	"github.com/Snow-kal/meeting-to-action-agent/internal/pipeline"
	"github.com/Snow-kal/meeting-to-action-agent/internal/syncer"
)

type FactoryOptions struct {
	DryRun            bool
	MaxRetries        int
	MappingConfigPath string
	LLMMode           pipeline.LLMMode
}

func NewOrchestrator(opts FactoryOptions) (*pipeline.Orchestrator, error) {
	mapping, err := config.LoadMappingConfig(opts.MappingConfigPath)
	if err != nil {
		return nil, err
	}

	jiraClient := syncer.NewJiraClientFromEnv(opts.DryRun)
	notionClient := syncer.NewNotionClientFromEnv(opts.DryRun)
	jiraClient.Retry.MaxAttempts = opts.MaxRetries
	notionClient.Retry.MaxAttempts = opts.MaxRetries
	jiraClient.Mapping = mapping.Jira
	notionClient.Mapping = mapping.Notion

	var extractor pipeline.LLMExtractor
	if opts.LLMMode == pipeline.LLMHybrid {
		llmClient := llm.NewOpenAIClientFromEnv()
		if llmClient.APIKey == "" {
			return nil, fmt.Errorf("LLM 混合模式需要配置 OPENAI_API_KEY")
		}
		extractor = llmClient
	}

	return pipeline.NewOrchestrator(
		agents.NewRecorderAgent(),
		agents.NewDecisionAgent(),
		agents.NewTaskPlannerAgent(),
		agents.NewReviewerAgent(),
		extractor,
		jiraClient,
		notionClient,
	), nil
}
