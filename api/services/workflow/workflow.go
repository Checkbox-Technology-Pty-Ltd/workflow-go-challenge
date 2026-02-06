package workflow

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func (s *Service) HandleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := mux.Vars(r)["id"]
	slog.Debug("Fetching workflow", "id", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid workflow id", http.StatusBadRequest)
		return
	}

	workflow, err := s.repo.GetWorkflow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "workflow not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get workflow", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	nodes, err := s.repo.GetNodesByWorkflowID(ctx, id)
	if err != nil {
		slog.Error("failed to get nodes", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	edges, err := s.repo.GetEdgesByWorkflowID(ctx, id)
	if err != nil {
		slog.Error("failed to get edges", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	nodeResponses := make([]NodeResponse, len(nodes))
	for i := range nodes {
		nodeResponses[i] = nodes[i].ToResponse()
	}

	edgeResponses := make([]EdgeResponse, len(edges))
	for i := range edges {
		edgeResponses[i] = edges[i].ToResponse()
	}

	response := WorkflowResponse{
		ID:    workflow.ID,
		Nodes: nodeResponses,
		Edges: edgeResponses,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (s *Service) HandleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := mux.Vars(r)["id"]
	slog.Debug("Executing workflow", "id", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid workflow id", http.StatusBadRequest)
		return
	}

	var req ExecuteWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err = s.repo.GetWorkflow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "workflow not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get workflow", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch nodes and edges for graph construction
	nodes, err := s.repo.GetNodesByWorkflowID(ctx, id)
	if err != nil {
		slog.Error("failed to get nodes", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	edges, err := s.repo.GetEdgesByWorkflowID(ctx, id)
	if err != nil {
		slog.Error("failed to get edges", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Build workflow graph
	graph, err := BuildGraph(nodes, edges)
	if err != nil {
		slog.Error("failed to build graph", "error", err)
		http.Error(w, "invalid workflow structure", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	executionID := uuid.New()

	// Execute workflow using graph traversal
	steps, err := s.executor.Execute(ctx, graph, req.FormData)
	status := "completed"
	if err != nil {
		slog.Error("workflow execution failed", "error", err)
		status = "failed"
		// Continue with partial results if we have any steps
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	response := ExecutionResponse{
		ExecutionID:   executionID.String(),
		Status:        status,
		StartTime:     startTime.Format(time.RFC3339),
		EndTime:       endTime.Format(time.RFC3339),
		TotalDuration: duration,
		Steps:         steps,
	}

	// Persist execution to database
	finalContext, err := json.Marshal(req.FormData)
	if err != nil {
		slog.Error("failed to marshal final context", "error", err)
		// Not returning error to client, but logging it
	}

	executionTrace, err := json.Marshal(steps)
	if err != nil {
		slog.Error("failed to marshal execution trace", "error", err)
		// Not returning error to client, but logging it
	}

	exec := &WorkflowExecution{
		ID:             executionID,
		WorkflowID:     &id,
		Status:         status,
		FinalContext:   finalContext,
		ExecutionTrace: executionTrace,
	}

	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		slog.Error("failed to save execution", "error", err)
		// Continue returning response even if persistence fails
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (s *Service) HandleGetExecutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := mux.Vars(r)["id"]
	slog.Debug("Fetching executions", "workflow_id", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid workflow id", http.StatusBadRequest)
		return
	}

	// Verify workflow exists
	_, err = s.repo.GetWorkflow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "workflow not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get workflow", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	executions, err := s.repo.GetExecutionsByWorkflowID(ctx, id)
	if err != nil {
		slog.Error("failed to get executions", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	summaries := make([]ExecutionSummary, len(executions))
	for i, exec := range executions {
		workflowID := ""
		if exec.WorkflowID != nil {
			workflowID = exec.WorkflowID.String()
		}
		summaries[i] = ExecutionSummary{
			ID:         exec.ID.String(),
			WorkflowID: workflowID,
			Status:     exec.Status,
			ExecutedAt: exec.ExecutedAt.Format(time.RFC3339),
		}
	}

	response := ExecutionsListResponse{Executions: summaries}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
