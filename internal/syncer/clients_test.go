package syncer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

func TestJiraClientDryRun(t *testing.T) {
	client := &JiraClient{DryRun: true}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "a", Description: "b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "dry-run" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestJiraClientLive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if _, _, ok := r.BasicAuth(); !ok {
			t.Fatalf("expected basic auth")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"10001","key":"PROJ-1"}`))
	}))
	defer srv.Close()

	client := &JiraClient{
		BaseURL:    srv.URL,
		ProjectKey: "PROJ",
		Email:      "a@b.com",
		Token:      "token",
		DryRun:     false,
		HTTPClient: srv.Client(),
	}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "a", Description: "b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "synced" || results[0].RemoteID != "PROJ-1" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestNotionClientDryRun(t *testing.T) {
	client := &NotionClient{DryRun: true}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "a", Description: "b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "dry-run" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestNotionClientLive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatalf("expected authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"notion-page-1"}`))
	}))
	defer srv.Close()

	client := &NotionClient{
		BaseURL:    srv.URL + "/v1/pages",
		Token:      "token",
		DatabaseID: "db1",
		DryRun:     false,
		HTTPClient: srv.Client(),
	}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "a", Description: "b", Owner: "张三"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "synced" || results[0].RemoteID != "notion-page-1" {
		t.Fatalf("unexpected results: %+v", results)
	}
}
