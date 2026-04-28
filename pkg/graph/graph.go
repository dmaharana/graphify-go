package graph

import (
	"encoding/json"
	"os"
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
