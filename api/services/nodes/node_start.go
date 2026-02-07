package nodes

import "context"

// PassthroughNode handles node types with no execution logic (start, end).
// It preserves the raw DB metadata for ToJSON() and is a no-op on Execute().
type PassthroughNode struct {
	base BaseFields
}

func NewPassthroughNode(base BaseFields) *PassthroughNode {
	return &PassthroughNode{base: base}
}

func (n *PassthroughNode) ToJSON() NodeJSON {
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

func (n *PassthroughNode) Execute(_ context.Context, _ *NodeContext) (*ExecutionResult, error) {
	return &ExecutionResult{Status: "completed"}, nil
}
