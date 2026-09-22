package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"graphify-go/pkg/graph"
	"io"
	"os"
	"strings"
)

type Server struct {
	graph *graph.Graph
}

func NewServer(g *graph.Graph) *Server {
	return &Server{graph: g}
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func (s *Server) ServeStdio() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		if err := s.HandleMessage(strings.NewReader(string(line)), os.Stdout); err != nil {
			// Log to stderr so stdout remains pure JSON-RPC
			fmt.Fprintf(os.Stderr, "mcp handler error: %v\n", err)
		}
	}
}

func (s *Server) HandleMessage(r io.Reader, w io.Writer) error {
	decoder := json.NewDecoder(r)
	var req JSONRPCRequest
	if err := decoder.Decode(&req); err != nil {
		return err
	}

	var resp JSONRPCResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "graphify-go",
				"version": "1.0.0",
			},
		}

	case "notifications/initialized":
		return nil

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "query_graph",
					"description": "Query the codebase architecture knowledge graph to extract a relevant subgraph",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"question": map[string]interface{}{
								"type":        "string",
								"description": "Plain language architectural question",
							},
						},
						"required": []string{"question"},
					},
				},
				{
					"name":        "get_node",
					"description": "Get detailed symbol properties and location",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"description": "Symbol node identifier",
							},
						},
						"required": []string{"id"},
					},
				},
				{
					"name":        "get_neighbors",
					"description": "Get incoming callers and outgoing targets for a symbol",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"description": "Symbol node identifier",
							},
						},
						"required": []string{"id"},
					},
				},
				{
					"name":        "shortest_path",
					"description": "Find the shortest architectural path between two code symbols",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"from": map[string]interface{}{
								"type":        "string",
								"description": "Source symbol ID",
							},
							"to": map[string]interface{}{
								"type":        "string",
								"description": "Target symbol ID",
							},
						},
						"required": []string{"from", "to"},
					},
				},
				{
					"name":        "get_impact",
					"description": "Calculate the transitive blast radius of upstream callers/dependents for a symbol or file",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"target": map[string]interface{}{
								"type":        "string",
								"description": "Target symbol ID or file path",
							},
						},
						"required": []string{"target"},
					},
				},
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			resp.Error = map[string]interface{}{"code": -32602, "message": "Invalid params"}
			break
		}

		resultText, err := s.executeTool(callParams.Name, callParams.Arguments)
		if err != nil {
			resp.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
				},
				"isError": true,
			}
		} else {
			resp.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": resultText},
				},
			}
		}

	default:
		resp.Error = map[string]interface{}{"code": -32601, "message": fmt.Sprintf("Method %s not found", req.Method)}
	}

	encoder := json.NewEncoder(w)
	return encoder.Encode(resp)
}

func (s *Server) executeTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "query_graph":
		q, _ := args["question"].(string)
		subgraph := s.graph.Query(q)
		return subgraph.ToMarkdown(), nil

	case "get_node":
		id, _ := args["id"].(string)
		node := s.graph.GetNode(id)
		if node == nil {
			return "", fmt.Errorf("node not found: %s", id)
		}
		bytes, _ := json.MarshalIndent(node, "", "  ")
		return string(bytes), nil

	case "get_neighbors":
		id, _ := args["id"].(string)
		exp, err := s.graph.Explain(id)
		if err != nil {
			return "", err
		}
		bytes, _ := json.MarshalIndent(exp, "", "  ")
		return string(bytes), nil

	case "shortest_path":
		from, _ := args["from"].(string)
		to, _ := args["to"].(string)
		path, err := s.graph.Path(from, to)
		if err != nil {
			return "", err
		}
		return strings.Join(path, " -> "), nil

	case "get_impact":
		target, _ := args["target"].(string)
		affected := s.graph.Affected(target)
		if len(affected) == 0 {
			return "No upstream dependents affected.", nil
		}
		return "Affected upstream symbols:\n- " + strings.Join(affected, "\n- "), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
