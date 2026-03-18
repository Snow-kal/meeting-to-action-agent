package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/pipeline"
)

func TestRunEndpoint(t *testing.T) {
	server := NewServer(ServerOptions{
		DryRun:      true,
		MaxRetries:  2,
		SyncTimeout: 10 * time.Second,
		SyncTarget:  pipeline.SyncNone,
		LLMMode:     pipeline.LLMOff,
	})

	body := map[string]any{
		"content":        "行动项：@张三 明天提交上线计划",
		"meeting_date":   "2026-03-18",
		"sync_target":    "none",
		"include_report": true,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Result struct {
			Tasks []any `json:"tasks"`
		} `json:"result"`
		ReportMD string `json:"report_md"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Result.Tasks) == 0 {
		t.Fatalf("expected tasks")
	}
	if resp.ReportMD == "" {
		t.Fatalf("expected report")
	}
}

func TestRunEndpointBadRequest(t *testing.T) {
	server := NewServer(ServerOptions{
		DryRun:      true,
		MaxRetries:  2,
		SyncTimeout: 10 * time.Second,
		SyncTarget:  pipeline.SyncNone,
		LLMMode:     pipeline.LLMOff,
	})

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}

func TestWebIndexPage(t *testing.T) {
	server := NewServer(ServerOptions{
		DryRun:      true,
		MaxRetries:  2,
		SyncTimeout: 10 * time.Second,
		SyncTarget:  pipeline.SyncNone,
		LLMMode:     pipeline.LLMOff,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("Meeting To Action Console")) {
		t.Fatalf("index page missing expected title")
	}
}
