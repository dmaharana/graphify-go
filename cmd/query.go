package cmd

import (
	"fmt"
	"graphify-go/pkg/graph"
	"os"

	"github.com/spf13/cobra"
)

var queryGraphPath string

func resolveGraphPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if _, err := os.Stat("graphify-out/graph.json"); err == nil {
		return "graphify-out/graph.json"
	}
	return "graph.json"
}

var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Query the graph to extract a relevant subgraph",
	Long: `Query the architecture graph using plain-language keywords.
It extracts a focused 1-hop subgraph around relevant seed concepts and formats the output
in clean GitHub-flavored Markdown, optimized for AI assistant context windows.`,
	Example: `  # Query using default graphify-out/graph.json
  graphify-go query "how does the scanner handle ignores?"

  # Query an explicit graph JSON file
  graphify-go query "database connection pooling" --graph /path/to/graph.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := args[0]
		targetPath := resolveGraphPath(queryGraphPath)

		g, err := graph.LoadGraph(targetPath)
		if err != nil {
			fmt.Printf("Error reading graph from %s: %v\n", targetPath, err)
			os.Exit(1)
		}

		subgraph := g.Query(question)
		fmt.Println(subgraph.ToMarkdown())
	},
}

func init() {
	queryCmd.Flags().StringVarP(&queryGraphPath, "graph", "g", "", "Path to the graph JSON file (defaults to graphify-out/graph.json)")
	rootCmd.AddCommand(queryCmd)
}
