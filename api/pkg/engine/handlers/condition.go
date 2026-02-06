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

	// Parse metadata for operator and threshold
	var metadata struct {
		Operator  string  `json:"operator"`
		Threshold float64 `json:"threshold"`
	}

	if len(node.Metadata) > 0 {
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return engine.ExecutionStep{}, fmt.Errorf("failed to parse condition metadata: %w", err)
		}
	}

	if metadata.Operator == "" {
		return engine.ExecutionStep{}, fmt.Errorf("operator not specified in condition node metadata")
	}

	temperature := ec.GetFloat("weather.temperature")

	result := evaluateCondition(temperature, metadata.Operator, metadata.Threshold)

	duration := time.Since(startTime).Milliseconds()

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "condition",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   duration,
		Output: map[string]interface{}{
			"message": fmt.Sprintf("Condition evaluated: temperature %.1f°C %s %.1f°C", temperature, metadata.Operator, metadata.Threshold),
			"conditionResult": map[string]interface{}{
				"expression":  fmt.Sprintf("temperature %s %.1f", metadata.Operator, metadata.Threshold),
				"result":      result,
				"temperature": temperature,
				"operator":    metadata.Operator,
				"threshold":   metadata.Threshold,
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
