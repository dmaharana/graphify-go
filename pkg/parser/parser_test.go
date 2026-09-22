package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGoFile(t *testing.T) {
	code := `package testpkg

import (
	"fmt"
	"strings"
)

type Config struct {
	Timeout int
}

type Service struct {
	Config
	name string
}

func (s *Service) Start() {
	s.init()
	fmt.Println("Starting...")
}

func (s *Service) init() {
	_ = strings.ToLower(s.name)
}

func Helper() {
	s := &Service{}
	s.Start()
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	res, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if res.PackageName != "testpkg" {
		t.Errorf("expected package name testpkg, got %s", res.PackageName)
	}

	// Check nodes
	nodeMap := make(map[string]bool)
	for _, n := range res.Nodes {
		nodeMap[n.Name] = true
	}

	for _, expected := range []string{"Config", "Service", "Start", "init", "Helper"} {
		if !nodeMap[expected] {
			t.Errorf("expected node %s in parsed nodes", expected)
		}
	}

	// Check edges: Service embeds Config
	hasEmbed := false
	for _, e := range res.Edges {
		if e.Type == "embeds" {
			hasEmbed = true
			break
		}
	}
	if !hasEmbed {
		t.Errorf("expected 'embeds' edge from Service to Config")
	}

	// Check raw calls
	hasFmtCall := false
	for _, rc := range res.RawCalls {
		if rc.Receiver == "fmt" && rc.CalleeName == "Println" {
			hasFmtCall = true
			break
		}
	}
	if !hasFmtCall {
		t.Errorf("expected RawCall for fmt.Println")
	}

	// Builtin check: append or len shouldn't produce raw calls
	// Check imports map
	if res.Imports["fmt"] != "fmt" {
		t.Errorf("expected import 'fmt', got %s", res.Imports["fmt"])
	}
	if res.Imports["strings"] != "strings" {
		t.Errorf("expected import 'strings', got %s", res.Imports["strings"])
	}
}

func TestParseGoInterfaceAndBuiltins(t *testing.T) {
	code := `package ifacepkg

type Reader interface {
	Read(p []byte) (n int, err error)
}

func Process(r Reader) {
	s := make([]int, 0)
	s = append(s, 1)
	_ = len(s)
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "iface.go")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	res, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Make sure builtins didn't create RawCalls
	for _, rc := range res.RawCalls {
		if rc.CalleeName == "make" || rc.CalleeName == "append" || rc.CalleeName == "len" {
			t.Errorf("builtin %s should not have generated a RawCall", rc.CalleeName)
		}
	}

	// Make sure interface method was captured
	hasReadMethod := false
	for _, n := range res.Nodes {
		if n.Type == "method" && n.Name == "Read" {
			hasReadMethod = true
			break
		}
	}
	if !hasReadMethod {
		t.Errorf("expected interface method 'Read'")
	}
}
