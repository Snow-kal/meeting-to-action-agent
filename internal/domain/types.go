package domain

import "time"

type MeetingRecord struct {
	RawText          string    `json:"raw_text"`
	Lines            []string  `json:"lines"`
	MeetingDate      time.Time `json:"meeting_date"`
	Topics           []string  `json:"topics,omitempty"`
	DiscussionPoints []string  `json:"discussion_points,omitempty"`
}

type Decision struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	OwnerHint  string  `json:"owner_hint,omitempty"`
	DueHint    string  `json:"due_hint,omitempty"`
	SourceText string  `json:"source_text,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type Task struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Owner              string     `json:"owner"`
	DueDate            *time.Time `json:"due_date,omitempty"`
	Dependencies       []string   `json:"dependencies,omitempty"`
	SourceDecisionID   string     `json:"source_decision_id,omitempty"`
	SourceText         string     `json:"source_text,omitempty"`
	AcceptanceCriteria string     `json:"acceptance_criteria,omitempty"`
	RiskFlags          []string   `json:"risk_flags,omitempty"`
	Confidence         float64    `json:"confidence,omitempty"`
	OwnerInferred      bool       `json:"owner_inferred,omitempty"`
	DueDateInferred    bool       `json:"due_date_inferred,omitempty"`
}

type ReviewIssue struct {
	TaskID    string `json:"task_id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Severity  string `json:"severity,omitempty"`
	Inference bool   `json:"inference,omitempty"`
}

type Conflict struct {
	Type     string   `json:"type"`
	TaskIDs  []string `json:"task_ids,omitempty"`
	Message  string   `json:"message"`
	Severity string   `json:"severity,omitempty"`
}

type SyncResult struct {
	TaskID   string `json:"task_id"`
	Target   string `json:"target"`
	Status   string `json:"status"`
	RemoteID string `json:"remote_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type PipelineResult struct {
	MeetingSummary    string        `json:"meeting_summary,omitempty"`
	Record            MeetingRecord `json:"record"`
	Decisions         []Decision    `json:"decisions"`
	Tasks             []Task        `json:"tasks"`
	Issues            []ReviewIssue `json:"issues"`
	Conflicts         []Conflict    `json:"conflicts,omitempty"`
	FollowUpQuestions []string      `json:"follow_up_questions,omitempty"`
	Synced            []SyncResult  `json:"synced,omitempty"`
	Warnings          []string      `json:"warnings,omitempty"`
}
