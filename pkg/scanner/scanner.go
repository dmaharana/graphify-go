package scanner

import (
	"graphify-go/pkg/graph"
	"graphify-go/pkg/parser"
	"os"
	"path/filepath"
	"strings"
)

type Scanner struct {
	RootDir string
}

func NewScanner(rootDir string) *Scanner {
	return &Scanner{RootDir: rootDir}
}

func (s *Scanner) Scan() (*graph.Graph, error) {
	g := graph.NewGraph()
	
	err := filepath.Walk(s.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if s.isIgnored(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if s.isIgnored(path) {
			return nil
		}

		// Process file
		if isSupported(path) {
			fileNodes, fileEdges, err := parser.ParseFile(path)
			if err != nil {
				return nil
			}
			
			// Add file node
			fileID := path
			g.AddNode(graph.Node{
				ID:   fileID,
				Type: graph.FileNode,
				Name: filepath.Base(path),
			})

			for _, n := range fileNodes {
				g.AddNode(n)
				g.AddEdge(fileID, n.ID, "contains")
			}
			for _, e := range fileEdges {
				g.AddEdge(e.From, e.To, e.Type)
			}
		}

		return nil
	})

	return g, err
}

func (s *Scanner) isIgnored(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	ignored := []string{"node_modules", "vendor", "dist", "build", "target"}
	for _, i := range ignored {
		if strings.Contains(path, i) {
			return true
		}
	}
	return false
}

func isSupported(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h":
		return true
	}
	return false
}
