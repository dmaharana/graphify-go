package mcp

import (
	"bytes"
	"encoding/json"
	"graphify-go/pkg/graph"
	"strings"
	"testing"
)

func TestMCPServerToolsListAndCall(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "main", Type: graph.FunctionNode, Name: "main", FilePath: "main.go"})
	g.AddNode(graph.Node{ID: "run", Type: graph.FunctionNode, Name: "run", FilePath: "run.go"})
	g.AddEdge("main", "run", graph.EdgeCalls, graph.ConfidenceExtracted, 5, "main.go")

	server := NewServer(g)

	// 1. Test initialize
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}` + "\n"
	in := bytes.NewBufferString(initReq)
	out := &bytes.Buffer{}
	if err := server.HandleMessage(in, out); err != nil {
		t.Fatal(err)
	}

	var initResp map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &initResp); err != nil {
		t.Fatalf("failed to parse init resp: %v", err)
	}
	if initResp["error"] != nil {
		t.Fatalf("unexpected error in init: %v", initResp["error"])
	}

	// 2. Test tools/list
	listReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	in = bytes.NewBufferString(listReq)
	out = &bytes.Buffer{}
	if err := server.HandleMessage(in, out); err != nil {
		t.Fatal(err)
	}

	var listResp map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse list resp: %v", err)
	}
	result := listResp["result"].(map[string]interface{})
	tools := result["tools"].([]interface{})
	if len(tools) < 4 {
		t.Errorf("expected at least 4 tools, got %d", len(tools))
	}

	// 3. Test tools/call: shortest_path
	callReq := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"shortest_path","arguments":{"from":"main","to":"run"}}}` + "\n"
	in = bytes.NewBufferString(callReq)
	out = &bytes.Buffer{}
	if err := server.HandleMessage(in, out); err != nil {
		t.Fatal(err)
	}

	var callResp map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &callResp); err != nil {
		t.Fatalf("failed to parse call resp: %v", err)
	}
	callResult := callResp["result"].(map[string]interface{})
	content := callResult["content"].([]interface{})
	first := content[0].(map[string]interface{})
	text := first["text"].(string)
	if !strings.Contains(text, "main -> run") {
		t.Errorf("expected path result containing 'main -> run', got: %s", text)
	}
}
