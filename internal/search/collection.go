package search

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// maxSlugLen caps the human-readable portion of a collection name so the
	// final identifier stays comfortably short across qmd's tooling.
	maxSlugLen = 30
	// shortHashLen is the number of hex chars taken from the path digest.
	shortHashLen = 6
)

var slugReplaceRe = regexp.MustCompile(`[^a-z0-9]+`)

// CollectionName derives a stable, per-project qmd collection name from the
// absolute .rpi/ directory path. Two different projects always produce
// different names — the short hash of the absolute path makes collisions
// astronomically unlikely even when two projects share a parent name.
func CollectionName(rpiDir string) (string, error) {
	abs, err := filepath.Abs(rpiDir)
	if err != nil {
		return "", fmt.Errorf("resolve rpiDir: %w", err)
	}
	repoName := filepath.Base(filepath.Dir(abs))
	slug := slugify(repoName)
	if slug == "" {
		slug = "project"
	}

	sum := sha1.Sum([]byte(abs))
	short := hex.EncodeToString(sum[:])[:shortHashLen]

	return "rpi-" + slug + "-" + short, nil
}

// slugify normalizes a path component into a lowercase, hyphen-separated
// token suitable for an identifier. Sequences of non-alphanumeric characters
// collapse to a single hyphen; leading and trailing hyphens are trimmed; the
// result is capped at maxSlugLen.
func slugify(s string) string {
	lower := strings.ToLower(s)
	replaced := slugReplaceRe.ReplaceAllString(lower, "-")
	trimmed := strings.Trim(replaced, "-")
	if len(trimmed) > maxSlugLen {
		trimmed = strings.TrimRight(trimmed[:maxSlugLen], "-")
	}
	return trimmed
}
