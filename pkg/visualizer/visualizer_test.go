package visualizer

import (
	"fmt"
	"graphify-go/pkg/graph"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPruneGraph_NoPruneNeeded(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "n1", Type: graph.FunctionNode, Name: "n1", Community: 1})
	g.AddNode(graph.Node{ID: "n2", Type: graph.FunctionNode, Name: "n2", Community: 2})
	g.AddEdge("n1", "n2", graph.EdgeCalls, graph.ConfidenceExtracted, 1, "test.go")

	res := PruneGraph(g, 10)
	if res.IsPruned {
		t.Errorf("expected IsPruned to be false for small graph")
	}
	if len(res.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(res.Nodes))
	}
	if len(res.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(res.Edges))
	}
}

func TestPruneGraph_RetainsCommunityRepresentatives(t *testing.T) {
	g := graph.NewGraph()

	// Community 1: 5 nodes, node1_hub has highest degree
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("c1_node%d", i)
		g.AddNode(graph.Node{ID: id, Type: graph.FunctionNode, Name: id, Community: 1})
	}
	for i := 1; i < 5; i++ {
		g.AddEdge("c1_node0", fmt.Sprintf("c1_node%d", i), graph.EdgeCalls, graph.ConfidenceExtracted, 1, "c1.go")
	}

	// Community 2: 5 nodes, node2_hub has highest degree
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("c2_node%d", i)
		g.AddNode(graph.Node{ID: id, Type: graph.FunctionNode, Name: id, Community: 2})
	}
	for i := 1; i < 5; i++ {
		g.AddEdge("c2_node0", fmt.Sprintf("c2_node%d", i), graph.EdgeCalls, graph.ConfidenceExtracted, 1, "c2.go")
	}

	// Community 3: 5 nodes, node3_hub has highest degree
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("c3_node%d", i)
		g.AddNode(graph.Node{ID: id, Type: graph.FunctionNode, Name: id, Community: 3})
	}
	for i := 1; i < 5; i++ {
		g.AddEdge("c3_node0", fmt.Sprintf("c3_node%d", i), graph.EdgeCalls, graph.ConfidenceExtracted, 1, "c3.go")
	}

	// Total 15 nodes, prune to 5 nodes
	res := PruneGraph(g, 5)
	if !res.IsPruned {
		t.Fatalf("expected IsPruned to be true")
	}
	if len(res.Nodes) != 5 {
		t.Fatalf("expected 5 pruned nodes, got %d", len(res.Nodes))
	}

	// Must contain representatives from each of communities 1, 2, 3
	foundComm := make(map[int]bool)
	foundIDs := make(map[string]bool)
	for _, n := range res.Nodes {
		foundComm[n.Community] = true
		foundIDs[n.ID] = true
	}

	for cid := 1; cid <= 3; cid++ {
		if !foundComm[cid] {
			t.Errorf("expected representative for community %d", cid)
		}
	}

	// Community hubs must be in the pruned graph
	if !foundIDs["c1_node0"] || !foundIDs["c2_node0"] || !foundIDs["c3_node0"] {
		t.Errorf("expected community hubs to be retained: %v", foundIDs)
	}
}

func TestSeedNodePositions(t *testing.T) {
	nodes := []VisNode{
		{ID: "n1", Community: 1, Degree: 10},
		{ID: "n2", Community: 1, Degree: 5},
		{ID: "n3", Community: 1, Degree: 2},
		{ID: "n4", Community: 2, Degree: 8},
		{ID: "n5", Community: 2, Degree: 3},
	}

	SeedNodePositions(nodes)

	// Distinct positions
	posMap := make(map[string]bool)
	for _, n := range nodes {
		posKey := fmt.Sprintf("%.1f,%.1f", n.X, n.Y)
		if posMap[posKey] {
			t.Errorf("duplicate node position detected at %s for node %s", posKey, n.ID)
		}
		posMap[posKey] = true
	}
}

func TestShapeForNodeType(t *testing.T) {
	tests := []struct {
		nodeType graph.NodeType
		expected string
	}{
		{graph.FunctionNode, "dot"},
		{graph.MethodNode, "diamond"},
		{graph.StructNode, "square"},
		{graph.InterfaceNode, "triangle"},
		{graph.PackageNode, "star"},
		{graph.FileNode, "hexagon"},
	}

	for _, tc := range tests {
		got := ShapeForNodeType(tc.nodeType)
		if got != tc.expected {
			t.Errorf("ShapeForNodeType(%s) = %s; want %s", tc.nodeType, got, tc.expected)
		}
	}
}

func TestGenerateHTML(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "main.go", Type: graph.FileNode, Name: "main.go", Community: 1})
	g.AddNode(graph.Node{ID: "main", Type: graph.FunctionNode, Name: "main", Community: 1})
	g.AddNode(graph.Node{ID: "Service", Type: graph.StructNode, Name: "Service", Community: 1})
	g.AddEdge("main.go", "main", graph.EdgeContains, graph.ConfidenceExtracted, 1, "main.go")
	g.AddEdge("main", "Service", graph.EdgeReferences, graph.ConfidenceInferred, 10, "main.go")

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "graph.html")

	cfg := Config{
		MaxNodes: 500,
		Layout:   "force",
		Theme:    "dark",
	}

	if err := GenerateHTMLWithConfig(g, outPath, cfg); err != nil {
		t.Fatalf("GenerateHTMLWithConfig failed: %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read generated HTML: %v", err)
	}

	htmlStr := string(content)
	if !strings.Contains(htmlStr, "<!DOCTYPE html>") {
		t.Errorf("expected HTML doctype")
	}
	if !strings.Contains(htmlStr, "main.go") {
		t.Errorf("expected graph data with main.go in HTML")
	}
	if !strings.Contains(htmlStr, "Graphify-Go Knowledge Graph") {
		t.Errorf("expected title in HTML")
	}
	if !strings.Contains(htmlStr, "meta-data") {
		t.Errorf("expected metadata script tag in HTML")
	}
	if !strings.Contains(htmlStr, "square") {
		t.Errorf("expected struct shape square in HTML data")
	}
}
