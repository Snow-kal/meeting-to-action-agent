package domain

import "time"

type MeetingRecord struct {
	RawText     string    `json:"raw_text"`
	Lines       []string  `json:"lines"`
	MeetingDate time.Time `json:"meeting_date"`
}

type Decision struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	OwnerHint string `json:"owner_hint,omitempty"`
	DueHint   string `json:"due_hint,omitempty"`
}

type Task struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Owner            string     `json:"owner"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	Dependencies     []string   `json:"dependencies,omitempty"`
	SourceDecisionID string     `json:"source_decision_id,omitempty"`
}

type ReviewIssue struct {
	TaskID  string `json:"task_id"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type SyncResult struct {
	TaskID   string `json:"task_id"`
	Target   string `json:"target"`
	Status   string `json:"status"`
	RemoteID string `json:"remote_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type PipelineResult struct {
	Record    MeetingRecord `json:"record"`
	Decisions []Decision    `json:"decisions"`
	Tasks     []Task        `json:"tasks"`
	Issues    []ReviewIssue `json:"issues"`
	Synced    []SyncResult  `json:"synced,omitempty"`
}
