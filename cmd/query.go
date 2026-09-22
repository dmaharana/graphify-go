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
	Args:  cobra.ExactArgs(1),
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
