package workflow

import (
	"context"
	"fmt"
	"time"
)

// Executor handles workflow execution using graph traversal
type Executor struct {
	registry *HandlerRegistry
}

// NewExecutor creates a new workflow executor with the given handler registry
func NewExecutor(registry *HandlerRegistry) *Executor {
	return &Executor{registry: registry}
}

// WorkflowGraph represents the workflow as an adjacency list
type WorkflowGraph struct {
	nodes     map[string]*Node
	adjacency map[string][]EdgeTarget
	startNode string
}

// EdgeTarget represents an outgoing edge with optional condition handle
type EdgeTarget struct {
	TargetID     string
	SourceHandle string // "true", "false", or empty for unconditional
}

// BuildGraph constructs a WorkflowGraph from nodes and edges
func BuildGraph(nodes []Node, edges []Edge) (*WorkflowGraph, error) {
	graph := &WorkflowGraph{
		nodes:     make(map[string]*Node),
		adjacency: make(map[string][]EdgeTarget),
	}

	// Index nodes by ID
	for i := range nodes {
		node := &nodes[i]
		graph.nodes[node.NodeID] = node

		if node.NodeType == "start" {
			graph.startNode = node.NodeID
		}

		// Initialize adjacency list
		graph.adjacency[node.NodeID] = []EdgeTarget{}
	}

	if graph.startNode == "" {
		return nil, fmt.Errorf("workflow has no start node")
	}

	// Build adjacency list from edges
	for _, edge := range edges {
		handle := ""
		if edge.SourceHandle != nil {
			handle = *edge.SourceHandle
		}

		target := EdgeTarget{
			TargetID:     edge.TargetID,
			SourceHandle: handle,
		}

		graph.adjacency[edge.SourceID] = append(graph.adjacency[edge.SourceID], target)
	}

	return graph, nil
}

// GetNode returns the node with the given ID
func (g *WorkflowGraph) GetNode(id string) (*Node, bool) {
	node, ok := g.nodes[id]
	return node, ok
}

// GetOutgoingEdges returns all outgoing edges from a node
func (g *WorkflowGraph) GetOutgoingEdges(nodeID string) []EdgeTarget {
	return g.adjacency[nodeID]
}

// GetNextNode returns the next node to visit based on condition result
// For condition nodes, uses sourceHandle to select the correct branch
// For other nodes, returns the first (and usually only) target
func (g *WorkflowGraph) GetNextNode(nodeID string, conditionResult *bool) (string, bool) {
	edges := g.adjacency[nodeID]
	if len(edges) == 0 {
		return "", false
	}

	// For non-condition nodes or when no condition result provided
	if conditionResult == nil {
		return edges[0].TargetID, true
	}

	// For condition nodes, select based on sourceHandle
	handleToFind := "false"
	if *conditionResult {
		handleToFind = "true"
	}

	for _, edge := range edges {
		if edge.SourceHandle == handleToFind {
			return edge.TargetID, true
		}
	}

	// Fallback: return first edge if no matching handle found
	return edges[0].TargetID, true
}

// Execute runs the workflow and returns execution steps
func (e *Executor) Execute(ctx context.Context, graph *WorkflowGraph, formData FormData) ([]ExecutionStep, error) {
	var steps []ExecutionStep

	ec := &ExecutionContext{
		Ctx:        ctx,
		FormData:   formData,
		StartTime:  time.Now(),
		StepNumber: 0,
	}

	currentNodeID := graph.startNode
	visited := make(map[string]bool)

	for currentNodeID != "" {
		// Prevent infinite loops
		if visited[currentNodeID] {
			return nil, fmt.Errorf("cycle detected at node %s", currentNodeID)
		}
		visited[currentNodeID] = true

		node, ok := graph.GetNode(currentNodeID)
		if !ok {
			return nil, fmt.Errorf("node not found: %s", currentNodeID)
		}

		handler, ok := e.registry.Get(node.NodeType)
		if !ok {
			return nil, fmt.Errorf("no handler for node type: %s", node.NodeType)
		}

		ec.StepNumber++
		step, err := handler.Execute(ec, node)
		if err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			steps = append(steps, step)
			return steps, err
		}

		steps = append(steps, step)

		// Determine next node
		var conditionResult *bool
		if node.NodeType == "condition" && step.Output.ConditionResult != nil {
			result := step.Output.ConditionResult.Result
			conditionResult = &result
		}

		nextNodeID, hasNext := graph.GetNextNode(currentNodeID, conditionResult)
		if !hasNext || node.NodeType == "end" {
			break
		}

		currentNodeID = nextNodeID
	}

	return steps, nil
}
