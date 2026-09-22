package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Confidence string

const (
	ConfidenceExtracted Confidence = "EXTRACTED"
	ConfidenceInferred  Confidence = "INFERRED"
	ConfidenceAmbiguous Confidence = "AMBIGUOUS"
)

type NodeType string

const (
	FileNode      NodeType = "file"
	PackageNode   NodeType = "package"
	FunctionNode  NodeType = "function"
	MethodNode    NodeType = "method"
	StructNode    NodeType = "struct"
	InterfaceNode NodeType = "interface"
)

type EdgeType string

const (
	EdgeContains   EdgeType = "contains"
	EdgeCalls      EdgeType = "calls"
	EdgeMethod     EdgeType = "method"
	EdgeReferences EdgeType = "references"
	EdgeEmbeds     EdgeType = "embeds"
	EdgeImports    EdgeType = "imports"
	EdgeImplements EdgeType = "implements"
)

type Node struct {
	ID        string                 `json:"id"`
	Type      NodeType               `json:"type"`
	Name      string                 `json:"name"`
	FilePath  string                 `json:"file_path,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Community int                    `json:"community,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type Edge struct {
	From       string     `json:"from"`
	To         string     `json:"to"`
	Type       EdgeType   `json:"type"`
	Confidence Confidence `json:"confidence,omitempty"`
	Line       int        `json:"line,omitempty"`
	SourceFile string     `json:"source_file,omitempty"`
	Weight     float64    `json:"weight,omitempty"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`

	nodeIndex map[string]*Node `json:"-"`
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:     []Node{},
		Edges:     []Edge{},
		nodeIndex: make(map[string]*Node),
	}
}

func (g *Graph) buildIndex() {
	if g.nodeIndex == nil || len(g.nodeIndex) != len(g.Nodes) {
		g.nodeIndex = make(map[string]*Node, len(g.Nodes))
		for i := range g.Nodes {
			g.nodeIndex[g.Nodes[i].ID] = &g.Nodes[i]
		}
	}
}

func (g *Graph) AddNode(n Node) {
	g.Nodes = append(g.Nodes, n)
	if g.nodeIndex != nil {
		g.nodeIndex[n.ID] = &g.Nodes[len(g.Nodes)-1]
	}
}

func (g *Graph) AddEdge(from, to string, edgeType EdgeType, confidence Confidence, line int, sourceFile string) {
	if confidence == "" {
		confidence = ConfidenceExtracted
	}
	g.Edges = append(g.Edges, Edge{
		From:       from,
		To:         to,
		Type:       edgeType,
		Confidence: confidence,
		Line:       line,
		SourceFile: sourceFile,
		Weight:     1.0,
	})
}

func (g *Graph) AddSimpleEdge(from, to string, edgeType EdgeType) {
	g.AddEdge(from, to, edgeType, ConfidenceExtracted, 0, "")
}

func (g *Graph) GetNode(id string) *Node {
	g.buildIndex()
	return g.nodeIndex[id]
}

func (g *Graph) Save(filename string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func LoadGraph(filename string) (*Graph, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	g.buildIndex()
	return &g, nil
}

// Path finds the shortest path between two nodes using BFS (undirected graph traversal).
func (g *Graph) Path(from, to string) ([]string, error) {
	if g.GetNode(from) == nil {
		return nil, fmt.Errorf("source node not found: %s", from)
	}
	if g.GetNode(to) == nil {
		return nil, fmt.Errorf("target node not found: %s", to)
	}
	if from == to {
		return []string{from}, nil
	}

	// Build adjacency list (undirected for conceptual paths)
	adj := make(map[string][]string)
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
		adj[e.To] = append(adj[e.To], e.From)
	}

	queue := []string{from}
	parent := make(map[string]string)
	visited := make(map[string]bool)
	visited[from] = true

	found := false
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == to {
			found = true
			break
		}

		for _, nbr := range adj[curr] {
			if !visited[nbr] {
				visited[nbr] = true
				parent[nbr] = curr
				queue = append(queue, nbr)
			}
		}
	}

	if !found {
		return nil, errors.New("no path found between symbols")
	}

	// Reconstruct path
	var path []string
	curr := to
	for curr != "" {
		path = append([]string{curr}, path...)
		curr = parent[curr]
	}

	return path, nil
}

