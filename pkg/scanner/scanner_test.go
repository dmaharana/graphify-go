package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScannerMultiDirectoryAndIgnores(t *testing.T) {
	tmpDir := t.TempDir()

	// Create dir1/file1.go
	dir1 := filepath.Join(tmpDir, "dir1")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	code1 := `package dir1
func Func1() {}
`
	if err := os.WriteFile(filepath.Join(dir1, "file1.go"), []byte(code1), 0644); err != nil {
		t.Fatal(err)
	}

	// Create dir2/file2.go
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}
	code2 := `package dir2
func Func2() {}
`
	if err := os.WriteFile(filepath.Join(dir2, "file2.go"), []byte(code2), 0644); err != nil {
		t.Fatal(err)
	}

	// Create ignored file in dir2/vendor/secret.go
	vendorDir := filepath.Join(dir2, "vendor")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vendorDir, "secret.go"), []byte("package vendor"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .gitignore
	gitIgnorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitIgnorePath, []byte("vendor/\n*.tmp\n"), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScanner(dir1, dir2)
	s.LoadIgnores(tmpDir)

	g, err := s.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Should contain Func1 and Func2
	foundFunc1 := false
	foundFunc2 := false
	foundSecret := false
	for _, n := range g.Nodes {
		if n.Name == "Func1" {
			foundFunc1 = true
		}
		if n.Name == "Func2" {
			foundFunc2 = true
		}
		if n.Name == "secret.go" {
			foundSecret = true
		}
	}

	if !foundFunc1 {
		t.Errorf("expected Func1 from dir1")
	}
	if !foundFunc2 {
		t.Errorf("expected Func2 from dir2")
	}
	if foundSecret {
		t.Errorf("expected vendor/secret.go to be ignored")
	}
}

func TestManifestSaveAndDetect(t *testing.T) {
	tmpDir := t.TempDir()
	fpath := filepath.Join(tmpDir, "example.go")
	if err := os.WriteFile(fpath, []byte("package main\nfunc A() {}"), 0644); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	m := NewManifest()
	entry, err := ComputeFileEntry(fpath)
	if err != nil {
		t.Fatal(err)
	}
	m.Entries[fpath] = entry

	if err := m.Save(manifestPath); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[fpath].SHA256 != entry.SHA256 {
		t.Errorf("hash mismatch")
	}
}
