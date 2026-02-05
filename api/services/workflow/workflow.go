package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
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

	startTime := time.Now()
	executionID := uuid.New()

	// Mock temperature based on city (simulating API call)
	temperature := getMockTemperature(req.FormData.City)

	// Evaluate condition
	conditionMet := evaluateCondition(temperature, req.Condition.Operator, req.Condition.Threshold)

	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	steps := buildExecutionSteps(req.FormData, temperature, conditionMet, startTime)

	response := ExecutionResponse{
		ExecutionID:   executionID.String(),
		Status:        "completed",
		StartTime:     startTime.Format(time.RFC3339),
		EndTime:       endTime.Format(time.RFC3339),
		TotalDuration: duration,
		Steps:         steps,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func getMockTemperature(city string) float64 {
	temperatures := map[string]float64{
		"Sydney":    28.5,
		"Melbourne": 22.3,
		"Brisbane":  31.2,
		"Perth":     35.1,
		"Adelaide":  26.8,
	}
	if temp, ok := temperatures[city]; ok {
		return temp
	}
	return 25.0
}

func evaluateCondition(temperature float64, operator string, threshold float64) bool {
	switch operator {
	case "greater_than":
		return temperature > threshold
	case "less_than":
		return temperature < threshold
	case "equals":
		return temperature == threshold
	case "greater_than_or_equal":
		return temperature >= threshold
	case "less_than_or_equal":
		return temperature <= threshold
	default:
		return false
	}
}

func buildExecutionSteps(formData FormData, temperature float64, conditionMet bool, startTime time.Time) []ExecutionStep {
	steps := []ExecutionStep{
		{
			StepNumber: 1,
			NodeType:   "start",
			Status:     "success",
			Duration:   10,
			Output:     StepOutput{Message: "Workflow started"},
			Timestamp:  startTime.Format(time.RFC3339),
		},
		{
			StepNumber: 2,
			NodeType:   "form",
			Status:     "success",
			Duration:   15,
			Output: StepOutput{
				Message:  "User input collected",
				FormData: &formData,
			},
			Timestamp: startTime.Add(10 * time.Millisecond).Format(time.RFC3339),
		},
		{
			StepNumber: 3,
			NodeType:   "integration",
			Status:     "success",
			Duration:   150,
			Output: StepOutput{
				Message: fmt.Sprintf("Fetched weather data for %s", formData.City),
				APIResponse: &APIResponse{
					Endpoint:   "https://api.open-meteo.com/v1/forecast",
					Method:     "GET",
					StatusCode: 200,
					Data: map[string]any{
						"temperature": temperature,
						"city":        formData.City,
					},
				},
			},
			Timestamp: startTime.Add(25 * time.Millisecond).Format(time.RFC3339),
		},
		{
			StepNumber: 4,
			NodeType:   "condition",
			Status:     "success",
			Duration:   5,
			Output: StepOutput{
				Message: fmt.Sprintf("Condition evaluated: temperature %.1f°C %s %.1f°C", temperature, formData.Operator, formData.Threshold),
				ConditionResult: &ConditionResult{
					Expression:  fmt.Sprintf("temperature %s %.1f", formData.Operator, formData.Threshold),
					Result:      conditionMet,
					Temperature: temperature,
					Operator:    formData.Operator,
					Threshold:   formData.Threshold,
				},
			},
			Timestamp: startTime.Add(175 * time.Millisecond).Format(time.RFC3339),
		},
	}

	if conditionMet {
		steps = append(steps, ExecutionStep{
			StepNumber: 5,
			NodeType:   "email",
			Status:     "success",
			Duration:   50,
			Output: StepOutput{
				Message: "Alert email sent",
				EmailContent: &EmailContent{
					To:        formData.Email,
					Subject:   "Weather Alert",
					Body:      fmt.Sprintf("Hi %s, weather alert for %s! Temperature is %.1f°C!", formData.Name, formData.City, temperature),
					Timestamp: startTime.Add(180 * time.Millisecond).Format(time.RFC3339),
				},
			},
			Timestamp: startTime.Add(180 * time.Millisecond).Format(time.RFC3339),
		})
	}

	steps = append(steps, ExecutionStep{
		StepNumber: len(steps) + 1,
		NodeType:   "end",
		Status:     "success",
		Duration:   5,
		Output:     StepOutput{Message: "Workflow completed"},
		Timestamp:  startTime.Add(235 * time.Millisecond).Format(time.RFC3339),
	})

	return steps
}
