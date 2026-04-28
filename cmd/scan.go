package cmd

import (
	"fmt"
	"graphify-go/pkg/scanner"
	"os"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan a directory and generate a graph",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		fmt.Printf("Scanning directory: %s\n", dir)
		
		s := scanner.NewScanner(dir)
		graph, err := s.Scan()
		if err != nil {
			fmt.Printf("Error scanning: %v\n", err)
			os.Exit(1)
		}

		// For now, just print the number of nodes
		fmt.Printf("Scan complete. Found %d nodes.\n", len(graph.Nodes))
		
		// Write to graph.json
		if err := graph.Save("graph.json"); err != nil {
			fmt.Printf("Error saving graph: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Graph saved to graph.json")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
