package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMeetingInputText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meeting.txt")
	content := "行动项：@张三 明天完成"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := LoadMeetingInput(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.Content != content {
		t.Fatalf("unexpected content: %s", got.Content)
	}
	if got.MeetingDate != nil {
		t.Fatalf("unexpected meeting date")
	}
}

func TestLoadMeetingInputJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meeting.json")
	body := `{"content":"决策：确定本周上线","meeting_date":"2026-03-18"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := LoadMeetingInput(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.Content != "决策：确定本周上线" {
		t.Fatalf("unexpected content: %s", got.Content)
	}
	if got.MeetingDate == nil || got.MeetingDate.Format("2006-01-02") != "2026-03-18" {
		t.Fatalf("unexpected meeting date: %+v", got.MeetingDate)
	}
}
