package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildBasic(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "main.go", `package main

func main() {}
func helper() {}
`)
	writeTestFile(t, dir, "app.py", `class App:
    def run(self):
        pass
`)

	idx, err := Build(dir, BuildOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if idx.Metadata.FileCount != 2 {
		t.Errorf("FileCount = %d, want 2", idx.Metadata.FileCount)
	}
	if idx.Metadata.SymbolCount < 4 {
		t.Errorf("SymbolCount = %d, want at least 4", idx.Metadata.SymbolCount)
	}
	if idx.Metadata.Version != CurrentVersion {
		t.Errorf("Version = %q, want %q", idx.Metadata.Version, CurrentVersion)
	}
}

func TestBuildSkipsDirs(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	writeTestFile(t, dir, "node_modules/dep/index.js", "function dep() {}\n")
	writeTestFile(t, dir, "vendor/lib/lib.go", "package lib\nfunc Lib() {}\n")
	writeTestFile(t, dir, ".git/config.go", "package git\nfunc Config() {}\n")
	writeTestFile(t, dir, "__pycache__/mod.py", "def cached(): pass\n")
	writeTestFile(t, dir, ".rpi/index.json", "{}")

	idx, err := Build(dir, BuildOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if idx.Metadata.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1 (only main.go)", idx.Metadata.FileCount)
	}
	for _, f := range idx.Files {
		if f.Path != "main.go" {
			t.Errorf("unexpected file in index: %s", f.Path)
		}
	}
}

func TestBuildLanguageFilter(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	writeTestFile(t, dir, "app.py", "def app(): pass\n")
	writeTestFile(t, dir, "index.ts", "export function main() {}\n")

	idx, err := Build(dir, BuildOptions{Languages: []string{"go"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if idx.Metadata.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1", idx.Metadata.FileCount)
	}
	if idx.Files[0].Language != "go" {
		t.Errorf("Language = %q, want go", idx.Files[0].Language)
	}
}

func TestBuildRelativePaths(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "pkg/server/handler.go", "package server\nfunc Handle() {}\n")

	idx, err := Build(dir, BuildOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(idx.Files) != 1 {
		t.Fatalf("got %d files, want 1", len(idx.Files))
	}
	want := filepath.Join("pkg", "server", "handler.go")
	if idx.Files[0].Path != want {
		t.Errorf("file path = %q, want %q", idx.Files[0].Path, want)
	}
	if len(idx.Symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}
	if idx.Symbols[0].File != want {
		t.Errorf("symbol file = %q, want %q", idx.Symbols[0].File, want)
	}
}

func TestBuildEmptyDir(t *testing.T) {
	dir := t.TempDir()

	idx, err := Build(dir, BuildOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx.Metadata.FileCount != 0 {
		t.Errorf("FileCount = %d, want 0", idx.Metadata.FileCount)
	}
	if idx.Metadata.SymbolCount != 0 {
		t.Errorf("SymbolCount = %d, want 0", idx.Metadata.SymbolCount)
	}
	if idx.Files == nil {
		t.Error("Files should be empty slice, not nil")
	}
	if idx.Symbols == nil {
		t.Error("Symbols should be empty slice, not nil")
	}
}

func TestBuildFileHashes(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "main.go", "package main\nfunc main() {}\n")

	idx, err := Build(dir, BuildOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(idx.Files) != 1 {
		t.Fatalf("got %d files, want 1", len(idx.Files))
	}
	if idx.Files[0].Hash == "" {
		t.Error("expected non-empty hash")
	}
	if len(idx.Files[0].Hash) != 64 { // SHA256 hex
		t.Errorf("hash length = %d, want 64", len(idx.Files[0].Hash))
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, ".rpi", "index.json")

	original := &Index{
		Metadata: Metadata{
			Version:     CurrentVersion,
			FileCount:   1,
			SymbolCount: 2,
			RootPath:    dir,
		},
		Files: []FileEntry{
			{Path: "main.go", Language: "go", Size: 100, Hash: "abc123"},
		},
		Symbols: []Symbol{
			{Name: "main", Kind: "function", File: "main.go", Line: 3},
			{Name: "Server", Kind: "struct", File: "main.go", Line: 10, Exported: true},
		},
	}

	if err := Save(original, indexPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(indexPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Metadata.Version != original.Metadata.Version {
		t.Errorf("Version = %q, want %q", loaded.Metadata.Version, original.Metadata.Version)
	}
	if loaded.Metadata.FileCount != original.Metadata.FileCount {
		t.Errorf("FileCount = %d, want %d", loaded.Metadata.FileCount, original.Metadata.FileCount)
	}
	if len(loaded.Files) != len(original.Files) {
		t.Errorf("Files count = %d, want %d", len(loaded.Files), len(original.Files))
	}
	if len(loaded.Symbols) != len(original.Symbols) {
		t.Errorf("Symbols count = %d, want %d", len(loaded.Symbols), len(original.Symbols))
	}
	if loaded.Symbols[1].Exported != true {
		t.Error("expected Symbols[1].Exported = true")
	}
}

func TestLoadVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.json")

	content := `{"metadata":{"version":"99"},"files":[],"symbols":[]}`
	if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(indexPath)
	if err == nil {
		t.Fatal("expected error for version mismatch")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/index.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIsGitignored(t *testing.T) {
	dir := t.TempDir()

	if IsGitignored(dir) {
		t.Error("expected false with no .gitignore")
	}

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules/\n.rpi/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsGitignored(dir) {
		t.Error("expected true with .rpi/ in .gitignore")
	}
}
