package agents

import (
	"fmt"
	"strings"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

type DecisionAgent struct {
	keywords []string
}

func NewDecisionAgent() *DecisionAgent {
	return &DecisionAgent{
		keywords: []string{"决定", "决策", "结论", "拍板", "确定", "agreed", "decision"},
	}
}

func (a *DecisionAgent) Extract(record domain.MeetingRecord) []domain.Decision {
	decisions := make([]domain.Decision, 0)
	index := 1
	for _, line := range record.Lines {
		if !a.isDecisionLine(line) {
			continue
		}
		decisions = append(decisions, domain.Decision{
			ID:         fmt.Sprintf("DEC-%03d", index),
			Text:       cleanupDecisionText(line),
			OwnerHint:  extractOwner(line),
			DueHint:    extractDueHint(line),
			SourceText: line,
			Confidence: decisionConfidence(line),
		})
		index++
	}
	return decisions
}

func (a *DecisionAgent) isDecisionLine(line string) bool {
	lower := strings.ToLower(line)
	if isAmbiguousDecision(line) {
		return false
	}
	for _, kw := range a.keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func cleanupDecisionText(line string) string {
	result := strings.TrimSpace(line)
	prefixes := []string{"决策：", "决策:", "决定：", "决定:", "结论：", "结论:", "Decision:", "decision:"}
	for _, p := range prefixes {
		result = strings.TrimPrefix(result, p)
	}
	return strings.TrimSpace(result)
}

func isAmbiguousDecision(line string) bool {
	keywords := []string{"讨论", "建议", "待确认", "是否", "考虑", "可能", "待评估"}
	for _, kw := range keywords {
		if strings.Contains(line, kw) && !strings.HasPrefix(line, "决策") && !strings.HasPrefix(line, "决定") {
			return true
		}
	}
	return false
}

func extractDueHint(text string) string {
	keywords := []string{"今天", "明天", "后天", "本周", "下周", "月底", "月", "-"}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return text
		}
	}
	return ""
}

func decisionConfidence(line string) float64 {
	if strings.HasPrefix(line, "决策") || strings.HasPrefix(line, "决定") || strings.HasPrefix(line, "结论") {
		return 0.92
	}
	return 0.76
}
