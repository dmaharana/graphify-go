package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type NodeType string

const (
	FileNode     NodeType = "file"
	FunctionNode NodeType = "function"
	StructNode   NodeType = "struct"
	InterfaceNode NodeType = "interface"
)

type Node struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Name     string                 `json:"name"`
	FilePath string                 `json:"file_path,omitempty"`
	Line     int                    `json:"line,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // e.g., "contains", "calls", "imports"
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: []Node{},
		Edges: []Edge{},
	}
}

func (g *Graph) AddNode(n Node) {
	g.Nodes = append(g.Nodes, n)
}

func (g *Graph) AddEdge(from, to, edgeType string) {
	g.Edges = append(g.Edges, Edge{From: from, To: to, Type: edgeType})
}

func (g *Graph) Save(filename string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (g *Graph) Query(question string) *Graph {
	// For MVP, we use a simple keyword-based subgraph extraction.
	// 1. Identify "seed" nodes that match keywords in the question.
	// 2. Include their direct neighbors (1-hop expansion).
	// 3. If "connects" or "path" is in the question, we look for paths (not implemented in MVP but could be).

	keywords := strings.Fields(strings.ToLower(question))
	seedNodes := make(map[string]bool)
	
	for _, node := range g.Nodes {
		nodeLower := strings.ToLower(node.Name)
		for _, kw := range keywords {
			if len(kw) > 3 && strings.Contains(nodeLower, kw) {
				seedNodes[node.ID] = true
				break
			}
		}
	}

	// Expand to 1-hop neighbors
	relevantNodes := make(map[string]bool)
	for id := range seedNodes {
		relevantNodes[id] = true
	}

	relevantEdges := []Edge{}
	for _, edge := range g.Edges {
		if seedNodes[edge.From] || seedNodes[edge.To] {
			relevantNodes[edge.From] = true
			relevantNodes[edge.To] = true
			relevantEdges = append(relevantEdges, edge)
		}
	}

	subgraph := NewGraph()
	for _, node := range g.Nodes {
		if relevantNodes[node.ID] {
			subgraph.AddNode(node)
		}
	}
	subgraph.Edges = relevantEdges

	return subgraph
}

func (g *Graph) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("### Subgraph Query Results\n\n")

	sb.WriteString("#### Nodes\n")
	for _, n := range g.Nodes {
		sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", n.Name, n.Type))
		if n.FilePath != "" {
			sb.WriteString(fmt.Sprintf("  - Source: `%s`", n.FilePath))
			if n.Line > 0 {
				sb.WriteString(fmt.Sprintf(" (Line %d)", n.Line))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n#### Relationships\n")
	for _, e := range g.Edges {
		fromName := e.From
		toName := e.To
		// Try to find names for better readability
		for _, n := range g.Nodes {
			if n.ID == e.From {
				fromName = n.Name
			}
			if n.ID == e.To {
				toName = n.Name
			}
		}
		sb.WriteString(fmt.Sprintf("- %s --(%s)--> %s\n", fromName, e.Type, toName))
	}

	return sb.String()
}
