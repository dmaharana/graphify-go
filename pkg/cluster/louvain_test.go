package cluster

import (
	"graphify-go/pkg/graph"
	"testing"
)

func TestLouvainClustering(t *testing.T) {
	g := graph.NewGraph()
	// Cluster 1: Auth
	g.AddNode(graph.Node{ID: "auth.go", Type: graph.FileNode, Name: "auth.go", FilePath: "auth/auth.go"})
	g.AddNode(graph.Node{ID: "Login", Type: graph.FunctionNode, Name: "Login", FilePath: "auth/auth.go"})
	g.AddNode(graph.Node{ID: "Token", Type: graph.StructNode, Name: "Token", FilePath: "auth/auth.go"})
	g.AddEdge("Login", "Token", graph.EdgeReferences, graph.ConfidenceExtracted, 1, "auth/auth.go")

	// Cluster 2: Database
	g.AddNode(graph.Node{ID: "db.go", Type: graph.FileNode, Name: "db.go", FilePath: "db/db.go"})
	g.AddNode(graph.Node{ID: "Connect", Type: graph.FunctionNode, Name: "Connect", FilePath: "db/db.go"})
	g.AddNode(graph.Node{ID: "Query", Type: graph.FunctionNode, Name: "Query", FilePath: "db/db.go"})
	g.AddEdge("Connect", "Query", graph.EdgeCalls, graph.ConfidenceExtracted, 1, "db/db.go")

	// Weak cross-link: Login calls Query
	g.AddEdge("Login", "Query", graph.EdgeCalls, graph.ConfidenceInferred, 1, "auth/auth.go")

	detector := NewDetector(g)
	communities := detector.Detect()

	if len(communities) < 2 {
		t.Fatalf("expected at least 2 communities, got %d", len(communities))
	}

	// Login and Token should share a community
	loginNode := g.GetNode("Login")
	tokenNode := g.GetNode("Token")
	if loginNode.Community != tokenNode.Community {
		t.Errorf("expected Login and Token to share community, got %d and %d", loginNode.Community, tokenNode.Community)
	}

	// Connect and Query should share a community
	connectNode := g.GetNode("Connect")
	queryNode := g.GetNode("Query")
	if connectNode.Community != queryNode.Community {
		t.Errorf("expected Connect and Query to share community, got %d and %d", connectNode.Community, queryNode.Community)
	}
}
