package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// MockRepository implements Repository for testing.
// All methods panic if the corresponding function is not set,
// ensuring tests explicitly configure the behavior they expect.
type MockRepository struct {
	GetWorkflowFunc               func(ctx context.Context, id uuid.UUID) (*Workflow, error)
	GetNodesByWorkflowIDFunc      func(ctx context.Context, workflowID uuid.UUID) ([]Node, error)
	GetEdgesByWorkflowIDFunc      func(ctx context.Context, workflowID uuid.UUID) ([]Edge, error)
	CreateExecutionFunc           func(ctx context.Context, exec *WorkflowExecution) error
	GetExecutionFunc              func(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error)
	GetExecutionsByWorkflowIDFunc func(ctx context.Context, workflowID uuid.UUID) ([]WorkflowExecution, error)
}

func (m *MockRepository) GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error) {
	if m.GetWorkflowFunc == nil {
		panic(fmt.Sprintf("MockRepository.GetWorkflow called but GetWorkflowFunc not set (id: %s)", id))
	}
	return m.GetWorkflowFunc(ctx, id)
}

func (m *MockRepository) GetNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Node, error) {
	if m.GetNodesByWorkflowIDFunc == nil {
		panic(fmt.Sprintf("MockRepository.GetNodesByWorkflowID called but GetNodesByWorkflowIDFunc not set (workflowID: %s)", workflowID))
	}
	return m.GetNodesByWorkflowIDFunc(ctx, workflowID)
}

func (m *MockRepository) GetEdgesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Edge, error) {
	if m.GetEdgesByWorkflowIDFunc == nil {
		panic(fmt.Sprintf("MockRepository.GetEdgesByWorkflowID called but GetEdgesByWorkflowIDFunc not set (workflowID: %s)", workflowID))
	}
	return m.GetEdgesByWorkflowIDFunc(ctx, workflowID)
}

func (m *MockRepository) CreateExecution(ctx context.Context, exec *WorkflowExecution) error {
	if m.CreateExecutionFunc == nil {
		panic("MockRepository.CreateExecution called but CreateExecutionFunc not set")
	}
	return m.CreateExecutionFunc(ctx, exec)
}

func (m *MockRepository) GetExecution(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error) {
	if m.GetExecutionFunc == nil {
		panic(fmt.Sprintf("MockRepository.GetExecution called but GetExecutionFunc not set (id: %s)", id))
	}
	return m.GetExecutionFunc(ctx, id)
}

func (m *MockRepository) GetExecutionsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]WorkflowExecution, error) {
	if m.GetExecutionsByWorkflowIDFunc == nil {
		panic(fmt.Sprintf("MockRepository.GetExecutionsByWorkflowID called but GetExecutionsByWorkflowIDFunc not set (workflowID: %s)", workflowID))
	}
	return m.GetExecutionsByWorkflowIDFunc(ctx, workflowID)
}
