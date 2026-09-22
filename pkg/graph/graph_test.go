package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func buildSampleGraph() *Graph {
	g := NewGraph()
	// Nodes
	g.AddNode(Node{ID: "main.go", Type: FileNode, Name: "main.go", FilePath: "main.go"})
	g.AddNode(Node{ID: "main", Type: FunctionNode, Name: "main", FilePath: "main.go", Line: 5, Community: 1})
	g.AddNode(Node{ID: "service.go", Type: FileNode, Name: "service.go", FilePath: "service.go"})
	g.AddNode(Node{ID: "Service", Type: StructNode, Name: "Service", FilePath: "service.go", Line: 10, Community: 1})
	g.AddNode(Node{ID: "DoWork", Type: FunctionNode, Name: "DoWork", FilePath: "service.go", Line: 15, Community: 1})
	g.AddNode(Node{ID: "db.go", Type: FileNode, Name: "db.go", FilePath: "db.go"})
	g.AddNode(Node{ID: "QueryDB", Type: FunctionNode, Name: "QueryDB", FilePath: "db.go", Line: 20, Community: 2})

	// Edges
	g.AddEdge("main.go", "main", EdgeContains, ConfidenceExtracted, 5, "main.go")
	g.AddEdge("main", "DoWork", EdgeCalls, ConfidenceInferred, 6, "main.go")
	g.AddEdge("DoWork", "QueryDB", EdgeCalls, ConfidenceInferred, 16, "service.go")
	g.AddEdge("service.go", "Service", EdgeContains, ConfidenceExtracted, 10, "service.go")
	g.AddEdge("service.go", "DoWork", EdgeContains, ConfidenceExtracted, 15, "service.go")
	g.AddEdge("db.go", "QueryDB", EdgeContains, ConfidenceExtracted, 20, "db.go")

	return g
}

func TestShortestPath(t *testing.T) {
	g := buildSampleGraph()

	path, err := g.Path("main", "QueryDB")
	if err != nil {
		t.Fatalf("expected path, got error: %v", err)
	}

	expected := []string{"main", "DoWork", "QueryDB"}
	if len(path) != len(expected) {
		t.Fatalf("expected path length %d, got %d: %v", len(expected), len(path), path)
	}
	for i, node := range path {
		if node != expected[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expected[i], node)
		}
	}
}

func TestAffected(t *testing.T) {
	g := buildSampleGraph()

	// If QueryDB changes, DoWork and main are affected
	affected := g.Affected("QueryDB")
	affectedSet := make(map[string]bool)
	for _, id := range affected {
		affectedSet[id] = true
	}

	if !affectedSet["DoWork"] {
		t.Errorf("expected DoWork to be affected by QueryDB")
	}
	if !affectedSet["main"] {
		t.Errorf("expected main to be affected by QueryDB")
	}
}

func TestExplain(t *testing.T) {
	g := buildSampleGraph()

	exp, err := g.Explain("DoWork")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exp.Node.ID != "DoWork" {
		t.Errorf("expected node ID DoWork, got %s", exp.Node.ID)
	}
	if len(exp.Incoming) == 0 {
		t.Errorf("expected incoming edges for DoWork")
	}
	if len(exp.Outgoing) == 0 {
		t.Errorf("expected outgoing edges for DoWork")
	}
}

func TestGodNodes(t *testing.T) {
	g := buildSampleGraph()

	gods := g.GodNodes(3)
	if len(gods) == 0 {
		t.Fatalf("expected god nodes, got none")
	}
	// DoWork has 3 edges (contains from service.go, calls from main, calls to QueryDB)
	if gods[0].Node.ID != "DoWork" && gods[0].Degree < 2 {
		t.Errorf("expected high degree node at top, got %s (degree %d)", gods[0].Node.ID, gods[0].Degree)
	}
}

func TestSaveAndLoad(t *testing.T) {
	g := buildSampleGraph()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "graph.json")

	if err := g.Save(path); err != nil {
		t.Fatalf("failed to save graph: %v", err)
	}

	loaded, err := LoadGraph(path)
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	if len(loaded.Nodes) != len(g.Nodes) {
		t.Errorf("expected %d nodes, got %d", len(g.Nodes), len(loaded.Nodes))
	}
	if len(loaded.Edges) != len(g.Edges) {
		t.Errorf("expected %d edges, got %d", len(g.Edges), len(loaded.Edges))
	}
	_ = os.Remove(path)
}
