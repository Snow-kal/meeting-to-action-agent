package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type JiraFieldMapping struct {
	Summary      string `json:"summary"`
	Description  string `json:"description"`
	DueDate      string `json:"due_date"`
	Owner        string `json:"owner"`
	Dependencies string `json:"dependencies"`
	TaskID       string `json:"task_id"`
}

type NotionFieldMapping struct {
	Title        string `json:"title"`
	Owner        string `json:"owner"`
	DueDate      string `json:"due_date"`
	Description  string `json:"description"`
	Dependencies string `json:"dependencies"`
	TaskID       string `json:"task_id"`
}

type MappingConfig struct {
	Jira   JiraFieldMapping   `json:"jira"`
	Notion NotionFieldMapping `json:"notion"`
}

func DefaultMappingConfig() MappingConfig {
	return MappingConfig{
		Jira: JiraFieldMapping{
			Summary:     "summary",
			Description: "description",
			DueDate:     "duedate",
		},
		Notion: NotionFieldMapping{
			Title:        "Name",
			Owner:        "Owner",
			DueDate:      "Due",
			Description:  "Description",
			Dependencies: "Dependencies",
			TaskID:       "TaskID",
		},
	}
}

func LoadMappingConfig(path string) (MappingConfig, error) {
	cfg := DefaultMappingConfig()
	if strings.TrimSpace(path) == "" {
		return cfg, nil
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return MappingConfig{}, fmt.Errorf("读取映射配置失败: %w", err)
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		return MappingConfig{}, fmt.Errorf("解析映射配置失败: %w", err)
	}
	normalizeMappingDefaults(&cfg)
	return cfg, nil
}

func normalizeMappingDefaults(cfg *MappingConfig) {
	if strings.TrimSpace(cfg.Jira.Summary) == "" {
		cfg.Jira.Summary = "summary"
	}
	if strings.TrimSpace(cfg.Jira.Description) == "" {
		cfg.Jira.Description = "description"
	}
	if strings.TrimSpace(cfg.Jira.DueDate) == "" {
		cfg.Jira.DueDate = "duedate"
	}
	if strings.TrimSpace(cfg.Notion.Title) == "" {
		cfg.Notion.Title = "Name"
	}
	if strings.TrimSpace(cfg.Notion.Owner) == "" {
		cfg.Notion.Owner = "Owner"
	}
	if strings.TrimSpace(cfg.Notion.DueDate) == "" {
		cfg.Notion.DueDate = "Due"
	}
	if strings.TrimSpace(cfg.Notion.Description) == "" {
		cfg.Notion.Description = "Description"
	}
	if strings.TrimSpace(cfg.Notion.Dependencies) == "" {
		cfg.Notion.Dependencies = "Dependencies"
	}
	if strings.TrimSpace(cfg.Notion.TaskID) == "" {
		cfg.Notion.TaskID = "TaskID"
	}
}