// Affected returns the transitive upstream blast radius for a node (who calls or references it).
func (g *Graph) Affected(target string) []string {
	// Incoming edges where relation is calls, references, embeds, imports
	incomingAdj := make(map[string][]string)
	for _, e := range g.Edges {
		// Only traverse dependencies (skip contains)
		if e.Type != EdgeContains {
			incomingAdj[e.To] = append(incomingAdj[e.To], e.From)
		}
	}

	visited := make(map[string]bool)
	var queue []string

	// Target can be a node ID or match a source file path
	matched := false
	if g.GetNode(target) != nil {
		queue = append(queue, target)
		matched = true
	} else {
		// Check if target is a file path
		for _, n := range g.Nodes {
			if n.FilePath == target || strings.HasSuffix(n.FilePath, target) {
				queue = append(queue, n.ID)
				matched = true
			}
		}
	}

	if !matched {
		return nil
	}

	var affected []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, caller := range incomingAdj[curr] {
			if !visited[caller] {
				visited[caller] = true
				affected = append(affected, caller)
				queue = append(queue, caller)
			}
		}
	}

	return affected
}

type SymbolExplanation struct {
	Node     Node   `json:"node"`
	Degree   int    `json:"degree"`
	Incoming []Edge `json:"incoming"`
	Outgoing []Edge `json:"outgoing"`
}

func (g *Graph) Explain(symbol string) (*SymbolExplanation, error) {
	node := g.GetNode(symbol)
	if node == nil {
		// Try case-insensitive search or suffix match
		lower := strings.ToLower(symbol)
		for _, n := range g.Nodes {
			if strings.ToLower(n.Name) == lower || strings.HasSuffix(strings.ToLower(n.ID), ":"+lower) {
				node = &n
				break
			}
		}
	}

	if node == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	var incoming []Edge
	var outgoing []Edge

	for _, e := range g.Edges {
		if e.To == node.ID {
			incoming = append(incoming, e)
		}
		if e.From == node.ID {
			outgoing = append(outgoing, e)
		}
	}

	return &SymbolExplanation{
		Node:     *node,
		Degree:   len(incoming) + len(outgoing),
		Incoming: incoming,
		Outgoing: outgoing,
	}, nil
}

type NodeCentrality struct {
	Node   Node `json:"node"`
	Degree int  `json:"degree"`
}

func (g *Graph) GodNodes(topN int) []NodeCentrality {
	degreeMap := make(map[string]int)
	for _, e := range g.Edges {
		// Ignore hierarchical file-contains edges in centrality calculations
		if e.Type != EdgeContains {
			degreeMap[e.From]++
			degreeMap[e.To]++
		}
	}

	var centralities []NodeCentrality
	for _, n := range g.Nodes {
		// Skip file nodes
		if n.Type != FileNode {
			centralities = append(centralities, NodeCentrality{
				Node:   n,
				Degree: degreeMap[n.ID],
			})
		}
	}

	sort.Slice(centralities, func(i, j int) bool {
		return centralities[i].Degree > centralities[j].Degree
	})

	if len(centralities) > topN {
		return centralities[:topN]
	}
	return centralities
}

func (g *Graph) Query(question string) *Graph {
	keywords := strings.Fields(strings.ToLower(question))
	seedNodes := make(map[string]bool)

	for _, node := range g.Nodes {
		nodeLower := strings.ToLower(node.Name)
		for _, kw := range keywords {
			if len(kw) > 2 && strings.Contains(nodeLower, kw) {
				seedNodes[node.ID] = true
				break
			}
		}
	}

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
		sb.WriteString(fmt.Sprintf("- **%s** (`%s`)", n.Name, n.Type))
		if n.FilePath != "" {
			sb.WriteString(fmt.Sprintf(" - `%s`", n.FilePath))
			if n.Line > 0 {
				sb.WriteString(fmt.Sprintf(" (Line %d)", n.Line))
			}
		}
		if n.Community > 0 {
			sb.WriteString(fmt.Sprintf(" [Community %d]", n.Community))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n#### Relationships\n")
	for _, e := range g.Edges {
		confTag := ""
		if e.Confidence != "" {
			confTag = fmt.Sprintf(" [%s]", e.Confidence)
		}
		sb.WriteString(fmt.Sprintf("- `%s` --(%s%s)--> `%s`\n", e.From, e.Type, confTag, e.To))
	}

	return sb.String()
}
