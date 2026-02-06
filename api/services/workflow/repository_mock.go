package workflow

import (
	"context"

	"github.com/google/uuid"
)

// MockRepository implements Repository for testing
type MockRepository struct {
	GetWorkflowFunc               func(ctx context.Context, id uuid.UUID) (*Workflow, error)
	GetNodesByWorkflowIDFunc      func(ctx context.Context, workflowID uuid.UUID) ([]Node, error)
	GetEdgesByWorkflowIDFunc      func(ctx context.Context, workflowID uuid.UUID) ([]Edge, error)
	CreateExecutionFunc           func(ctx context.Context, exec *WorkflowExecution) error
	GetExecutionFunc              func(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error)
	GetExecutionsByWorkflowIDFunc func(ctx context.Context, workflowID uuid.UUID) ([]WorkflowExecution, error)
}

func (m *MockRepository) GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error) {
	if m.GetWorkflowFunc != nil {
		return m.GetWorkflowFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Node, error) {
	if m.GetNodesByWorkflowIDFunc != nil {
		return m.GetNodesByWorkflowIDFunc(ctx, workflowID)
	}
	return nil, nil
}

func (m *MockRepository) GetEdgesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Edge, error) {
	if m.GetEdgesByWorkflowIDFunc != nil {
		return m.GetEdgesByWorkflowIDFunc(ctx, workflowID)
	}
	return nil, nil
}

func (m *MockRepository) CreateExecution(ctx context.Context, exec *WorkflowExecution) error {
	if m.CreateExecutionFunc != nil {
		return m.CreateExecutionFunc(ctx, exec)
	}
	return nil
}

func (m *MockRepository) GetExecution(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error) {
	if m.GetExecutionFunc != nil {
		return m.GetExecutionFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetExecutionsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]WorkflowExecution, error) {
	if m.GetExecutionsByWorkflowIDFunc != nil {
		return m.GetExecutionsByWorkflowIDFunc(ctx, workflowID)
	}
	return nil, nil
}
