package agents

import (
	"strings"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

type RecorderAgent struct{}

func NewRecorderAgent() *RecorderAgent {
	return &RecorderAgent{}
}

func (a *RecorderAgent) Record(rawText string, meetingDate time.Time) domain.MeetingRecord {
	cleaned := strings.ReplaceAll(rawText, "\r\n", "\n")
	rawLines := strings.Split(cleaned, "\n")
	lines := make([]string, 0, len(rawLines))
	topics := make([]string, 0)
	discussions := make([]string, 0)
	for _, line := range rawLines {
		normalized := normalizeLine(line)
		if normalized != "" {
			lines = append(lines, normalized)
			if topic := extractTopic(normalized); topic != "" {
				topics = append(topics, topic)
			}
			if isDiscussionPoint(normalized) {
				discussions = append(discussions, normalized)
			}
		}
	}

	return domain.MeetingRecord{
		RawText:          rawText,
		Lines:            lines,
		MeetingDate:      meetingDate,
		Topics:           dedupeStrings(topics),
		DiscussionPoints: dedupeStrings(discussions),
	}
}

func normalizeLine(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimLeft(s, "-*•0123456789. ")
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func extractTopic(line string) string {
	prefixes := []string{"会议主题：", "会议主题:", "议题：", "议题:", "Topic:", "topic:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func isDiscussionPoint(line string) bool {
	keywords := []string{"讨论", "问题", "风险", "备注", "阻塞", "待确认"}
	for _, kw := range keywords {
		if strings.Contains(line, kw) {
			return true
		}
	}
	return false
}

func dedupeStrings(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
