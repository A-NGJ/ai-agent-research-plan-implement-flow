package frontmatter

import (
	"testing"
)

const testBody = `# Document Title

Some intro text.

## Summary

This is the summary section.

## Design Decisions

### Decision 1: Use maps

We decided to use maps.

### Decision 2: Use prefix matching

Prefix matching is flexible.

## Phase 1: Foundation

Build the foundation layer.

## Phase 2: Integration

Wire everything together.

## References

- Link 1
- Link 2
`

func TestExtractSectionBasic(t *testing.T) {
	content, ok := ExtractSection(testBody, "Summary")
	if !ok {
		t.Fatal("expected section to be found")
	}
	if content != "## Summary\n\nThis is the summary section.\n" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestExtractSectionCaseInsensitive(t *testing.T) {
	content, ok := ExtractSection(testBody, "summary")
	if !ok {
		t.Fatal("expected section to be found")
	}
	if content != "## Summary\n\nThis is the summary section.\n" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestExtractSectionPrefixMatch(t *testing.T) {
	content, ok := ExtractSection(testBody, "Phase 2")
	if !ok {
		t.Fatal("expected section to be found")
	}
	if content != "## Phase 2: Integration\n\nWire everything together.\n" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestExtractSectionsMultiplePrefixMatches(t *testing.T) {
	result := ExtractSections(testBody, []string{"Phase"})
	if len(result) != 2 {
		t.Fatalf("expected 2 matches, got %d: %v", len(result), result)
	}
	if _, ok := result["## Phase 1: Foundation"]; !ok {
		t.Error("missing Phase 1")
	}
	if _, ok := result["## Phase 2: Integration"]; !ok {
		t.Error("missing Phase 2")
	}
}

func TestExtractSectionLastSection(t *testing.T) {
	content, ok := ExtractSection(testBody, "References")
	if !ok {
		t.Fatal("expected section to be found")
	}
	expected := "## References\n\n- Link 1\n- Link 2\n"
	if content != expected {
		t.Errorf("unexpected content: %q, want %q", content, expected)
	}
}

func TestExtractSectionNoMatch(t *testing.T) {
	content, ok := ExtractSection(testBody, "Nonexistent")
	if ok {
		t.Error("expected no match")
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestExtractSectionEmptyBody(t *testing.T) {
	content, ok := ExtractSection("", "Summary")
	if ok {
		t.Error("expected no match on empty body")
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestExtractSectionSubHeadingsIncluded(t *testing.T) {
	content, ok := ExtractSection(testBody, "Design Decisions")
	if !ok {
		t.Fatal("expected section to be found")
	}
	expected := "## Design Decisions\n\n### Decision 1: Use maps\n\nWe decided to use maps.\n\n### Decision 2: Use prefix matching\n\nPrefix matching is flexible.\n"
	if content != expected {
		t.Errorf("unexpected content:\ngot:  %q\nwant: %q", content, expected)
	}
}

func TestExtractSectionsBatch(t *testing.T) {
	result := ExtractSections(testBody, []string{"Summary", "References"})
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if _, ok := result["## Summary"]; !ok {
		t.Error("missing Summary")
	}
	if _, ok := result["## References"]; !ok {
		t.Error("missing References")
	}
}

func TestExtractSectionsEmptyHeadings(t *testing.T) {
	result := ExtractSections(testBody, []string{})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestExtractSectionBodyStartsWithH2(t *testing.T) {
	body := "## Only Section\n\nContent here.\n"
	content, ok := ExtractSection(body, "Only Section")
	if !ok {
		t.Fatal("expected section to be found")
	}
	if content != body {
		t.Errorf("unexpected content: %q, want %q", content, body)
	}
}

func TestListSectionsMultiple(t *testing.T) {
	got := ListSections(testBody)
	want := []string{
		"## Summary",
		"## Design Decisions",
		"## Phase 1: Foundation",
		"## Phase 2: Integration",
		"## References",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d sections, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("section[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListSectionsEmptyBody(t *testing.T) {
	got := ListSections("")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestListSectionsNoHeadings(t *testing.T) {
	body := "# Title\n\nSome content without any ## headings.\n"
	got := ListSections(body)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestListSectionsBodyStartsWithH2(t *testing.T) {
	body := "## First\n\nContent.\n\n## Second\n\nMore content.\n"
	got := ListSections(body)
	want := []string{"## First", "## Second"}
	if len(got) != len(want) {
		t.Fatalf("expected %d sections, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("section[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
