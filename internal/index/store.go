package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	CurrentVersion   = "1"
	DefaultIndexPath = ".rpi/index.json"
)

// Save writes an index to a JSON file, creating parent directories as needed.
func Save(idx *Index, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create index directory: %w", err)
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// Load reads an index from a JSON file and validates the schema version.
func Load(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("unmarshal index: %w", err)
	}
	if idx.Metadata.Version != CurrentVersion {
		return nil, fmt.Errorf("index version %q is not supported (expected %q) — run 'rpi index build' to rebuild", idx.Metadata.Version, CurrentVersion)
	}
	return &idx, nil
}
