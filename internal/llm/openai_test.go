package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAIClientExtract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "choices":[
    {
      "message":{
        "content":"{\"decisions\":[{\"text\":\"确定下周上线\",\"owner_hint\":\"张三\",\"due_hint\":\"下周一\"}],\"tasks\":[{\"title\":\"准备发布清单\",\"description\":\"补齐回滚方案\",\"owner\":\"李四\",\"due_hint\":\"明天\",\"dependencies\":[\"TASK-101\"]}]}"
      }
    }
  ]
}`))
	}))
	defer srv.Close()

	client := &OpenAIClient{
		APIKey:     "token",
		BaseURL:    srv.URL,
		Model:      "mock-model",
		HTTPClient: srv.Client(),
	}

	decisions, tasks, err := client.Extract(context.Background(), "会议记录", time.Date(2026, 3, 18, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if len(decisions) != 1 || len(tasks) != 1 {
		t.Fatalf("unexpected result: decisions=%d tasks=%d", len(decisions), len(tasks))
	}
	if tasks[0].DueDate == nil || tasks[0].DueDate.Format("2006-01-02") != "2026-03-19" {
		t.Fatalf("unexpected due date: %+v", tasks[0].DueDate)
	}
}
