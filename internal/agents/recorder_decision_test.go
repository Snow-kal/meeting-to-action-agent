package agents

import (
	"testing"
	"time"
)

func TestRecorderExtractTopicsAndDiscussion(t *testing.T) {
	recorder := NewRecorderAgent()
	record := recorder.Record("会议主题：发布准备\n讨论：回滚方案待确认\n备注：需关注跨团队依赖", time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local))

	if len(record.Topics) != 1 || record.Topics[0] != "发布准备" {
		t.Fatalf("unexpected topics: %+v", record.Topics)
	}
	if len(record.DiscussionPoints) != 2 {
		t.Fatalf("unexpected discussion points: %+v", record.DiscussionPoints)
	}
}

func TestDecisionAgentFiltersAmbiguousDiscussion(t *testing.T) {
	record := NewRecorderAgent().Record("讨论：是否下周上线待确认\n决策：本周冻结版本", time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local))
	decisions := NewDecisionAgent().Extract(record)

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].SourceText == "" || decisions[0].Confidence == 0 {
		t.Fatalf("expected enriched decision fields, got %+v", decisions[0])
	}
}
