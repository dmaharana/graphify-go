package cmd

import (
	"encoding/json"
	"fmt"
	"graphify-go/pkg/graph"
	"os"

	"github.com/spf13/cobra"
)

var graphPath string

var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Query the graph to extract a relevant subgraph",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := args[0]
		
		data, err := os.ReadFile(graphPath)
		if err != nil {
			fmt.Printf("Error reading graph: %v\n", err)
			os.Exit(1)
		}

		var g graph.Graph
		if err := json.Unmarshal(data, &g); err != nil {
			fmt.Printf("Error parsing graph: %v\n", err)
			os.Exit(1)
		}

		subgraph := g.Query(question)
		
		fmt.Println(subgraph.ToMarkdown())
	},
}

func init() {
	queryCmd.Flags().StringVarP(&graphPath, "graph", "g", "graph.json", "Path to the graph JSON file")
	rootCmd.AddCommand(queryCmd)
}
