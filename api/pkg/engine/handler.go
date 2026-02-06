package engine

import (
	"encoding/json"
	"sync"
)

// Node represents a node in a workflow graph.
// This is a generic representation that handlers can use.
type Node struct {
	ID       string
	Type     string
	Metadata json.RawMessage
}

// ExecutionStep represents the result of executing a single node
type ExecutionStep struct {
	StepNumber int                    `json:"stepNumber"`
	NodeType   string                 `json:"nodeType"`
	NodeID     string                 `json:"nodeId,omitempty"`
	Status     string                 `json:"status"`
	Duration   int64                  `json:"duration"`
	Output     map[string]interface{} `json:"output"`
	Timestamp  string                 `json:"timestamp"`
	Error      string                 `json:"error,omitempty"`
}

// NodeHandler defines the interface for handling different node types.
// Implementations should be stateless and thread-safe.
type NodeHandler interface {
	// NodeType returns the type of node this handler processes
	NodeType() string

	// Execute processes the node and returns an ExecutionStep.
	// The handler may read from and write to the ExecutionContext.
	Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error)
}

// Registry maps node types to their handlers.
// It is safe for concurrent use.
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]NodeHandler
}

// NewRegistry creates a new empty handler registry
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]NodeHandler),
	}
}

// Register adds a handler for a specific node type.
// If a handler for this type already exists, it will be replaced.
func (r *Registry) Register(handler NodeHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[handler.NodeType()] = handler
}

// Get returns the handler for a given node type
func (r *Registry) Get(nodeType string) (NodeHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[nodeType]
	return h, ok
}

// NodeTypes returns all registered node types
func (r *Registry) NodeTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
