package timeutil

import (
	"testing"
	"time"
)

func TestExtractDueDate(t *testing.T) {
	base := time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "absolute", text: "请在 2026-03-25 前完成", want: "2026-03-25"},
		{name: "tomorrow", text: "明天给出方案", want: "2026-03-19"},
		{name: "next_weekday", text: "下周一交付", want: "2026-03-23"},
		{name: "month_end", text: "月底前上线", want: "2026-03-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ExtractDueDate(tt.text, base)
			if !ok || got == nil {
				t.Fatalf("expected due date, got nil")
			}
			if got.Format("2006-01-02") != tt.want {
				t.Fatalf("want %s, got %s", tt.want, got.Format("2006-01-02"))
			}
		})
	}
}
