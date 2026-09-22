package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type IgnoreMatcher struct {
	patterns []string
}

func NewIgnoreMatcher() *IgnoreMatcher {
	return &IgnoreMatcher{
		patterns: []string{
			".git",
			"node_modules",
			"vendor",
			"dist",
			"build",
			"target",
			"graphify-out",
			".graphify",
		},
	}
}

func (im *IgnoreMatcher) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Trim trailing slash for uniform matching
		line = strings.TrimSuffix(line, "/")
		line = strings.TrimPrefix(line, "./")
		im.patterns = append(im.patterns, line)
	}
	return scanner.Err()
}

func (im *IgnoreMatcher) IsIgnored(path string, isDir bool) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(path))
	base := filepath.Base(cleanPath)

	// Always ignore dot-directories except current directory
	if isDir && strings.HasPrefix(base, ".") && base != "." && base != ".." {
		return true
	}

	parts := strings.Split(cleanPath, "/")

	for _, pattern := range im.patterns {
		// Exact segment match e.g. "vendor" or "node_modules"
		for _, part := range parts {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}

		// Full path match or suffix match
		if matched, _ := filepath.Match(pattern, cleanPath); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		if strings.Contains(cleanPath, "/"+pattern+"/") || strings.HasSuffix(cleanPath, "/"+pattern) {
			return true
		}
	}

	return false
}
