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
	for _, line := range rawLines {
		normalized := normalizeLine(line)
		if normalized != "" {
			lines = append(lines, normalized)
		}
	}

	return domain.MeetingRecord{
		RawText:     rawText,
		Lines:       lines,
		MeetingDate: meetingDate,
	}
}

func normalizeLine(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimLeft(s, "-*•0123456789. ")
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}
