package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// ConditionNode evaluates a condition expression against runtime variables.
// It outputs conditionMet (bool) and sets Branch to "true" or "false",
// which the execution engine uses to follow the correct outgoing edge.
type ConditionNode struct {
	base BaseFields

	ConditionExpression string   `json:"conditionExpression"`
	OutputVariables     []string `json:"outputVariables"`
}

func NewConditionNode(base BaseFields) (*ConditionNode, error) {
	n := &ConditionNode{base: base}
	if err := json.Unmarshal(base.Metadata, n); err != nil {
		return nil, fmt.Errorf("invalid condition metadata: %w", err)
	}
	return n, nil
}

func (n *ConditionNode) ToJSON() NodeJSON {
	return NodeJSON{
		ID:       n.base.ID,
		Type:     n.base.NodeType,
		Position: n.base.Position,
		Data: NodeData{
			Label:       n.base.Label,
			Description: n.base.Description,
			Metadata:    n.base.Metadata,
		},
	}
}

// Execute evaluates the condition using operator and threshold from context.
// The expression "temperature {{operator}} {{threshold}}" is resolved by
// reading the actual temperature, operator, and threshold from variables.
func (n *ConditionNode) Execute(_ context.Context, nCtx *NodeContext) (*ExecutionResult, error) {
	temperature, ok := toFloat64(nCtx.Variables["temperature"])
	if !ok {
		return nil, fmt.Errorf("missing or invalid variable: temperature")
	}

	operator, _ := nCtx.Variables["operator"].(string)
	if operator == "" {
		operator = "greater_than" // default
	}

	threshold, ok := toFloat64(nCtx.Variables["threshold"])
	if !ok {
		threshold = 25 // default
	}

	conditionMet, err := evaluate(temperature, operator, threshold)
	if err != nil {
		return nil, err
	}

	branch := "false"
	if conditionMet {
		branch = "true"
	}

	return &ExecutionResult{
		Status: "completed",
		Branch: branch,
		Output: map[string]any{
			"conditionMet": conditionMet,
			"threshold":    threshold,
			"operator":     operator,
			"actualValue":  temperature,
			"message": fmt.Sprintf(
				"Temperature %.1f°C is %s %.1f°C - condition %s",
				temperature, operator, threshold, branchLabel(conditionMet),
			),
		},
	}, nil
}

func evaluate(value float64, operator string, threshold float64) (bool, error) {
	switch operator {
	case "greater_than":
		return value > threshold, nil
	case "less_than":
		return value < threshold, nil
	case "equal_to":
		return value == threshold, nil
	case "greater_than_or_equal":
		return value >= threshold, nil
	case "less_than_or_equal":
		return value <= threshold, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

func branchLabel(met bool) string {
	if met {
		return "met"
	}
	return "not met"
}

// toFloat64 converts a variable to float64, handling both float64 and
// json.Number types that may appear depending on the source.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
