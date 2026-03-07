package chain

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestResolveSimpleChain(t *testing.T) {
	dir := t.TempDir()

	designPath := writeFile(t, dir, ".thoughts/designs/design.md",
		"---\ntopic: \"My Design\"\nstatus: complete\nrelated_research: .thoughts/research/research.md\n---\n# Design\n")

	// Make the research path relative to match what's in frontmatter — but actually
	// the resolver uses the path as-is from frontmatter. Let's create both files.
	writeFile(t, dir, ".thoughts/research/research.md",
		"---\ntopic: \"My Research\"\nstatus: complete\n---\n# Research\n")

	ticketPath := writeFile(t, dir, ".thoughts/tickets/test-001.md",
		"---\ntopic: \"Test Ticket\"\nstatus: draft\nticket_id: test-001\ndesign: "+designPath+"\n---\n# Ticket\n")

	result, err := Resolve(ticketPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Root != ticketPath {
		t.Errorf("root = %s, want %s", result.Root, ticketPath)
	}

	if len(result.Artifacts) != 2 {
		// ticket + design (research uses relative path so won't resolve from absolute design path)
		t.Fatalf("got %d artifacts, want 2", len(result.Artifacts))
	}

	// First artifact is the root
	if result.Artifacts[0].Path != ticketPath {
		t.Errorf("first artifact path = %s, want %s", result.Artifacts[0].Path, ticketPath)
	}
	if result.Artifacts[0].Type != "ticket" {
		t.Errorf("first artifact type = %s, want ticket", result.Artifacts[0].Type)
	}
	if result.Artifacts[0].TicketID == nil || *result.Artifacts[0].TicketID != "test-001" {
		t.Errorf("ticket_id = %v, want test-001", result.Artifacts[0].TicketID)
	}
}

func TestResolveCycleDetection(t *testing.T) {
	dir := t.TempDir()

	aPath := filepath.Join(dir, ".thoughts/designs/a.md")
	bPath := filepath.Join(dir, ".thoughts/designs/b.md")

	writeFile(t, dir, ".thoughts/designs/a.md",
		"---\ntopic: A\nstatus: draft\ndesign: "+bPath+"\n---\n# A\n")
	writeFile(t, dir, ".thoughts/designs/b.md",
		"---\ntopic: B\nstatus: draft\ndesign: "+aPath+"\n---\n# B\n")

	result, err := Resolve(aPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Artifacts) != 2 {
		t.Fatalf("got %d artifacts, want 2 (cycle should not cause duplicates)", len(result.Artifacts))
	}
}

func TestResolveMissingFile(t *testing.T) {
	dir := t.TempDir()

	ticketPath := writeFile(t, dir, ".thoughts/tickets/t.md",
		"---\ntopic: T\nstatus: draft\ndesign: /nonexistent/design.md\n---\n# T\n")

	result, err := Resolve(ticketPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Artifacts) != 1 {
		t.Fatalf("got %d artifacts, want 1", len(result.Artifacts))
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for missing file, got none")
	}
}

func TestResolveNoFrontmatterFallback(t *testing.T) {
	dir := t.TempDir()

	designPath := writeFile(t, dir, ".thoughts/designs/design.md",
		"---\ntopic: Design\nstatus: complete\n---\n# Design\n")

	planPath := writeFile(t, dir, ".thoughts/plans/plan.md",
		"# Plan\n\n## Source Documents\n- Design: `"+designPath+"`\n- Research: `.thoughts/research/r.md`\n")

	result, err := Resolve(planPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Plan itself + design (research file missing → warning)
	if len(result.Artifacts) < 2 {
		t.Fatalf("got %d artifacts, want at least 2", len(result.Artifacts))
	}

	if result.Artifacts[0].Type != "plan" {
		t.Errorf("first artifact type = %s, want plan", result.Artifacts[0].Type)
	}
	if result.Artifacts[0].Status != nil {
		t.Errorf("plan status should be nil (no frontmatter), got %v", *result.Artifacts[0].Status)
	}
}

func TestResolveSingleFile(t *testing.T) {
	dir := t.TempDir()

	path := writeFile(t, dir, ".thoughts/designs/solo.md",
		"---\ntopic: Solo\nstatus: draft\n---\n# Solo\n")

	result, err := Resolve(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Artifacts) != 1 {
		t.Fatalf("got %d artifacts, want 1", len(result.Artifacts))
	}
	if result.Artifacts[0].LinksTo == nil || len(result.Artifacts[0].LinksTo) != 0 {
		t.Errorf("links_to should be empty slice, got %v", result.Artifacts[0].LinksTo)
	}
}

func TestResolveDependsOnList(t *testing.T) {
	dir := t.TempDir()

	dep1 := writeFile(t, dir, ".thoughts/tickets/dep1.md",
		"---\ntopic: Dep1\nstatus: complete\n---\n# Dep1\n")
	dep2 := writeFile(t, dir, ".thoughts/tickets/dep2.md",
		"---\ntopic: Dep2\nstatus: complete\n---\n# Dep2\n")

	mainPath := writeFile(t, dir, ".thoughts/tickets/main.md",
		"---\ntopic: Main\nstatus: draft\ndepends_on:\n  - "+dep1+"\n  - "+dep2+"\n---\n# Main\n")

	result, err := Resolve(mainPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Artifacts) != 3 {
		t.Fatalf("got %d artifacts, want 3", len(result.Artifacts))
	}
}

func TestResolveMaxDepth(t *testing.T) {
	dir := t.TempDir()

	// Create a chain of 15 files, each linking to the next
	paths := make([]string, 15)
	for i := 14; i >= 0; i-- {
		link := ""
		if i < 14 {
			link = "design: " + paths[i+1] + "\n"
		}
		paths[i] = writeFile(t, dir, filepath.Join(".thoughts/designs", fmt.Sprintf("d%d.md", i)),
			"---\ntopic: D"+fmt.Sprintf("%d", i)+"\nstatus: draft\n"+link+"---\n# D\n")
	}

	result, err := Resolve(paths[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should stop at max depth (11 artifacts: depth 0 through 10)
	if len(result.Artifacts) > 12 {
		t.Errorf("got %d artifacts, expected max depth to limit resolution", len(result.Artifacts))
	}
	if len(result.Warnings) == 0 {
		t.Error("expected max depth warning")
	}
}

func TestInferType(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{".thoughts/plans/foo.md", "plan"},
		{".thoughts/tickets/foo.md", "ticket"},
		{".thoughts/designs/foo.md", "design"},
		{".thoughts/research/foo.md", "research"},
		{".thoughts/structures/foo.md", "structure"},
		{".thoughts/prs/foo.md", "pr"},
		{".thoughts/reviews/foo.md", "review"},
		{".thoughts/archive/plans/foo.md", "archive"},
		{"random/path.md", "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := inferType(tc.path)
			if got != tc.want {
				t.Errorf("inferType(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}
