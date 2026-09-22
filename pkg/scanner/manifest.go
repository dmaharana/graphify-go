package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type FileEntry struct {
	Path    string `json:"path"`
	ModTime int64  `json:"mod_time"`
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
}

type Manifest struct {
	Entries map[string]FileEntry `json:"entries"`
}

func NewManifest() *Manifest {
	return &Manifest{
		Entries: make(map[string]FileEntry),
	}
}

func ComputeFileEntry(path string) (FileEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileEntry{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return FileEntry{}, err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return FileEntry{}, err
	}

	hashStr := hex.EncodeToString(hasher.Sum(nil))

	return FileEntry{
		Path:    filepath.ToSlash(path),
		ModTime: info.ModTime().UnixNano(),
		SHA256:  hashStr,
		Size:    info.Size(),
	}, nil
}

func (m *Manifest) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Entries == nil {
		m.Entries = make(map[string]FileEntry)
	}
	return &m, nil
}
