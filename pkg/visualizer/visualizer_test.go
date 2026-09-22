package visualizer

import (
	"graphify-go/pkg/graph"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHTML(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "main.go", Type: graph.FileNode, Name: "main.go", Community: 1})
	g.AddNode(graph.Node{ID: "main", Type: graph.FunctionNode, Name: "main", Community: 1})
	g.AddEdge("main.go", "main", graph.EdgeContains, graph.ConfidenceExtracted, 1, "main.go")

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "graph.html")

	if err := GenerateHTML(g, outPath); err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
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
}
