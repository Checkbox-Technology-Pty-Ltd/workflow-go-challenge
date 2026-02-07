package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormNode collects user input. The metadata defines which fields to
// collect (inputFields) and which variables they produce (outputVariables).
// During execution, it reads the expected fields from the runtime context
// (pre-populated from the execute request payload).
type FormNode struct {
	base BaseFields

	InputFields     []string `json:"inputFields"`
	OutputVariables []string `json:"outputVariables"`
}

func NewFormNode(base BaseFields) (*FormNode, error) {
	n := &FormNode{base: base}
	if err := json.Unmarshal(base.Metadata, n); err != nil {
		return nil, fmt.Errorf("invalid form metadata: %w", err)
	}
	return n, nil
}

func (n *FormNode) ToJSON() NodeJSON {
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

// Execute extracts the declared input fields from the runtime context
// and passes them through as output variables for downstream nodes.
func (n *FormNode) Execute(_ context.Context, nCtx *NodeContext) (*ExecutionResult, error) {
	output := make(map[string]any)

	for _, field := range n.InputFields {
		val, ok := nCtx.Variables[field]
		if !ok {
			return nil, fmt.Errorf("missing required form field: %s", field)
		}
		output[field] = val
	}

	return &ExecutionResult{
		Status: "completed",
		Output: output,
	}, nil
}
