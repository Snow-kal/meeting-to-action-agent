package pipeline

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/agents"
	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/syncer"
)

type SyncTarget string
type LLMMode string

const (
	SyncNone   SyncTarget = "none"
	SyncJira   SyncTarget = "jira"
	SyncNotion SyncTarget = "notion"
	SyncBoth   SyncTarget = "both"

	LLMOff    LLMMode = "off"
	LLMHybrid LLMMode = "hybrid"
)

type Options struct {
	MeetingDate time.Time
	SyncTarget  SyncTarget
	SyncTimeout time.Duration
	LLMMode     LLMMode
}

type LLMExtractor interface {
	Extract(ctx context.Context, rawText string, meetingDate time.Time) ([]domain.Decision, []domain.Task, error)
}

type Orchestrator struct {
	Recorder *agents.RecorderAgent
	Decision *agents.DecisionAgent
	Planner  *agents.TaskPlannerAgent
	Reviewer *agents.ReviewerAgent
	LLM      LLMExtractor
	Jira     syncer.TaskSyncer
	Notion   syncer.TaskSyncer
}

func NewOrchestrator(
	recorder *agents.RecorderAgent,
	decision *agents.DecisionAgent,
	planner *agents.TaskPlannerAgent,
	reviewer *agents.ReviewerAgent,
	llmExtractor LLMExtractor,
	jira syncer.TaskSyncer,
	notion syncer.TaskSyncer,
) *Orchestrator {
	return &Orchestrator{
		Recorder: recorder,
		Decision: decision,
		Planner:  planner,
		Reviewer: reviewer,
		LLM:      llmExtractor,
		Jira:     jira,
		Notion:   notion,
	}
}

func (o *Orchestrator) Run(ctx context.Context, rawText string, opts Options) (*domain.PipelineResult, error) {
	if strings.TrimSpace(rawText) == "" {
		return nil, fmt.Errorf("会议记录不能为空")
	}

	meetingDate := opts.MeetingDate
	if meetingDate.IsZero() {
		meetingDate = time.Now()
	}

	record := o.Recorder.Record(rawText, meetingDate)
	ruleDecisions := o.Decision.Extract(record)
	decisions := ruleDecisions
	warnings := make([]string, 0)

	llmMode := opts.LLMMode
	if llmMode == "" {
		llmMode = LLMOff
	}
	var llmTasks []domain.Task
	if llmMode == LLMHybrid {
		if o.LLM == nil {
			warnings = append(warnings, "LLM 模式已开启，但未配置 LLM 提取器，已回退到规则模式")
		} else {
			enhancedDecisions, enhancedTasks, err := o.LLM.Extract(ctx, rawText, meetingDate)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("LLM 提取失败，已回退到规则模式: %v", err))
			} else {
				decisions = mergeDecisions(ruleDecisions, enhancedDecisions)
				llmTasks = enhancedTasks
			}
		}
	}

	plannedTasks := o.Planner.Plan(record, decisions)
	plannedTasks = mergeTasks(plannedTasks, llmTasks)
	reviewedTasks, issues := o.Reviewer.Review(plannedTasks, meetingDate)

	result := &domain.PipelineResult{
		Record:    record,
		Decisions: decisions,
		Tasks:     reviewedTasks,
		Issues:    issues,
		Warnings:  warnings,
	}

	switch opts.SyncTarget {
	case SyncJira:
		if o.Jira == nil {
			return nil, fmt.Errorf("jira syncer 未配置")
		}
		synced, err := runSyncWithTimeout(ctx, o.Jira, reviewedTasks, opts.SyncTimeout)
		if err != nil {
			return nil, err
		}
		result.Synced = append(result.Synced, synced...)
	case SyncNotion:
		if o.Notion == nil {
			return nil, fmt.Errorf("notion syncer 未配置")
		}
		synced, err := runSyncWithTimeout(ctx, o.Notion, reviewedTasks, opts.SyncTimeout)
		if err != nil {
			return nil, err
		}
		result.Synced = append(result.Synced, synced...)
	case SyncBoth:
		if o.Jira == nil || o.Notion == nil {
			return nil, fmt.Errorf("jira/notion syncer 未配置")
		}

		var (
			jiraSynced   []domain.SyncResult
			notionSynced []domain.SyncResult
			jiraErr      error
			notionErr    error
		)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			jiraSynced, jiraErr = runSyncWithTimeout(ctx, o.Jira, reviewedTasks, opts.SyncTimeout)
		}()
		go func() {
			defer wg.Done()
			notionSynced, notionErr = runSyncWithTimeout(ctx, o.Notion, reviewedTasks, opts.SyncTimeout)
		}()
		wg.Wait()

		if jiraErr != nil {
			return nil, jiraErr
		}
		if notionErr != nil {
			return nil, notionErr
		}

		result.Synced = append(result.Synced, jiraSynced...)
		result.Synced = append(result.Synced, notionSynced...)
	case SyncNone:
	default:
		return nil, fmt.Errorf("不支持的 sync target: %s", opts.SyncTarget)
	}

	return result, nil
}

func runSyncWithTimeout(
	ctx context.Context,
	client syncer.TaskSyncer,
	tasks []domain.Task,
	timeout time.Duration,
) ([]domain.SyncResult, error) {
	if timeout <= 0 {
		return client.SyncTasks(ctx, tasks)
	}
	syncCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.SyncTasks(syncCtx, tasks)
}

func mergeDecisions(rule, llm []domain.Decision) []domain.Decision {
	result := make([]domain.Decision, 0, len(rule)+len(llm))
	seen := map[string]struct{}{}
	add := func(d domain.Decision) {
		key := normalizeText(d.Text)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, d)
	}
	for _, d := range rule {
		add(d)
	}
	for _, d := range llm {
		add(d)
	}
	for i := range result {
		if strings.TrimSpace(result[i].ID) == "" {
			result[i].ID = fmt.Sprintf("DEC-%03d", i+1)
		}
	}
	return result
}

func mergeTasks(rule, llm []domain.Task) []domain.Task {
	result := make([]domain.Task, 0, len(rule)+len(llm))
	seen := map[string]struct{}{}
	add := func(t domain.Task) {
		key := normalizeText(t.Title) + "|" + normalizeText(t.Owner)
		if key == "|" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, t)
	}
	for _, t := range rule {
		add(t)
	}
	for _, t := range llm {
		add(t)
	}
	for i := range result {
		result[i].ID = fmt.Sprintf("TASK-%03d", i+1)
	}
	return result
}

func normalizeText(s string) string {
	trimmed := strings.TrimSpace(strings.ToLower(s))
	return strings.Join(strings.Fields(trimmed), " ")
}
