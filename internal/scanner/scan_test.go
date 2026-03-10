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

	writeFile(t, dir, "tickets/t1.md",
		"---\ntopic: \"Ticket One\"\nstatus: draft\nticket_id: t-001\ndesign: designs/d1.md\n---\n# T1\n")
	writeFile(t, dir, "tickets/t2.md",
		"---\ntopic: \"Ticket Two\"\nstatus: complete\nticket_id: t-002\n---\n# T2\n")
	writeFile(t, dir, "designs/d1.md",
		"---\ntopic: \"Design One\"\nstatus: draft\n---\n# D1\n")
	writeFile(t, dir, "designs/d2.md",
		"---\ntopic: \"Design Two\"\nstatus: complete\n---\n# D2\n")
	writeFile(t, dir, "plans/p1.md",
		"# Plan One\n\n## Source Documents\n- Design: `designs/d1.md`\n")
	writeFile(t, dir, "research/r1.md",
		"---\ntopic: \"Research One\"\nstatus: superseded\n---\n# R1\nReferences designs/d1.md in body.\n")
	writeFile(t, dir, "archive/tickets/old.md",
		"---\ntopic: \"Archived\"\nstatus: archived\n---\n# Old\n")

	return dir
}

func TestScanNoFilters(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 6 non-archive files: d1, d2, p1, r1, t1, t2 (archive/ is skipped)
	if len(results) != 6 {
		t.Errorf("got %d results, want 6", len(results))
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

	// t1 (draft) + d1 (draft) = 2
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
		for _, r := range results {
			t.Logf("  %s (%v)", r.Path, r.Status)
		}
	}
}

func TestScanTypeFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Type: "ticket"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.Type != "ticket" {
			t.Errorf("got type %s, want ticket", r.Type)
		}
	}
}

func TestScanCombinedFilters(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Type: "ticket", Status: "draft"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if len(results) > 0 && (results[0].TicketID == nil || *results[0].TicketID != "t-001") {
		t.Errorf("expected ticket t-001, got %v", results[0].TicketID)
	}
}

func TestScanDesignFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{Design: "designs/d1.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only t1 has design: designs/d1.md
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestScanReferencesFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := Scan(dir, Filters{References: "designs/d1.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// t1 (frontmatter design field), p1 (body reference), r1 (body reference) = 3
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

	// t2 (complete), d2 (complete), r1 (superseded) = 3
	// archive/ is skipped, so the archived file doesn't count
	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
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

	// p1.md has no frontmatter — should appear with nil status/title
	results, err := Scan(dir, Filters{Type: "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Status != nil {
		t.Errorf("plan status should be nil, got %v", *results[0].Status)
	}
	if results[0].Title != nil {
		t.Errorf("plan title should be nil, got %v", *results[0].Title)
	}
}

func TestFindReferencesFrontmatter(t *testing.T) {
	dir := setupTestDir(t)

	refs, err := FindReferences(dir, "designs/d1.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// t1 has design: designs/d1.md in frontmatter
	// p1 has body reference
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
		if r.FieldOrLine == "design: designs/d1.md" {
			foundField = true
		}
	}
	if !foundField {
		t.Error("expected a frontmatter field reference 'design: designs/d1.md'")
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

	count, err := CountReferences(dir, "designs/d1.md")
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

	refs, err := FindReferences(dir, "designs/d1.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// r1 references designs/d1.md in body
	foundBody := false
	for _, r := range refs {
		if strings.Contains(r.ReferencingFile, "r1.md") && strings.Contains(r.FieldOrLine, "References designs/d1.md") {
			foundBody = true
		}
	}
	if !foundBody {
		t.Error("expected a body line reference from r1.md")
	}
}

func TestFindByTicketID(t *testing.T) {
	dir := setupTestDir(t)

	path, err := FindByTicketID(dir, "t-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected to find ticket t-001")
	}
	if !strings.HasSuffix(path, "t1.md") {
		t.Errorf("expected path ending in t1.md, got %s", path)
	}
}

func TestFindByTicketIDNotFound(t *testing.T) {
	dir := setupTestDir(t)

	path, err := FindByTicketID(dir, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path, got %s", path)
	}
}

func TestFindByTicketIDNoTicketsDir(t *testing.T) {
	dir := t.TempDir()

	path, err := FindByTicketID(dir, "t-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path, got %s", path)
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
