package cmd

import (
	"fmt"
	"graphify-go/pkg/graph"
	"os"

	"github.com/spf13/cobra"
)

var explainGraphPath string

var explainCmd = &cobra.Command{
	Use:   "explain [symbol]",
	Short: "Explain a code symbol, its community, and incoming/outgoing connections",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		symbol := args[0]
		targetPath := resolveGraphPath(explainGraphPath)

		g, err := graph.LoadGraph(targetPath)
		if err != nil {
			fmt.Printf("Error reading graph from %s: %v\n", targetPath, err)
			os.Exit(1)
		}

		exp, err := g.Explain(symbol)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Node: %s\n", exp.Node.Name)
		fmt.Printf("  ID:        %s\n", exp.Node.ID)
		fmt.Printf("  Type:      %s\n", exp.Node.Type)
		if exp.Node.FilePath != "" {
			fmt.Printf("  Source:    %s", exp.Node.FilePath)
			if exp.Node.Line > 0 {
				fmt.Printf(" L%d", exp.Node.Line)
			}
			fmt.Println()
		}
		if exp.Node.Community > 0 {
			fmt.Printf("  Community: %d\n", exp.Node.Community)
		}
		fmt.Printf("  Degree:    %d\n\n", exp.Degree)

		fmt.Printf("Connections (%d):\n", exp.Degree)
		for _, e := range exp.Outgoing {
			confTag := ""
			if e.Confidence != "" {
				confTag = fmt.Sprintf(" [%s]", e.Confidence)
			}
			fmt.Printf("  --> %s [%s]%s\n", e.To, e.Type, confTag)
		}
		for _, e := range exp.Incoming {
			confTag := ""
			if e.Confidence != "" {
				confTag = fmt.Sprintf(" [%s]", e.Confidence)
			}
			fmt.Printf("  <-- %s [%s]%s\n", e.From, e.Type, confTag)
		}
	},
}

func init() {
	explainCmd.Flags().StringVarP(&explainGraphPath, "graph", "g", "", "Path to the graph JSON file (defaults to graphify-out/graph.json)")
	rootCmd.AddCommand(explainCmd)
}
