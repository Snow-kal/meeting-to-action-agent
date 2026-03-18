package input

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MeetingInput struct {
	Content     string
	MeetingDate *time.Time
}

type jsonMeetingInput struct {
	Content     string `json:"content"`
	RawText     string `json:"raw_text"`
	Text        string `json:"text"`
	MeetingDate string `json:"meeting_date"`
	Date        string `json:"date"`
}

func LoadMeetingInput(path string) (MeetingInput, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return MeetingInput{}, fmt.Errorf("读取输入文件失败: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return parseJSONInput(body)
	case ".txt", ".md", ".markdown":
		return MeetingInput{Content: string(body)}, nil
	default:
		// 默认按文本处理，避免因为扩展名不规范而阻断流程。
		return MeetingInput{Content: string(body)}, nil
	}
}

func parseJSONInput(body []byte) (MeetingInput, error) {
	var payload jsonMeetingInput
	if err := json.Unmarshal(body, &payload); err != nil {
		return MeetingInput{}, fmt.Errorf("json 输入解析失败: %w", err)
	}

	content := firstNonEmpty(payload.Content, payload.RawText, payload.Text)
	if strings.TrimSpace(content) == "" {
		return MeetingInput{}, fmt.Errorf("json 输入缺少 content/raw_text/text")
	}

	dateRaw := firstNonEmpty(payload.MeetingDate, payload.Date)
	if strings.TrimSpace(dateRaw) == "" {
		return MeetingInput{Content: content}, nil
	}

	parsed, ok := parseDate(dateRaw)
	if !ok {
		return MeetingInput{}, fmt.Errorf("json meeting_date/date 解析失败: %s", dateRaw)
	}
	return MeetingInput{
		Content:     content,
		MeetingDate: &parsed,
	}, nil
}

func firstNonEmpty(candidates ...string) string {
	for _, c := range candidates {
		if strings.TrimSpace(c) != "" {
			return strings.TrimSpace(c)
		}
	}
	return ""
}

func parseDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
