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
