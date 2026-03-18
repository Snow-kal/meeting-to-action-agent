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
		Record: domain.MeetingRecord{
			MeetingDate: time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local),
		},
		Decisions: []domain.Decision{
			{ID: "DEC-001", Text: "确定本周发布"},
		},
		Tasks: []domain.Task{
			{ID: "TASK-001", Title: "准备发布说明", Owner: "张三", DueDate: &due},
		},
	}

	md := BuildMarkdown(result)
	for _, mustContain := range []string{"# Meeting To Action Report", "DEC-001", "TASK-001", "张三"} {
		if !strings.Contains(md, mustContain) {
			t.Fatalf("report missing %s", mustContain)
		}
	}
}
