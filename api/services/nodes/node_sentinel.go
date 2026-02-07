package nodes

import "context"

// SentinelNode marks the boundaries of a workflow graph (start, end).
// It preserves the raw DB metadata for ToJSON() and is a no-op on Execute().
type SentinelNode struct {
	base BaseFields
}

func NewSentinelNode(base BaseFields) *SentinelNode {
	return &SentinelNode{base: base}
}

func (n *SentinelNode) ToJSON() NodeJSON {
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

func (n *SentinelNode) Execute(_ context.Context, _ *NodeContext) (*ExecutionResult, error) {
	return &ExecutionResult{Status: "completed"}, nil
}
