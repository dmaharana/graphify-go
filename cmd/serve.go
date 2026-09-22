package cmd

import (
	"fmt"
	"graphify-go/pkg/graph"
	"graphify-go/pkg/mcp"
	"os"

	"github.com/spf13/cobra"
)

var serveGraphPath string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a stdio JSON-RPC 2.0 MCP server for AI coding assistants",
	Long: `Start a Model Context Protocol (MCP) server over standard I/O (stdio).
Exposes structured tools so AI coding assistants (Claude Code, Gemini CLI, Cursor, Codex)
can query and traverse the codebase architecture graph directly via tool calls.

Exposed MCP Tools:
  - query_graph:    Plain-language architectural question returning Markdown subgraphs.
  - get_node:       Fetch properties and source line for a specific node ID.
  - get_neighbors:  Fetch incoming callers and outgoing dependencies.
  - shortest_path:  Compute the shortest path between two symbols.
  - get_impact:     Calculate the transitive blast radius of upstream dependents.`,
	Example: `  # Start MCP server with default graphify-out/graph.json
  graphify-go serve

  # Start MCP server with custom graph file
  graphify-go serve --graph /path/to/graph.json`,
	Run: func(cmd *cobra.Command, args []string) {
		targetPath := resolveGraphPath(serveGraphPath)

		g, err := graph.LoadGraph(targetPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading graph from %s: %v\n", targetPath, err)
			os.Exit(1)
		}

		server := mcp.NewServer(g)
		if err := server.ServeStdio(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server exited with error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().StringVarP(&serveGraphPath, "graph", "g", "", "Path to the graph JSON file (defaults to graphify-out/graph.json)")
	rootCmd.AddCommand(serveCmd)
}
