package report

import (
	"strings"
	"testing"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

func TestBuildMarkdown(t *testing.T) {
	due := time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local)
	result := &domain.PipelineResult{
		MeetingSummary: "会议围绕发布准备展开，形成 1 条决策与 1 条任务。",
		Record: domain.MeetingRecord{
			MeetingDate: time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local),
		},
		Decisions: []domain.Decision{
			{ID: "DEC-001", Text: "确定本周发布", SourceText: "决策：确定本周发布"},
		},
		Tasks: []domain.Task{
			{ID: "TASK-001", Title: "准备发布说明", Owner: "张三", DueDate: &due, AcceptanceCriteria: "说明已提交", RiskFlags: []string{"cross_team"}},
		},
		Conflicts: []domain.Conflict{
			{Type: "owner_overload", Message: "张三任务过载"},
		},
		FollowUpQuestions: []string{"请确认回滚方案负责人。"},
	}

	md := BuildMarkdown(result)
	for _, mustContain := range []string{"# Meeting To Action Report", "DEC-001", "TASK-001", "张三", "Meeting Summary", "owner_overload", "请确认回滚方案负责人"} {
		if !strings.Contains(md, mustContain) {
			t.Fatalf("report missing %s", mustContain)
		}
	}
}
