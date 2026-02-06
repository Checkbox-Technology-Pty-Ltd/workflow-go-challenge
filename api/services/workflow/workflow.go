package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"workflow-code-test/api/pkg/engine"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

// emailRegex is a basic email format validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// HandleGetWorkflow godoc
// @Summary Get workflow definition
// @Description Retrieves a workflow by ID including all nodes and edges
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} WorkflowResponse
// @Failure 400 {string} string "invalid workflow id"
// @Failure 404 {string} string "workflow not found"
// @Failure 500 {string} string "internal server error"
// @Router /api/v1/workflows/{id} [get]
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

// HandleExecuteWorkflow godoc
// @Summary Execute a workflow
// @Description Executes a workflow with the provided form data and returns execution results
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Param request body ExecuteWorkflowRequest true "Execution request with form data"
// @Success 200 {object} ExecutionResponse
// @Failure 400 {string} string "invalid workflow id, invalid request body, or missing required fields"
// @Failure 404 {string} string "workflow not found"
// @Failure 500 {string} string "internal server error"
// @Router /api/v1/workflows/{id}/execute [post]
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

	// Validate required form data fields
	if req.FormData.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if !emailRegex.MatchString(req.FormData.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	if req.FormData.City == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}
	if !IsSupportedCity(req.FormData.City) {
		http.Error(w, fmt.Sprintf("unsupported city: %s (supported: Sydney, Melbourne, Brisbane, Perth, Adelaide)", req.FormData.City), http.StatusBadRequest)
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

	// Build engine graph from workflow nodes/edges
	graph := buildEngineGraph(nodes, edges)
	if err := graph.Validate(); err != nil {
		slog.Error("invalid workflow structure", "error", err)
		http.Error(w, "invalid workflow structure", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	executionID := uuid.New()

	// Build initial state from form data
	initialState := map[string]interface{}{
		"form.name":  req.FormData.Name,
		"form.email": req.FormData.Email,
		"form.city":  req.FormData.City,
	}

	// Execute workflow using graph traversal
	engineSteps, err := s.executor.Execute(ctx, graph, initialState)
	steps := convertEngineSteps(engineSteps)
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	executionTrace, err := json.Marshal(steps)
	if err != nil {
		slog.Error("failed to marshal execution trace", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// HandleGetExecutions godoc
// @Summary Get workflow execution history
// @Description Retrieves a list of past executions for a workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} ExecutionsListResponse
// @Failure 400 {string} string "invalid workflow id"
// @Failure 404 {string} string "workflow not found"
// @Failure 500 {string} string "internal server error"
// @Router /api/v1/workflows/{id}/executions [get]
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

// buildEngineGraph converts workflow nodes/edges to an engine.Graph
func buildEngineGraph(nodes []Node, edges []Edge) *engine.Graph {
	graph := engine.NewGraph()

	for _, n := range nodes {
		graph.AddNode(&engine.Node{
			ID:       n.NodeID,
			Type:     n.NodeType,
			Metadata: n.Metadata,
		})
	}

	for _, e := range edges {
		sourceHandle := ""
		if e.SourceHandle != nil {
			sourceHandle = *e.SourceHandle
		}
		graph.AddEdge(engine.Edge{
			SourceID:     e.SourceID,
			TargetID:     e.TargetID,
			SourceHandle: sourceHandle,
		})
	}

	return graph
}

// convertEngineSteps converts engine.ExecutionStep to workflow.ExecutionStep for HTTP response
func convertEngineSteps(engineSteps []engine.ExecutionStep) []ExecutionStep {
	steps := make([]ExecutionStep, len(engineSteps))
	for i, es := range engineSteps {
		steps[i] = ExecutionStep{
			StepNumber: es.StepNumber,
			NodeType:   es.NodeType,
			Status:     es.Status,
			Duration:   es.Duration,
			Timestamp:  es.Timestamp,
			Error:      es.Error,
			Output:     convertStepOutput(es.Output),
		}
	}
	return steps
}

// convertStepOutput converts engine output map to typed StepOutput
func convertStepOutput(output map[string]interface{}) StepOutput {
	result := StepOutput{}

	if msg, ok := output["message"].(string); ok {
		result.Message = msg
	}

	if fd, ok := output["formData"].(map[string]interface{}); ok {
		result.FormData = &FormData{
			Name:  getString(fd, "name"),
			Email: getString(fd, "email"),
			City:  getString(fd, "city"),
		}
	}

	if ar, ok := output["apiResponse"].(map[string]interface{}); ok {
		result.APIResponse = &APIResponse{
			Endpoint:   getString(ar, "endpoint"),
			Method:     getString(ar, "method"),
			StatusCode: getInt(ar, "statusCode"),
			Data:       ar["data"],
		}
	}

	if cr, ok := output["conditionResult"].(map[string]interface{}); ok {
		result.ConditionResult = &ConditionResult{
			Expression:  getString(cr, "expression"),
			Result:      getBool(cr, "result"),
			Temperature: getFloat(cr, "temperature"),
			Operator:    getString(cr, "operator"),
			Threshold:   getFloat(cr, "threshold"),
		}
	}

	if ec, ok := output["emailContent"].(map[string]interface{}); ok {
		result.EmailContent = &EmailContent{
			To:        getString(ec, "to"),
			Subject:   getString(ec, "subject"),
			Body:      getString(ec, "body"),
			Timestamp: getString(ec, "timestamp"),
		}
	}

	return result
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
