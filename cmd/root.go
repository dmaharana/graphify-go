package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "graphify-go",
	Short: "Graphify-Go: High-performance code knowledge graph engine",
	Long: `Graphify-Go is a high-performance, deterministic codebase knowledge graph engine.
It uses Tree-Sitter AST parsing and pure Go graph algorithms to extract code architecture,
cross-file calls, type relationships, and community clusters without external LLM dependencies.

All generated artifacts are written to a dedicated output directory (defaults to 'graphify-out/'):
  - graph.json:       Full structured knowledge graph
  - GRAPH_REPORT.md:  Architecture audit report (God nodes, subsystems, cross-boundary links)
  - graph.html:       Self-contained interactive force-directed visualizer
  - manifest.json:    Incremental file cache tracking mtimes and SHA-256 hashes`,
	Example: `  # Scan codebase and generate all outputs in graphify-out/
  graphify-go scan .

  # Scan multiple directories incrementally
  graphify-go scan ./cmd ./pkg --update

  # Query architecture for context
  graphify-go query "how does the scanner handle ignores?"

  # Inspect a symbol, its community, and connections
  graphify-go explain "Scanner"

  # Trace shortest dependency path between two symbols
  graphify-go path "ScanWithManifest" "NewResolver"

  # Find blast radius (upstream callers/dependents)
  graphify-go affected "pkg/scanner/scanner.go"

  # Start stdio MCP server for AI coding assistants
  graphify-go serve`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
}
