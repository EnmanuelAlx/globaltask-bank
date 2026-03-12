package entity

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowProvider represents the mapping between a workflow step and a bank provider
type WorkflowProvider struct {
	ID           uuid.UUID `json:"id"`
	WorkflowName string    `json:"workflow_name"`
	ProviderID   int       `json:"provider_id"`
	EventStep    string    `json:"event_step"`
	EndpointPath string    `json:"endpoint_path"`
	Priority     int       `json:"priority"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
