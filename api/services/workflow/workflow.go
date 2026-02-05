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

// TODO: Update this
func (s *Service) HandleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Handling workflow execution for id", "id", id)

	// Generate current timestamp
	currentTime := time.Now().Format(time.RFC3339)

	executionJSON := fmt.Sprintf(`{
		"executedAt": "%s",
		"status": "completed",
		"steps": [
			{
				"nodeId": "start",
				"type": "start",
				"label": "Start",
				"description": "Begin weather check workflow",
				"status": "completed"
			},
			{
				"nodeId": "form",
				"type": "form",
				"label": "User Input",
				"description": "Process collected data - name, email, location",
				"status": "completed",
				"output": {
					"name": "Alice",
					"email": "alice@example.com",
					"city": "Sydney"
				}
			},
			{
				"nodeId": "weather-api",
				"type": "integration",
				"label": "Weather API",
				"description": "Fetch current temperature for Sydney",
				"status": "completed",
				"output": {
					"temperature": 28.5,
					"location": "Sydney"
				}
			},
			{
				"nodeId": "condition",
				"type": "condition",
				"label": "Check Condition",
				"description": "Evaluate temperature threshold",
				"status": "completed",
				"output": {
					"conditionMet": true,
					"threshold": 25,
					"operator": "greater_than",
					"actualValue": 28.5,
					"message": "Temperature 28.5°C is greater than 25°C - condition met"
				}
			},
			{
				"nodeId": "email",
				"type": "email",
				"label": "Send Alert",
				"description": "Email weather alert notification",
				"status": "completed",
				"output": {
					"emailDraft": {
						"to": "alice@example.com",
						"from": "weather-alerts@example.com",
						"subject": "Weather Alert",
						"body": "Weather alert for Sydney! Temperature is 28.5°C!",
						"timestamp": "2024-01-15T14:30:24.856Z"
					},
					"deliveryStatus": "sent",
					"messageId": "msg_abc123def456",
					"emailSent": true
				}
			},
			{
				"nodeId": "end",
				"type": "end",
				"label": "Complete",
				"description": "Workflow execution finished",
				"status": "completed"
			}
		]
	}`, currentTime)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(executionJSON))
}
