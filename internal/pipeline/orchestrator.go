package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/agents"
	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/syncer"
)

type SyncTarget string

const (
	SyncNone   SyncTarget = "none"
	SyncJira   SyncTarget = "jira"
	SyncNotion SyncTarget = "notion"
	SyncBoth   SyncTarget = "both"
)

type Options struct {
	MeetingDate time.Time
	SyncTarget  SyncTarget
}

type Orchestrator struct {
	Recorder *agents.RecorderAgent
	Decision *agents.DecisionAgent
	Planner  *agents.TaskPlannerAgent
	Reviewer *agents.ReviewerAgent
	Jira     syncer.TaskSyncer
	Notion   syncer.TaskSyncer
}

func NewOrchestrator(
	recorder *agents.RecorderAgent,
	decision *agents.DecisionAgent,
	planner *agents.TaskPlannerAgent,
	reviewer *agents.ReviewerAgent,
	jira syncer.TaskSyncer,
	notion syncer.TaskSyncer,
) *Orchestrator {
	return &Orchestrator{
		Recorder: recorder,
		Decision: decision,
		Planner:  planner,
		Reviewer: reviewer,
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
	decisions := o.Decision.Extract(record)
	plannedTasks := o.Planner.Plan(record, decisions)
	reviewedTasks, issues := o.Reviewer.Review(plannedTasks, meetingDate)

	result := &domain.PipelineResult{
		Record:    record,
		Decisions: decisions,
		Tasks:     reviewedTasks,
		Issues:    issues,
	}

	switch opts.SyncTarget {
	case SyncJira:
		if o.Jira == nil {
			return nil, fmt.Errorf("jira syncer 未配置")
		}
		synced, err := o.Jira.SyncTasks(ctx, reviewedTasks)
		if err != nil {
			return nil, err
		}
		result.Synced = append(result.Synced, synced...)
	case SyncNotion:
		if o.Notion == nil {
			return nil, fmt.Errorf("notion syncer 未配置")
		}
		synced, err := o.Notion.SyncTasks(ctx, reviewedTasks)
		if err != nil {
			return nil, err
		}
		result.Synced = append(result.Synced, synced...)
	case SyncBoth:
		if o.Jira == nil || o.Notion == nil {
			return nil, fmt.Errorf("jira/notion syncer 未配置")
		}
		jiraSynced, err := o.Jira.SyncTasks(ctx, reviewedTasks)
		if err != nil {
			return nil, err
		}
		notionSynced, err := o.Notion.SyncTasks(ctx, reviewedTasks)
		if err != nil {
			return nil, err
		}
		result.Synced = append(result.Synced, jiraSynced...)
		result.Synced = append(result.Synced, notionSynced...)
	case SyncNone:
	default:
		return nil, fmt.Errorf("不支持的 sync target: %s", opts.SyncTarget)
	}

	return result, nil
}
