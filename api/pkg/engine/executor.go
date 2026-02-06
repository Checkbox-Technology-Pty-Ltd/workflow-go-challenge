package engine

import (
	"context"
	"fmt"
)

// Edge represents a connection between two nodes
type Edge struct {
	SourceID     string
	TargetID     string
	SourceHandle string // "true", "false", or empty for unconditional
}

// Graph represents a workflow as an adjacency list
type Graph struct {
	nodes     map[string]*Node
	adjacency map[string][]Edge
	startNode string
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		nodes:     make(map[string]*Node),
		adjacency: make(map[string][]Edge),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node *Node) {
	g.nodes[node.ID] = node
	if _, exists := g.adjacency[node.ID]; !exists {
		g.adjacency[node.ID] = []Edge{}
	}
	if node.Type == "start" {
		g.startNode = node.ID
	}
}

// AddEdge adds an edge between two nodes
func (g *Graph) AddEdge(edge Edge) {
	g.adjacency[edge.SourceID] = append(g.adjacency[edge.SourceID], edge)
}

// Validate checks that the graph has a valid structure
func (g *Graph) Validate() error {
	if g.startNode == "" {
		return fmt.Errorf("graph has no start node")
	}
	return nil
}

// GetNode returns the node with the given ID
func (g *Graph) GetNode(id string) (*Node, bool) {
	node, ok := g.nodes[id]
	return node, ok
}

// GetStartNode returns the start node ID
func (g *Graph) GetStartNode() string {
	return g.startNode
}

// GetOutgoingEdges returns all outgoing edges from a node
func (g *Graph) GetOutgoingEdges(nodeID string) []Edge {
	return g.adjacency[nodeID]
}

// Executor handles workflow execution using graph traversal
type Executor struct {
	registry *Registry
}

// NewExecutor creates a new workflow executor with the given handler registry
func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute runs the workflow and returns execution steps.
// The initialState map is copied into the ExecutionContext before execution.
func (e *Executor) Execute(ctx context.Context, graph *Graph, initialState map[string]interface{}) ([]ExecutionStep, error) {
	if err := graph.Validate(); err != nil {
		return nil, err
	}

	var steps []ExecutionStep

	ec := NewExecutionContext(ctx)

	// Copy initial state
	for k, v := range initialState {
		ec.Set(k, v)
	}

	currentNodeID := graph.GetStartNode()
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

		handler, ok := e.registry.Get(node.Type)
		if !ok {
			return nil, fmt.Errorf("no handler for node type: %s", node.Type)
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

		// Determine next node based on condition result if present
		nextNodeID := e.getNextNode(graph, currentNodeID, step)
		if node.Type == "end" {
			break
		}

		currentNodeID = nextNodeID
	}

	return steps, nil
}

// getNextNode determines the next node to visit based on the current step's output
func (e *Executor) getNextNode(graph *Graph, currentNodeID string, step ExecutionStep) string {
	edges := graph.GetOutgoingEdges(currentNodeID)
	if len(edges) == 0 {
		return ""
	}

	// Check if this is a condition node with a result
	if step.Output != nil {
		if condResult, ok := step.Output["conditionResult"].(map[string]interface{}); ok {
			if result, ok := condResult["result"].(bool); ok {
				handleToFind := "false"
				if result {
					handleToFind = "true"
				}

				for _, edge := range edges {
					if edge.SourceHandle == handleToFind {
						return edge.TargetID
					}
				}
			}
		}
	}

	// Default: return first edge target
	return edges[0].TargetID
}
