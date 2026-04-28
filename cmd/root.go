package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "graphify-go",
	Short: "Graphify-Go is a Golang equivalent of the graphify tool",
	Long:  `A tool to extract architectural context from codebases using AST parsing and graph generation.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add flags if needed
}
