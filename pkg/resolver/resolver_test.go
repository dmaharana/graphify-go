package resolver

import (
	"graphify-go/pkg/graph"
	"graphify-go/pkg/parser"
	"testing"
)

func TestResolveCrossFileCalls(t *testing.T) {
	// Simulate File 1: pkg/scanner/scanner.go in package "scanner"
	file1Nodes := []graph.Node{
		{ID: "pkg/scanner/scanner.go", Type: graph.FileNode, Name: "scanner.go"},
		{ID: "pkg/scanner/scanner.go:Scanner", Type: graph.StructNode, Name: "Scanner"},
		{ID: "pkg/scanner/scanner.go:NewScanner", Type: graph.FunctionNode, Name: "NewScanner"},
		{ID: "pkg/scanner/scanner.go:Scanner.Scan", Type: graph.MethodNode, Name: "Scan"},
	}
	res1 := &parser.ParseResult{
		PackageName: "scanner",
		Nodes:       file1Nodes,
		Edges: []graph.Edge{
			{From: "pkg/scanner/scanner.go", To: "pkg/scanner/scanner.go:Scanner", Type: graph.EdgeContains},
			{From: "pkg/scanner/scanner.go", To: "pkg/scanner/scanner.go:NewScanner", Type: graph.EdgeContains},
			{From: "pkg/scanner/scanner.go:Scanner", To: "pkg/scanner/scanner.go:Scanner.Scan", Type: graph.EdgeMethod},
		},
	}

	// Simulate File 2: cmd/scan.go in package "cmd"
	file2Nodes := []graph.Node{
		{ID: "cmd/scan.go", Type: graph.FileNode, Name: "scan.go"},
		{ID: "cmd/scan.go:runScan", Type: graph.FunctionNode, Name: "runScan"},
	}
	res2 := &parser.ParseResult{
		PackageName: "cmd",
		Nodes:       file2Nodes,
		Edges: []graph.Edge{
			{From: "cmd/scan.go", To: "cmd/scan.go:runScan", Type: graph.EdgeContains},
		},
		Imports: map[string]string{
			"scanner": "graphify-go/pkg/scanner",
		},
		RawCalls: []parser.RawCall{
			{
				CallerID:   "cmd/scan.go:runScan",
				CalleeName: "NewScanner",
				Receiver:   "scanner",
				ImportPath: "graphify-go/pkg/scanner",
				IsMember:   false,
				Line:       23,
				SourceFile: "cmd/scan.go",
			},
		},
	}

	r := NewResolver()
	r.AddParseResult(res1)
	r.AddParseResult(res2)

	g := r.Resolve()

	// Verify that an INFERRED calls edge was created from runScan to NewScanner
	found := false
	for _, e := range g.Edges {
		if e.From == "cmd/scan.go:runScan" && e.To == "pkg/scanner/scanner.go:NewScanner" && e.Type == graph.EdgeCalls {
			if e.Confidence != graph.ConfidenceInferred {
				t.Errorf("expected ConfidenceInferred, got %s", e.Confidence)
			}
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected cross-file calls edge from cmd/scan.go:runScan to pkg/scanner/scanner.go:NewScanner, none found")
	}
}
