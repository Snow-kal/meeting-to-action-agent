package syncer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/config"
	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

type NotionClient struct {
	BaseURL    string
	Token      string
	DatabaseID string
	DryRun     bool
	Retry      RetryConfig
	Mapping    config.NotionFieldMapping
	HTTPClient *http.Client
}

func NewNotionClientFromEnv(dryRun bool) *NotionClient {
	baseURL := strings.TrimSpace(os.Getenv("NOTION_API_BASE"))
	if baseURL == "" {
		baseURL = "https://api.notion.com/v1/pages"
	}
	return &NotionClient{
		BaseURL:    baseURL,
		Token:      strings.TrimSpace(os.Getenv("NOTION_TOKEN")),
		DatabaseID: strings.TrimSpace(os.Getenv("NOTION_DATABASE_ID")),
		DryRun:     dryRun,
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseBackoff: 200 * time.Millisecond,
		},
		Mapping:    config.DefaultMappingConfig().Notion,
		HTTPClient: &http.Client{},
	}
}

func (c *NotionClient) SyncTasks(ctx context.Context, tasks []domain.Task) ([]domain.SyncResult, error) {
	results := make([]domain.SyncResult, 0, len(tasks))
	if c.DryRun {
		for i, task := range tasks {
			results = append(results, domain.SyncResult{
				TaskID:   task.ID,
				Target:   "notion",
				Status:   "dry-run",
				RemoteID: fmt.Sprintf("NOTION-DRY-%03d", i+1),
			})
		}
		return results, nil
	}

	if c.Token == "" || c.DatabaseID == "" {
		return nil, fmt.Errorf("notion 配置不完整，请设置 NOTION_TOKEN/NOTION_DATABASE_ID")
	}

	for _, task := range tasks {
		var remoteID string
		err := doWithRetry(ctx, c.Retry, func() error {
			var createErr error
			remoteID, createErr = c.createPage(ctx, task)
			return createErr
		})
		if err != nil {
			results = append(results, domain.SyncResult{
				TaskID: task.ID,
				Target: "notion",
				Status: "failed",
				Error:  err.Error(),
			})
			continue
		}
		results = append(results, domain.SyncResult{
			TaskID:   task.ID,
			Target:   "notion",
			Status:   "synced",
			RemoteID: remoteID,
		})
	}

	return results, nil
}

func (c *NotionClient) createPage(ctx context.Context, task domain.Task) (string, error) {
	props := map[string]any{}
	setNotionTitle(props, c.Mapping.Title, task.Title)
	setNotionRichText(props, c.Mapping.Owner, task.Owner)
	setNotionRichText(props, c.Mapping.Description, task.Description)
	setNotionRichText(props, c.Mapping.TaskID, task.ID)
	if len(task.Dependencies) > 0 {
		setNotionRichText(props, c.Mapping.Dependencies, strings.Join(task.Dependencies, ", "))
	}
	if task.DueDate != nil {
		dueField := strings.TrimSpace(c.Mapping.DueDate)
		if dueField == "" {
			dueField = "Due"
		}
		props[dueField] = map[string]any{
			"date": map[string]any{
				"start": task.DueDate.Format("2006-01-02"),
			},
		}
	}

	payload := map[string]any{
		"parent": map[string]any{
			"database_id": c.DatabaseID,
		},
		"properties": props,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Notion-Version", "2022-06-28")

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("notion 返回 %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var parsed struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if parsed.ID == "" {
		return "", fmt.Errorf("notion 返回体缺少 id")
	}
	return parsed.ID, nil
}

func setNotionTitle(props map[string]any, fieldName, value string) {
	name := strings.TrimSpace(fieldName)
	if name == "" {
		return
	}
	props[name] = map[string]any{
		"title": []map[string]any{
			{
				"text": map[string]any{
					"content": value,
				},
			},
		},
	}
}

func setNotionRichText(props map[string]any, fieldName, value string) {
	name := strings.TrimSpace(fieldName)
	if name == "" || strings.TrimSpace(value) == "" {
		return
	}
	props[name] = map[string]any{
		"rich_text": []map[string]any{
			{
				"text": map[string]any{
					"content": value,
				},
			},
		},
	}
}
