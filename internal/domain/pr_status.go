package domain

import "time"

type PRStatus struct {
	ID          int       `json:"-"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)
