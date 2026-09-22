package cmd

import (
	"fmt"
	"graphify-go/pkg/graph"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var pathGraphPath string

var pathCmd = &cobra.Command{
	Use:   "path [source] [target]",
	Short: "Find the shortest architectural path between two symbols",
	Long: `Find the shortest call, dependency, or containment path between two symbols.
Uses breadth-first search (BFS) over the architecture knowledge graph to reveal how
two concepts interact across file, module, or community boundaries.`,
	Example: `  # Find path between two symbols by name
  graphify-go path "ScanWithManifest" "NewResolver"

  # Find path using node IDs
  graphify-go path "cmd/scan.go:scanCmd" "pkg/scanner/manifest.go:Manifest"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		from := args[0]
		to := args[1]
		targetPath := resolveGraphPath(pathGraphPath)

		g, err := graph.LoadGraph(targetPath)
		if err != nil {
			fmt.Printf("Error reading graph from %s: %v\n", targetPath, err)
			os.Exit(1)
		}

		path, err := g.Path(from, to)
		if err != nil {
			// Try finding by symbol names if direct IDs failed
			foundFrom := from
			foundTo := to
			for _, n := range g.Nodes {
				if n.Name == from {
					foundFrom = n.ID
				}
				if n.Name == to {
					foundTo = n.ID
				}
			}
			path, err = g.Path(foundFrom, foundTo)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Shortest path (%d hops):\n", len(path)-1)
		fmt.Printf("  %s\n", strings.Join(path, " <--> "))
	},
}

func init() {
	pathCmd.Flags().StringVarP(&pathGraphPath, "graph", "g", "", "Path to the graph JSON file (defaults to graphify-out/graph.json)")
	rootCmd.AddCommand(pathCmd)
}
