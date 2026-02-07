package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"workflow-code-test/api/pkg/engine"
)

// ConditionHandler handles condition nodes.
// It reads operator and threshold from node metadata,
// and temperature from state (set by WeatherHandler).
type ConditionHandler struct{}

// NewConditionHandler creates a new ConditionHandler
func NewConditionHandler() *ConditionHandler {
	return &ConditionHandler{}
}

func (h *ConditionHandler) NodeType() string { return "condition" }

func (h *ConditionHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	startTime := time.Now()

	// Try to get operator and threshold from execution state first (from request),
	// then fall back to node metadata (from workflow definition)
	operator := ec.GetString("condition.operator")
	threshold := ec.GetFloat("condition.threshold")

	// Fall back to node metadata if not in state
	if operator == "" && len(node.Metadata) > 0 {
		var metadata struct {
			Operator  string  `json:"operator"`
			Threshold float64 `json:"threshold"`
		}
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return engine.ExecutionStep{}, fmt.Errorf("failed to parse condition metadata: %w", err)
		}
		operator = metadata.Operator
		threshold = metadata.Threshold
	}

	if operator == "" {
		return engine.ExecutionStep{}, fmt.Errorf("operator not specified in condition (provide in request or node metadata)")
	}

	temperature := ec.GetFloat("weather.temperature")

	result := evaluateCondition(temperature, operator, threshold)

	duration := time.Since(startTime).Milliseconds()

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "condition",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   duration,
		Output: map[string]interface{}{
			"message": fmt.Sprintf("Condition evaluated: temperature %.1f°C %s %.1f°C", temperature, operator, threshold),
			"conditionResult": map[string]interface{}{
				"expression":  fmt.Sprintf("temperature %s %.1f", operator, threshold),
				"result":      result,
				"temperature": temperature,
				"operator":    operator,
				"threshold":   threshold,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
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
