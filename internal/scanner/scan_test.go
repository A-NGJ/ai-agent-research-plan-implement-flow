package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, relPath, content string) string {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return full
}

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeFile(t, dir, "proposals/prop1.md",
		"---\ntopic: \"Proposal One\"\nstatus: draft\n---\n# Prop1\n")
	writeFile(t, dir, "proposals/prop2.md",
		"---\ntopic: \"Proposal Two\"\nstatus: complete\n---\n# Prop2\n")
	writeFile(t, dir, "plans/p1.md",
		"---\ntopic: \"Plan One\"\nstatus: draft\nproposal: proposals/prop1.md\n---\n# P1\n")
	writeFile(t, dir, "plans/p2.md",
		"# Plan Two\n\n## Source Documents\n- Proposal: `proposals/prop1.md`\n")
	writeFile(t, dir, "research/r1.md",
		"---\ntopic: \"Research One\"\nstatus: superseded\n---\n# R1\nReferences proposals/prop1.md in body.\n")
	writeFile(t, dir, "archive/proposals/old.md",
		"---\ntopic: \"Archived\"\nstatus: archived\n---\n# Old\n")

	return dir
}

func TestScanNoFilters(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 non-archive files: prop1, prop2, p1, p2, r1 (archive/ is skipped)
	if len(results) != 5 {
		t.Errorf("got %d results, want 5", len(results))
		for _, r := range results {
			t.Logf("  %s", r.Path)
		}
	}
}

func TestScanStatusFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Status: "draft"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// prop1 (draft) + p1 (draft) = 2
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
		for _, r := range results {
			t.Logf("  %s (%v)", r.Path, r.Status)
		}
	}
}

func TestScanTypeFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Type: "proposal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.Type != "proposal" {
			t.Errorf("got type %s, want proposal", r.Type)
		}
	}
}

func TestScanCombinedFilters(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Type: "proposal", Status: "draft"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if len(results) > 0 && (results[0].Title == nil || *results[0].Title != "Proposal One") {
		t.Errorf("expected Proposal One, got %v", results[0].Title)
	}
}

func TestScanProposalFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Proposal: "proposals/prop1.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only p1 has proposal: proposals/prop1.md
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestScanReferencesFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{References: "proposals/prop1.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// p1 (frontmatter proposal field), p2 (body reference), r1 (body reference) = 3
	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
		for _, r := range results {
			t.Logf("  %s", r.Path)
		}
	}
}

func TestScanArchivable(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Archivable: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// prop2 (complete), r1 (superseded) = 2
	// archive/ is skipped, so the archived file doesn't count
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
		for _, r := range results {
			t.Logf("  %s (%v)", r.Path, r.Status)
		}
	}
}

func TestScanArchiveExcluded(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, r := range results {
		if r.Type == "archive" || filepath.Base(r.Path) == "old.md" {
			t.Errorf("archive file should be excluded: %s", r.Path)
		}
	}
}

func TestScanEmptyResults(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Status: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestScanNoFrontmatter(t *testing.T) {
	dir := setupTestDir(t)

	// Filter for plans — p1 has frontmatter, p2 does not
	results, err := Scan(dir, Filters{Type: "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	// Find the one without frontmatter (p2)
	for _, r := range results {
		if strings.HasSuffix(r.Path, "p2.md") {
			if r.Status != nil {
				t.Errorf("p2 status should be nil, got %v", *r.Status)
			}
			if r.Title != nil {
				t.Errorf("p2 title should be nil, got %v", *r.Title)
			}
		}
	}
}

func TestFindReferencesFrontmatter(t *testing.T) {
	dir := setupTestDir(t)

	refs, err := FindReferences(dir, "proposals/prop1.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// p1 has proposal: proposals/prop1.md in frontmatter
	// p2 has body reference
	// r1 has body reference
	if len(refs) != 3 {
		t.Errorf("got %d refs, want 3", len(refs))
		for _, r := range refs {
			t.Logf("  %s -> %s", r.ReferencingFile, r.FieldOrLine)
		}
	}

	// Check that at least one is a frontmatter field reference
	foundField := false
	for _, r := range refs {
		if r.FieldOrLine == "proposal: proposals/prop1.md" {
			foundField = true
		}
	}
	if !foundField {
		t.Error("expected a frontmatter field reference 'proposal: proposals/prop1.md'")
	}
}

func TestFindReferencesUnreferenced(t *testing.T) {
	dir := setupTestDir(t)

	refs, err := FindReferences(dir, "nonexistent/file.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("got %d refs, want 0", len(refs))
	}
}

func TestCountReferences(t *testing.T) {
	dir := setupTestDir(t)

	count, err := CountReferences(dir, "proposals/prop1.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 3 {
		t.Errorf("got count %d, want 3", count)
	}
}

func TestCountReferencesZero(t *testing.T) {
	dir := setupTestDir(t)

	count, err := CountReferences(dir, "nonexistent.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("got count %d, want 0", count)
	}
}

func TestFindReferencesBodyLine(t *testing.T) {
	dir := setupTestDir(t)

	refs, err := FindReferences(dir, "proposals/prop1.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// r1 references proposals/prop1.md in body
	foundBody := false
	for _, r := range refs {
		if strings.Contains(r.ReferencingFile, "r1.md") && strings.Contains(r.FieldOrLine, "References proposals/prop1.md") {
			foundBody = true
		}
	}
	if !foundBody {
		t.Error("expected a body line reference from r1.md")
	}
}

func TestScanEmptyDir(t *testing.T) {
	dir := t.TempDir()

	results, err := Scan(dir, Filters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
