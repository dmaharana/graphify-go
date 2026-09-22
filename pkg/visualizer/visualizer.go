package visualizer

import (
	_ "embed"
	"encoding/json"
	"graphify-go/pkg/graph"
	"os"
	"path/filepath"
	"strings"
)

//go:embed template.html
var htmlTemplate string

func GenerateHTML(g *graph.Graph, outputPath string) error {
	dataBytes, err := json.Marshal(g)
	if err != nil {
		return err
	}

	htmlContent := strings.Replace(htmlTemplate, "{{GRAPH_DATA}}", string(dataBytes), 1)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(htmlContent), 0644)
}
