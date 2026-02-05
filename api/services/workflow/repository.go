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
	GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error)
	GetNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Node, error)
	GetEdgesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Edge, error)
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
