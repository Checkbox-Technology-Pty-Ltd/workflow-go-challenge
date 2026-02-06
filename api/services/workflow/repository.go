package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for workflow data access
type Repository interface {
	// Workflow operations
	GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error)
	GetNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Node, error)
	GetEdgesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Edge, error)

	// Execution operations
	CreateExecution(ctx context.Context, exec *WorkflowExecution) error
	GetExecution(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error)
	GetExecutionsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]WorkflowExecution, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository
func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetWorkflow retrieves a workflow by ID
func (r *PostgresRepository) GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error) {
	query := `
		SELECT id, name, description, version, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query workflow %s: %w", id, err)
	}
	defer rows.Close()

	workflow, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Workflow])
	if err != nil {
		return nil, fmt.Errorf("scan workflow %s: %w", id, err)
	}

	return &workflow, nil
}

// GetNodesByWorkflowID retrieves all nodes for a workflow
func (r *PostgresRepository) GetNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Node, error) {
	query := `
		SELECT workflow_id, node_id, node_type, label, description, x_pos, y_pos, metadata
		FROM nodes
		WHERE workflow_id = $1
	`

	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("query nodes for workflow %s: %w", workflowID, err)
	}
	defer rows.Close()

	nodes, err := pgx.CollectRows(rows, pgx.RowToStructByName[Node])
	if err != nil {
		return nil, fmt.Errorf("scan nodes for workflow %s: %w", workflowID, err)
	}

	return nodes, nil
}

// GetEdgesByWorkflowID retrieves all edges for a workflow
func (r *PostgresRepository) GetEdgesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Edge, error) {
	query := `
		SELECT workflow_id, edge_id, source_id, target_id, source_handle, edge_props
		FROM edges
		WHERE workflow_id = $1
	`

	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("query edges for workflow %s: %w", workflowID, err)
	}
	defer rows.Close()

	edges, err := pgx.CollectRows(rows, pgx.RowToStructByName[Edge])
	if err != nil {
		return nil, fmt.Errorf("scan edges for workflow %s: %w", workflowID, err)
	}

	return edges, nil
}

// CreateExecution inserts a new workflow execution record
func (r *PostgresRepository) CreateExecution(ctx context.Context, exec *WorkflowExecution) error {
	query := `
		INSERT INTO workflow_executions (id, workflow_id, status, final_context, execution_trace)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		exec.ID,
		exec.WorkflowID,
		exec.Status,
		exec.FinalContext,
		exec.ExecutionTrace,
	)
	if err != nil {
		return fmt.Errorf("insert execution %s: %w", exec.ID, err)
	}

	return nil
}

// GetExecution retrieves an execution by ID
func (r *PostgresRepository) GetExecution(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, status, executed_at, final_context, execution_trace
		FROM workflow_executions
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query execution %s: %w", id, err)
	}
	defer rows.Close()

	exec, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[WorkflowExecution])
	if err != nil {
		return nil, fmt.Errorf("scan execution %s: %w", id, err)
	}

	return &exec, nil
}

// GetExecutionsByWorkflowID retrieves all executions for a workflow
func (r *PostgresRepository) GetExecutionsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, status, executed_at, final_context, execution_trace
		FROM workflow_executions
		WHERE workflow_id = $1
		ORDER BY executed_at DESC
	`

	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("query executions for workflow %s: %w", workflowID, err)
	}
	defer rows.Close()

	execs, err := pgx.CollectRows(rows, pgx.RowToStructByName[WorkflowExecution])
	if err != nil {
		return nil, fmt.Errorf("scan executions for workflow %s: %w", workflowID, err)
	}

	return execs, nil
}
