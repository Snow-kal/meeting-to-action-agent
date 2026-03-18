package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/pipeline"
	"github.com/Snow-kal/meeting-to-action-agent/internal/report"
	"github.com/Snow-kal/meeting-to-action-agent/internal/runtime"
)

type ServerOptions struct {
	DryRun            bool
	MaxRetries        int
	SyncTimeout       time.Duration
	SyncTarget        pipeline.SyncTarget
	LLMMode           pipeline.LLMMode
	MappingConfigPath string
}

type Server struct {
	opts ServerOptions
}

type RunRequest struct {
	Content           string `json:"content"`
	MeetingDate       string `json:"meeting_date,omitempty"`
	SyncTarget        string `json:"sync_target,omitempty"`
	DryRun            *bool  `json:"dry_run,omitempty"`
	MaxRetries        *int   `json:"max_retries,omitempty"`
	SyncTimeout       string `json:"sync_timeout,omitempty"`
	LLMMode           string `json:"llm_mode,omitempty"`
	MappingConfigPath string `json:"mapping_config_path,omitempty"`
	IncludeReport     bool   `json:"include_report,omitempty"`
}

type RunResponse struct {
	Result   any    `json:"result"`
	ReportMD string `json:"report_md,omitempty"`
}

func NewServer(opts ServerOptions) *Server {
	return &Server{opts: opts}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/run", s.handleRun)
	mux.HandleFunc("/healthz", s.handleHealthz)
	return withCORS(mux)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	meetingDate := time.Now()
	if strings.TrimSpace(req.MeetingDate) != "" {
		parsed, err := time.Parse("2006-01-02", req.MeetingDate)
		if err != nil {
			http.Error(w, "meeting_date format should be YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		meetingDate = parsed
	}

	effective, err := s.resolveOptions(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orch, err := runtime.NewOrchestrator(runtime.FactoryOptions{
		DryRun:            effective.DryRun,
		MaxRetries:        effective.MaxRetries,
		MappingConfigPath: effective.MappingConfigPath,
		LLMMode:           effective.LLMMode,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("init failed: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := orch.Run(ctx, req.Content, pipeline.Options{
		MeetingDate: meetingDate,
		SyncTarget:  effective.SyncTarget,
		SyncTimeout: effective.SyncTimeout,
		LLMMode:     effective.LLMMode,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errorsIsContext(err) {
			status = http.StatusGatewayTimeout
		}
		http.Error(w, fmt.Sprintf("run failed: %v", err), status)
		return
	}

	resp := RunResponse{Result: result}
	if req.IncludeReport {
		resp.ReportMD = report.BuildMarkdown(result)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) resolveOptions(req RunRequest) (ServerOptions, error) {
	out := s.opts
	if req.DryRun != nil {
		out.DryRun = *req.DryRun
	}
	if req.MaxRetries != nil {
		out.MaxRetries = *req.MaxRetries
	}
	if strings.TrimSpace(req.SyncTimeout) != "" {
		d, err := time.ParseDuration(req.SyncTimeout)
		if err != nil {
			return ServerOptions{}, fmt.Errorf("sync_timeout 格式错误: %w", err)
		}
		out.SyncTimeout = d
	}
	if strings.TrimSpace(req.SyncTarget) != "" {
		out.SyncTarget = pipeline.SyncTarget(req.SyncTarget)
	}
	if strings.TrimSpace(req.LLMMode) != "" {
		out.LLMMode = pipeline.LLMMode(req.LLMMode)
	}
	if strings.TrimSpace(req.MappingConfigPath) != "" {
		out.MappingConfigPath = strings.TrimSpace(req.MappingConfigPath)
	}
	return out, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "marshal response failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		next.ServeHTTP(w, r)
	})
}

func errorsIsContext(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
