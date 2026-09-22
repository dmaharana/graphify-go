package cmd

import (
	"fmt"
	"graphify-go/pkg/graph"
	"os"

	"github.com/spf13/cobra"
)

var affectedGraphPath string

var affectedCmd = &cobra.Command{
	Use:   "affected [file|symbol]",
	Short: "Determine the blast radius of upstream callers/dependents for a file or symbol",
	Long: `Determine the transitive blast radius for a given file or code symbol.
Performs reverse-dependency graph traversal across calls, references, embeds, and imports
to reveal every upstream symbol that could be impacted if the target file or symbol is modified.`,
	Example: `  # Blast radius for a source file
  graphify-go affected "pkg/scanner/scanner.go"

  # Blast radius for a function or method symbol
  graphify-go affected "NewScanner"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		targetPath := resolveGraphPath(affectedGraphPath)

		g, err := graph.LoadGraph(targetPath)
		if err != nil {
			fmt.Printf("Error reading graph from %s: %v\n", targetPath, err)
			os.Exit(1)
		}

		affected := g.Affected(target)
		if len(affected) == 0 {
			// Try finding node ID matching the symbol name
			for _, n := range g.Nodes {
				if n.Name == target {
					affected = g.Affected(n.ID)
					break
				}
			}
		}

		if len(affected) == 0 {
			fmt.Printf("No upstream dependents affected by '%s'.\n", target)
			return
		}

		fmt.Printf("Blast radius for '%s' (%d affected upstream symbols):\n", target, len(affected))
		for _, id := range affected {
			fmt.Printf("  - %s\n", id)
		}
	},
}

func init() {
	affectedCmd.Flags().StringVarP(&affectedGraphPath, "graph", "g", "", "Path to the graph JSON file (defaults to graphify-out/graph.json)")
	rootCmd.AddCommand(affectedCmd)
}
