package report

import (
	"graphify-go/pkg/graph"
	"strings"
	"testing"
)

func TestGenerateReport(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "Auth", Type: graph.StructNode, Name: "Auth", FilePath: "auth/auth.go", Community: 1})
	g.AddNode(graph.Node{ID: "DB", Type: graph.StructNode, Name: "DB", FilePath: "db/db.go", Community: 2})
	g.AddEdge("Auth", "DB", graph.EdgeCalls, graph.ConfidenceInferred, 10, "auth/auth.go")

	report := GenerateReport(g, "test-project")
	if !strings.Contains(report, "test-project") {
		t.Errorf("expected report to contain project name")
	}
	if !strings.Contains(report, "God Nodes") {
		t.Errorf("expected report to contain God Nodes")
	}
	if !strings.Contains(report, "Communities") {
		t.Errorf("expected report to contain Communities")
	}
	if !strings.Contains(report, "Surprising Connections") {
		t.Errorf("expected report to contain Surprising Connections")
	}
}
