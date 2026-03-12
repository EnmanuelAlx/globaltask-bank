package repository

import (
	"context"

	"github.com/globaltask/bank/internal/domain/entity"
)

type WorkflowProviderRepository interface {
	GetByWorkflowAndStep(ctx context.Context, workflowName string, eventStep string) ([]*entity.WorkflowProvider, error)
}

type workflowProviderRepo struct {
	db DBTX
}

func NewWorkflowProviderRepository(db DBTX) WorkflowProviderRepository {
	return &workflowProviderRepo{db: db}
}

func (r *workflowProviderRepo) GetByWorkflowAndStep(ctx context.Context, workflowName string, eventStep string) ([]*entity.WorkflowProvider, error) {
	query := `
		SELECT id, workflow_name, provider_id, event_step, endpoint_path, priority, is_active, created_at, updated_at 
		FROM workflow_providers 
		WHERE workflow_name = $1 AND event_step = $2 AND is_active = true
		ORDER BY priority ASC
	`
	rows, err := r.db.Query(ctx, query, workflowName, eventStep)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*entity.WorkflowProvider
	for rows.Next() {
		var wp entity.WorkflowProvider
		err := rows.Scan(
			&wp.ID,
			&wp.WorkflowName,
			&wp.ProviderID,
			&wp.EventStep,
			&wp.EndpointPath,
			&wp.Priority,
			&wp.IsActive,
			&wp.CreatedAt,
			&wp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		providers = append(providers, &wp)
	}

	return providers, nil
}
