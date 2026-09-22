package cmd

import (
	"fmt"
	"graphify-go/pkg/cluster"
	"graphify-go/pkg/report"
	"graphify-go/pkg/scanner"
	"graphify-go/pkg/visualizer"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	outDir     string
	isUpdate   bool
	skipViz    bool
	skipReport bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [directories...]",
	Short: "Scan directories and generate an architecture knowledge graph",
	Long: `Scan one or more source directories to build a comprehensive codebase knowledge graph.

The scanner performs:
  1. Multi-root directory traversal respecting .gitignore and .graphifyignore rules.
  2. AST extraction of functions, method receivers, struct embeds, and imports.
  3. Predeclared builtins filtering (append, make, len, etc.) to prevent phantom hubs.
  4. Cross-file symbol resolution wiring call expressions across packages (INFERRED).
  5. Louvain community modularity clustering to identify architectural subsystems.
  6. Artifact generation (graph.json, GRAPH_REPORT.md, graph.html, and manifest.json).`,
	Example: `  # Scan current project (outputs to graphify-out/)
  graphify-go scan .

  # Scan multiple directories into custom output folder
  graphify-go scan ./cmd ./pkg -o my-graph-out

  # Incremental scan (only re-parses modified/added files)
  graphify-go scan . --update

  # Scan without generating visualization or report (JSON + manifest only)
  graphify-go scan . --no-viz --no-report`,
	Run: func(cmd *cobra.Command, args []string) {
		dirs := []string{"."}
		if len(args) > 0 {
			dirs = args
		}

		fmt.Printf("Scanning directories: %v\n", dirs)

		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		s := scanner.NewScanner(dirs...)
		manifestPath := filepath.Join(outDir, "manifest.json")
		if !isUpdate {
			manifestPath = ""
		}

		g, manifest, err := s.ScanWithManifest(manifestPath)
		if err != nil {
			fmt.Printf("Error scanning: %v\n", err)
			os.Exit(1)
		}

		// Run community clustering
		detector := cluster.NewDetector(g)
		communities := detector.Detect()

		// Save graph.json
		graphFile := filepath.Join(outDir, "graph.json")
		if err := g.Save(graphFile); err != nil {
			fmt.Printf("Error saving graph: %v\n", err)
			os.Exit(1)
		}

		// Save manifest.json
		if manifest != nil {
			_ = manifest.Save(filepath.Join(outDir, "manifest.json"))
		}

		// Generate GRAPH_REPORT.md
		if !skipReport {
			projectName := filepath.Base(dirs[0])
			if projectName == "." || projectName == "" {
				if wd, err := os.Getwd(); err == nil {
					projectName = filepath.Base(wd)
				}
			}
			rep := report.GenerateReport(g, projectName)
			reportFile := filepath.Join(outDir, "GRAPH_REPORT.md")
			_ = os.WriteFile(reportFile, []byte(rep), 0644)
		}

		// Generate graph.html
		if !skipViz {
			htmlFile := filepath.Join(outDir, "graph.html")
			_ = visualizer.GenerateHTML(g, htmlFile)
		}

		fmt.Printf("Scan complete. Extracted %d nodes, %d relationships in %d communities.\n",
			len(g.Nodes), len(g.Edges), len(communities))
		fmt.Printf("Outputs generated in %s/\n", outDir)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&outDir, "out", "o", "graphify-out", "Output directory for graph artifacts")
	scanCmd.Flags().BoolVarP(&isUpdate, "update", "u", false, "Incremental scan using manifest cache")
	scanCmd.Flags().BoolVar(&skipViz, "no-viz", false, "Skip generating interactive graph.html")
	scanCmd.Flags().BoolVar(&skipReport, "no-report", false, "Skip generating GRAPH_REPORT.md")
	rootCmd.AddCommand(scanCmd)
}
