package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMappingConfig_Default(t *testing.T) {
	cfg, err := LoadMappingConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jira.Summary != "summary" {
		t.Fatalf("unexpected jira summary field: %s", cfg.Jira.Summary)
	}
	if cfg.Notion.Title != "Name" {
		t.Fatalf("unexpected notion title field: %s", cfg.Notion.Title)
	}
}

func TestLoadMappingConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mapping.json")
	body := `{
  "jira":{"summary":"customfield_10001","description":"customfield_10002","due_date":"customfield_10003","owner":"customfield_10004"},
  "notion":{"title":"任务名","owner":"负责人","due_date":"截止日期"}
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cfg, err := LoadMappingConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Jira.Summary != "customfield_10001" {
		t.Fatalf("unexpected jira summary field: %s", cfg.Jira.Summary)
	}
	if cfg.Notion.Title != "任务名" {
		t.Fatalf("unexpected notion title field: %s", cfg.Notion.Title)
	}
}
