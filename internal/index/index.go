package index

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Metadata holds index-level information.
type Metadata struct {
	Version     string    `json:"version"`
	BuiltAt     time.Time `json:"built_at"`
	RootPath    string    `json:"root_path"`
	FileCount   int       `json:"file_count"`
	SymbolCount int       `json:"symbol_count"`
}

// FileEntry describes a single indexed file.
type FileEntry struct {
	Path     string    `json:"path"`
	Language string    `json:"language"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Hash     string    `json:"hash"`
}

// Symbol describes a single extracted symbol.
type Symbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Package   string `json:"package"`
	Scope     string `json:"scope"`
	Signature string `json:"signature"`
	Exported  bool   `json:"exported"`
}

// Index is the top-level structure persisted to .rpi/index.json.
type Index struct {
	Metadata Metadata    `json:"metadata"`
	Files    []FileEntry `json:"files"`
	Symbols  []Symbol    `json:"symbols"`
}

// BuildOptions controls what Build indexes.
type BuildOptions struct {
	Languages    []string // filter to these languages only (empty = all)
	ForceRebuild bool
}

// skipDirs are directories that should never be indexed.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".venv":        true,
	"build":        true,
	"dist":         true,
	".rpi":         true,
}

// Build walks a codebase and builds a symbol index.
func Build(rootPath string, opts BuildOptions) (*Index, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}

	langFilter := make(map[string]bool)
	for _, l := range opts.Languages {
		langFilter[l] = true
	}

	var files []FileEntry
	var symbols []Symbol

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		lang := DetectLanguage(path)
		if lang == "" {
			return nil
		}
		if len(langFilter) > 0 && !langFilter[lang] {
			return nil
		}

		cfg := GetConfig(lang)
		if cfg == nil {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			relPath = path
		}

		hash, err := hashFile(path)
		if err != nil {
			return nil // skip files we can't hash
		}

		fe := FileEntry{
			Path:     relPath,
			Language: lang,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Hash:     hash,
		}
		files = append(files, fe)

		syms, _, err := ExtractSymbols(path, cfg)
		if err != nil {
			return nil // skip files we can't parse
		}

		// Rewrite file paths to relative.
		for i := range syms {
			syms[i].File = relPath
		}
		symbols = append(symbols, syms...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk codebase: %w", err)
	}

	if files == nil {
		files = []FileEntry{}
	}
	if symbols == nil {
		symbols = []Symbol{}
	}

	idx := &Index{
		Metadata: Metadata{
			Version:     CurrentVersion,
			BuiltAt:     time.Now(),
			RootPath:    absRoot,
			FileCount:   len(files),
			SymbolCount: len(symbols),
		},
		Files:   files,
		Symbols: symbols,
	}
	return idx, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// IsGitignored checks if .rpi/ is in the project's .gitignore.
func IsGitignored(rootPath string) bool {
	data, err := os.ReadFile(filepath.Join(rootPath, ".gitignore"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == ".rpi/" || strings.TrimSpace(line) == ".rpi" {
			return true
		}
	}
	return false
}
