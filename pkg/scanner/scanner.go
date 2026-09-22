package scanner

import (
	"graphify-go/pkg/graph"
	"graphify-go/pkg/parser"
	"graphify-go/pkg/resolver"
	"os"
	"path/filepath"
)

type Scanner struct {
	Roots         []string
	IgnoreMatcher *IgnoreMatcher
	Manifest      *Manifest
}

func NewScanner(roots ...string) *Scanner {
	if len(roots) == 0 {
		roots = []string{"."}
	}
	return &Scanner{
		Roots:         roots,
		IgnoreMatcher: NewIgnoreMatcher(),
		Manifest:      NewManifest(),
	}
}

func (s *Scanner) LoadIgnores(baseDir string) {
	_ = s.IgnoreMatcher.LoadFile(filepath.Join(baseDir, ".gitignore"))
	_ = s.IgnoreMatcher.LoadFile(filepath.Join(baseDir, ".graphifyignore"))
}

func (s *Scanner) Scan() (*graph.Graph, error) {
	g, _, err := s.ScanWithManifest("")
	return g, err
}

func (s *Scanner) ScanWithManifest(manifestPath string) (*graph.Graph, *Manifest, error) {
	var oldManifest *Manifest
	if manifestPath != "" {
		oldManifest, _ = LoadManifest(manifestPath)
	}

	newManifest := NewManifest()
	r := resolver.NewResolver()

	for _, root := range s.Roots {
		s.LoadIgnores(root)

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath := filepath.ToSlash(path)

			if info.IsDir() {
				if s.IgnoreMatcher.IsIgnored(relPath, true) {
					return filepath.SkipDir
				}
				return nil
			}

			if s.IgnoreMatcher.IsIgnored(relPath, false) {
				return nil
			}

			if isSupported(path) {
				entry, err := ComputeFileEntry(path)
				if err == nil {
					newManifest.Entries[relPath] = entry
				}

				// Check if file is unchanged from old manifest
				// (For full resolution we still need symbols; in future we can cache AST json)
				_ = oldManifest

				res, err := parser.ParseFile(relPath)
				if err == nil {
					r.AddParseResult(res)
				}
			}

			return nil
		})

		if err != nil {
			return nil, nil, err
		}
	}

	s.Manifest = newManifest
	return r.Resolve(), newManifest, nil
}

func isSupported(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".js", ".ts", ".jsx", ".tsx":
		return true
	}
	return false
}
