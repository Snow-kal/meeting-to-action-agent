package syncer

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Snow-kal/meeting-to-action-agent/internal/config"
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
	var requestBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if _, _, ok := r.BasicAuth(); !ok {
			t.Fatalf("expected basic auth")
		}
		body, _ := io.ReadAll(r.Body)
		requestBody = string(body)
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
		Mapping: config.JiraFieldMapping{
			Summary:      "customfield_10001",
			Description:  "customfield_10002",
			DueDate:      "customfield_10003",
			Owner:        "customfield_10004",
			Dependencies: "customfield_10005",
			TaskID:       "customfield_10006",
		},
		HTTPClient: srv.Client(),
	}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "a", Description: "b", Owner: "张三", Dependencies: []string{"TASK-009"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "synced" || results[0].RemoteID != "PROJ-1" {
		t.Fatalf("unexpected results: %+v", results)
	}
	for _, field := range []string{"customfield_10001", "customfield_10004", "customfield_10006"} {
		if !strings.Contains(requestBody, field) {
			t.Fatalf("request body missing mapped field %s: %s", field, requestBody)
		}
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
	var requestBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/databases/db1":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
  "properties": {
    "任务": {"type":"title"},
    "负责人": {"type":"rich_text"},
    "截止": {"type":"date"},
    "描述": {"type":"rich_text"},
    "依赖": {"type":"rich_text"},
    "任务ID": {"type":"rich_text"}
  }
}`))
			return
		case "/v1/pages":
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatalf("expected authorization header")
		}
		body, _ := io.ReadAll(r.Body)
		requestBody = string(body)
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
		Mapping: config.NotionFieldMapping{
			Title:        "任务",
			Owner:        "负责人",
			DueDate:      "截止",
			Description:  "描述",
			Dependencies: "依赖",
			TaskID:       "任务ID",
		},
		HTTPClient: srv.Client(),
	}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "a", Description: "b", Owner: "张三", Dependencies: []string{"TASK-009"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "synced" || results[0].RemoteID != "notion-page-1" {
		t.Fatalf("unexpected results: %+v", results)
	}
	for _, field := range []string{"任务", "负责人", "任务ID"} {
		if !strings.Contains(requestBody, field) {
			t.Fatalf("request body missing mapped field %s: %s", field, requestBody)
		}
	}
}

func TestNotionClientMinimalSchemaTitleOnly(t *testing.T) {
	var requestBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/databases/db1":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
  "properties": {
    "名称": {"type":"title"}
  }
}`))
			return
		case "/v1/pages":
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		requestBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"notion-page-2"}`))
	}))
	defer srv.Close()

	client := &NotionClient{
		BaseURL:    srv.URL + "/v1/pages",
		Token:      "token",
		DatabaseID: "db1",
		DryRun:     false,
		Mapping: config.NotionFieldMapping{
			Title:        "Name",
			Owner:        "Owner",
			DueDate:      "Due",
			Description:  "Description",
			Dependencies: "Dependencies",
			TaskID:       "TaskID",
		},
		HTTPClient: srv.Client(),
	}
	results, err := client.SyncTasks(context.Background(), []domain.Task{
		{ID: "TASK-001", Title: "示例任务", Description: "描述", Owner: "A"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "synced" {
		t.Fatalf("unexpected results: %+v", results)
	}
	if !strings.Contains(requestBody, "名称") {
		t.Fatalf("request should write title field 名称, body=%s", requestBody)
	}
	if strings.Contains(requestBody, "Owner") || strings.Contains(requestBody, "Description") {
		t.Fatalf("request should skip non-existing properties, body=%s", requestBody)
	}
}
