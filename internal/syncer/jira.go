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

type JiraClient struct {
	BaseURL    string
	ProjectKey string
	Email      string
	Token      string
	DryRun     bool
	Retry      RetryConfig
	Mapping    config.JiraFieldMapping
	HTTPClient *http.Client
}

func NewJiraClientFromEnv(dryRun bool) *JiraClient {
	return &JiraClient{
		BaseURL:    strings.TrimSpace(os.Getenv("JIRA_API_BASE")),
		ProjectKey: strings.TrimSpace(os.Getenv("JIRA_PROJECT_KEY")),
		Email:      strings.TrimSpace(os.Getenv("JIRA_EMAIL")),
		Token:      strings.TrimSpace(os.Getenv("JIRA_TOKEN")),
		DryRun:     dryRun,
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseBackoff: 200 * time.Millisecond,
		},
		Mapping:    config.DefaultMappingConfig().Jira,
		HTTPClient: &http.Client{},
	}
}

func (c *JiraClient) SyncTasks(ctx context.Context, tasks []domain.Task) ([]domain.SyncResult, error) {
	results := make([]domain.SyncResult, 0, len(tasks))
	if c.DryRun {
		for i, task := range tasks {
			results = append(results, domain.SyncResult{
				TaskID:   task.ID,
				Target:   "jira",
				Status:   "dry-run",
				RemoteID: fmt.Sprintf("JIRA-DRY-%03d", i+1),
			})
		}
		return results, nil
	}

	if c.BaseURL == "" || c.ProjectKey == "" || c.Email == "" || c.Token == "" {
		return nil, fmt.Errorf("jira 配置不完整，请设置 JIRA_API_BASE/JIRA_PROJECT_KEY/JIRA_EMAIL/JIRA_TOKEN")
	}

	for _, task := range tasks {
		var remoteID string
		err := doWithRetry(ctx, c.Retry, func() error {
			var createErr error
			remoteID, createErr = c.createIssue(ctx, task)
			return createErr
		})
		if err != nil {
			results = append(results, domain.SyncResult{
				TaskID: task.ID,
				Target: "jira",
				Status: "failed",
				Error:  err.Error(),
			})
			continue
		}
		results = append(results, domain.SyncResult{
			TaskID:   task.ID,
			Target:   "jira",
			Status:   "synced",
			RemoteID: remoteID,
		})
	}
	return results, nil
}

func (c *JiraClient) createIssue(ctx context.Context, task domain.Task) (string, error) {
	fields := map[string]any{
		"project": map[string]any{
			"key": c.ProjectKey,
		},
	}
	setIfPresent(fields, c.Mapping.Summary, task.Title)
	setIfPresent(fields, c.Mapping.Description, task.Description)
	if task.DueDate != nil {
		setIfPresent(fields, c.Mapping.DueDate, task.DueDate.Format("2006-01-02"))
	}
	if task.Owner != "" {
		setIfPresent(fields, c.Mapping.Owner, task.Owner)
	}
	if len(task.Dependencies) > 0 {
		setIfPresent(fields, c.Mapping.Dependencies, strings.Join(task.Dependencies, ", "))
	}
	if task.ID != "" {
		setIfPresent(fields, c.Mapping.TaskID, task.ID)
	}
	body := map[string]any{"fields": fields}

	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(c.BaseURL, "/") + "/rest/api/3/issue"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.Email, c.Token)

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
			Message:    fmt.Sprintf("jira 返回 %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var parsed struct {
		Key string `json:"key"`
		ID  string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if parsed.Key != "" {
		return parsed.Key, nil
	}
	if parsed.ID != "" {
		return parsed.ID, nil
	}
	return "", fmt.Errorf("jira 返回体缺少 key/id")
}

func setIfPresent(fields map[string]any, key, value string) {
	k := strings.TrimSpace(key)
	if k == "" {
		return
	}
	fields[k] = value
}
