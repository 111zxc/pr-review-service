package domain

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypePRCreated          EventType = "pr_created"
	EventTypePRMerged           EventType = "pr_merged"
	EventTypeReviewerAssigned   EventType = "reviewer_assigned"
	EventTypeReviewerReassigned EventType = "reviewer_reassigned"
	EventTypeReviewerUnassigned EventType = "reviewer_unassigned"
)

type Event struct {
	ID        int       `json:"id"`
	EventType EventType `json:"event_type"`

	PRID   string `json:"pr_id"`
	UserID string `json:"user_id"`

	AdditionalData json.RawMessage `json:"additional_data,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type StatsResponse struct {
	EventCounts map[EventType]int `json:"event_counts"`
	TotalEvents int               `json:"total_events"`
}

type ReviewerAssignedData struct {
	AssignedAt time.Time `json:"assigned_at"`
}

type ReviewerReassignedData struct {
	OldUserID    string    `json:"old_user_id"`
	NewUserID    string    `json:"new_user_id"`
	ReassignedAt time.Time `json:"reassigned_at"`
}

type PRCreatedData struct {
	PRName    string    `json:"pr_name"`
	CreatedAt time.Time `json:"created_at"`
}

type PRMergedData struct {
	MergedAt time.Time `json:"merged_at"`
}
