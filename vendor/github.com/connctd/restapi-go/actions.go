package restapi

import "time"

type ActionRequest struct {
	ID       string              `json:"id"`
	Status   ActionRequestStatus `json:"status"`
	Error    string              `json:"error,omitempty"`
	Deadline time.Time           `json:"deadline"`
}

type ActionRequestStatus string

const (
	ActionRequestStatusPending   ActionRequestStatus = "PENDING"
	ActionRequestStatusCompleted ActionRequestStatus = "COMPLETED"
	ActionRequestStatusFailed    ActionRequestStatus = "FAILED"
	ActionRequestStatusCanceled  ActionRequestStatus = "CANCELED"
)
