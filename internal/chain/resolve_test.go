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

	proposalPath := writeFile(t, dir, ".thoughts/proposals/proposal.md",
		"---\ntopic: \"My Proposal\"\nstatus: complete\nrelated_research: .thoughts/research/research.md\n---\n# Proposal\n")

	writeFile(t, dir, ".thoughts/research/research.md",
		"---\ntopic: \"My Research\"\nstatus: complete\n---\n# Research\n")

	planPath := writeFile(t, dir, ".thoughts/plans/test-plan.md",
		"---\ntopic: \"Test Plan\"\nstatus: draft\nproposal: "+proposalPath+"\n---\n# Plan\n")

	result, err := Resolve(planPath, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Root != planPath {
		t.Errorf("root = %s, want %s", result.Root, planPath)
	}

	if len(result.Artifacts) != 2 {
		// plan + proposal (research uses relative path so won't resolve from absolute proposal path)
		t.Fatalf("got %d artifacts, want 2", len(result.Artifacts))
	}

	// First artifact is the root
	if result.Artifacts[0].Path != planPath {
		t.Errorf("first artifact path = %s, want %s", result.Artifacts[0].Path, planPath)
	}
	if result.Artifacts[0].Type != "plan" {
		t.Errorf("first artifact type = %s, want plan", result.Artifacts[0].Type)
	}
}

func TestResolveCycleDetection(t *testing.T) {
	dir := t.TempDir()

	aPath := filepath.Join(dir, ".thoughts/proposals/a.md")
	bPath := filepath.Join(dir, ".thoughts/proposals/b.md")

	writeFile(t, dir, ".thoughts/proposals/a.md",
		"---\ntopic: A\nstatus: draft\nproposal: "+bPath+"\n---\n# A\n")
	writeFile(t, dir, ".thoughts/proposals/b.md",
		"---\ntopic: B\nstatus: draft\nproposal: "+aPath+"\n---\n# B\n")

	result, err := Resolve(aPath, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Artifacts) != 2 {
		t.Fatalf("got %d artifacts, want 2 (cycle should not cause duplicates)", len(result.Artifacts))
	}
}

func TestResolveMissingFile(t *testing.T) {
	dir := t.TempDir()

	planPath := writeFile(t, dir, ".thoughts/plans/p.md",
		"---\ntopic: P\nstatus: draft\nproposal: /nonexistent/proposal.md\n---\n# P\n")

	result, err := Resolve(planPath, ResolveOptions{})
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

	proposalPath := writeFile(t, dir, ".thoughts/proposals/proposal.md",
		"---\ntopic: Proposal\nstatus: complete\n---\n# Proposal\n")

	planPath := writeFile(t, dir, ".thoughts/plans/plan.md",
		"# Plan\n\n## Source Documents\n- Proposal: `"+proposalPath+"`\n- Research: `.thoughts/research/r.md`\n")

	result, err := Resolve(planPath, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Plan itself + proposal (research file missing → warning)
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

	path := writeFile(t, dir, ".thoughts/proposals/solo.md",
		"---\ntopic: Solo\nstatus: draft\n---\n# Solo\n")

	result, err := Resolve(path, ResolveOptions{})
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

	dep1 := writeFile(t, dir, ".thoughts/plans/dep1.md",
		"---\ntopic: Dep1\nstatus: complete\n---\n# Dep1\n")
	dep2 := writeFile(t, dir, ".thoughts/plans/dep2.md",
		"---\ntopic: Dep2\nstatus: complete\n---\n# Dep2\n")

	mainPath := writeFile(t, dir, ".thoughts/plans/main.md",
		"---\ntopic: Main\nstatus: draft\ndepends_on:\n  - "+dep1+"\n  - "+dep2+"\n---\n# Main\n")

	result, err := Resolve(mainPath, ResolveOptions{})
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
			link = "proposal: " + paths[i+1] + "\n"
		}
		paths[i] = writeFile(t, dir, filepath.Join(".thoughts/proposals", fmt.Sprintf("p%d.md", i)),
			"---\ntopic: P"+fmt.Sprintf("%d", i)+"\nstatus: draft\n"+link+"---\n# P\n")
	}

	result, err := Resolve(paths[0], ResolveOptions{})
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

func TestResolveSectionsExtraction(t *testing.T) {
	dir := t.TempDir()

	path := writeFile(t, dir, ".thoughts/proposals/proposal.md",
		"---\ntopic: My Proposal\nstatus: complete\n---\n# Proposal\n\n## Summary\n\nThis is the summary.\n\n## Architecture\n\nDiagram here.\n")

	result, err := Resolve(path, ResolveOptions{Sections: []string{"Summary"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Artifacts) != 1 {
		t.Fatalf("got %d artifacts, want 1", len(result.Artifacts))
	}

	sections := result.Artifacts[0].Sections
	if sections == nil {
		t.Fatal("expected sections to be populated")
	}
	if _, ok := sections["## Summary"]; !ok {
		t.Errorf("expected '## Summary' key in sections, got keys: %v", sections)
	}
}

func TestResolveSectionsEmptyOptions(t *testing.T) {
	dir := t.TempDir()

	path := writeFile(t, dir, ".thoughts/proposals/proposal.md",
		"---\ntopic: My Proposal\nstatus: complete\n---\n# Proposal\n\n## Summary\n\nContent.\n")

	result, err := Resolve(path, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Artifacts[0].Sections != nil {
		t.Errorf("expected nil sections with empty options, got %v", result.Artifacts[0].Sections)
	}
}

func TestResolveSectionsNoMatch(t *testing.T) {
	dir := t.TempDir()

	path := writeFile(t, dir, ".thoughts/proposals/proposal.md",
		"---\ntopic: My Proposal\nstatus: complete\n---\n# Proposal\n\n## Summary\n\nContent.\n")

	result, err := Resolve(path, ResolveOptions{Sections: []string{"Nonexistent"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Artifacts[0].Sections != nil {
		t.Errorf("expected nil sections when no match, got %v", result.Artifacts[0].Sections)
	}
}

func TestInferType(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{".thoughts/plans/foo.md", "plan"},
		{".thoughts/proposals/foo.md", "proposal"},
		{".thoughts/research/foo.md", "research"},
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
