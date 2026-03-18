package llm

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

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
	"github.com/Snow-kal/meeting-to-action-agent/internal/timeutil"
)

type Extractor interface {
	Extract(ctx context.Context, rawText string, meetingDate time.Time) ([]domain.Decision, []domain.Task, error)
}

type OpenAIClient struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

func NewOpenAIClientFromEnv() *OpenAIClient {
	baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = "gpt-4.1-mini"
	}
	return &OpenAIClient{
		APIKey:     strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: &http.Client{},
	}
}

type llmOutput struct {
	Decisions []struct {
		Text      string `json:"text"`
		OwnerHint string `json:"owner_hint"`
		DueHint   string `json:"due_hint"`
	} `json:"decisions"`
	Tasks []struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Owner        string   `json:"owner"`
		DueHint      string   `json:"due_hint"`
		Dependencies []string `json:"dependencies"`
	} `json:"tasks"`
}

func (c *OpenAIClient) Extract(ctx context.Context, rawText string, meetingDate time.Time) ([]domain.Decision, []domain.Task, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return nil, nil, fmt.Errorf("OPENAI_API_KEY 未配置")
	}

	reqBody, err := c.buildRequest(rawText, meetingDate)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("llm 返回 %d: %s", resp.StatusCode, string(body))
	}

	content, err := parseChatCompletionContent(body)
	if err != nil {
		return nil, nil, err
	}

	var out llmOutput
	if err := json.Unmarshal([]byte(stripCodeFence(content)), &out); err != nil {
		return nil, nil, fmt.Errorf("llm 输出 json 解析失败: %w", err)
	}

	decisions := make([]domain.Decision, 0, len(out.Decisions))
	for i, d := range out.Decisions {
		if strings.TrimSpace(d.Text) == "" {
			continue
		}
		decisions = append(decisions, domain.Decision{
			ID:        fmt.Sprintf("LLM-DEC-%03d", i+1),
			Text:      strings.TrimSpace(d.Text),
			OwnerHint: strings.TrimSpace(d.OwnerHint),
			DueHint:   strings.TrimSpace(d.DueHint),
		})
	}

	tasks := make([]domain.Task, 0, len(out.Tasks))
	for i, t := range out.Tasks {
		title := strings.TrimSpace(t.Title)
		if title == "" {
			continue
		}
		due, _ := timeutil.ExtractDueDate(t.DueHint, meetingDate)
		tasks = append(tasks, domain.Task{
			ID:           fmt.Sprintf("LLM-TASK-%03d", i+1),
			Title:        title,
			Description:  strings.TrimSpace(t.Description),
			Owner:        strings.TrimSpace(t.Owner),
			DueDate:      due,
			Dependencies: normalizeDependencies(t.Dependencies),
		})
	}

	return decisions, tasks, nil
}

func (c *OpenAIClient) buildRequest(rawText string, meetingDate time.Time) ([]byte, error) {
	systemPrompt := "You are an assistant that extracts decisions and actionable tasks from Chinese meeting notes. Return JSON only."
	userPrompt := fmt.Sprintf(`meeting_date: %s
meeting_notes:
%s

Output JSON schema:
{
  "decisions":[{"text":"string","owner_hint":"string","due_hint":"string"}],
  "tasks":[{"title":"string","description":"string","owner":"string","due_hint":"string","dependencies":["string"]}]
}
`, meetingDate.Format("2006-01-02"), rawText)

	payload := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}
	return json.Marshal(payload)
}

func parseChatCompletionContent(body []byte) (string, error) {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm 响应无 choices")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("llm 响应内容为空")
	}
	return content, nil
}

func stripCodeFence(s string) string {
	out := strings.TrimSpace(s)
	out = strings.TrimPrefix(out, "```json")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	return strings.TrimSpace(out)
}

func normalizeDependencies(deps []string) []string {
	out := make([]string, 0, len(deps))
	seen := map[string]struct{}{}
	for _, dep := range deps {
		trimmed := strings.TrimSpace(dep)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
