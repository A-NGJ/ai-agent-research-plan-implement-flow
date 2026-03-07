package frontmatter

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseWithFrontmatter(t *testing.T) {
	input := "---\ntitle: \"Hello World\"\nstatus: draft\ntags: [a, b]\n---\n# Body\n\nSome content.\n"
	doc, err := ParseBytes([]byte(input), "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter["title"] != "Hello World" {
		t.Errorf("title = %v, want %q", doc.Frontmatter["title"], "Hello World")
	}
	if doc.Frontmatter["status"] != "draft" {
		t.Errorf("status = %v, want %q", doc.Frontmatter["status"], "draft")
	}
	if doc.Body != "# Body\n\nSome content.\n" {
		t.Errorf("body = %q, want %q", doc.Body, "# Body\n\nSome content.\n")
	}
}

func TestParseWithoutFrontmatter(t *testing.T) {
	input := "# Just a heading\n\nNo frontmatter here.\n"
	doc, err := ParseBytes([]byte(input), "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Frontmatter) != 0 {
		t.Errorf("frontmatter = %v, want empty map", doc.Frontmatter)
	}
	if doc.Body != input {
		t.Errorf("body = %q, want %q", doc.Body, input)
	}
}

func TestParseEmptyFile(t *testing.T) {
	doc, err := ParseBytes([]byte(""), "empty.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Frontmatter) != 0 {
		t.Errorf("frontmatter = %v, want empty map", doc.Frontmatter)
	}
	if doc.Body != "" {
		t.Errorf("body = %q, want empty", doc.Body)
	}
}

func TestParseFixtureWithFrontmatter(t *testing.T) {
	path := filepath.Join("testdata", "with-frontmatter.md")

	doc, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter["status"] != "complete" {
		t.Errorf("status = %v, want %q", doc.Frontmatter["status"], "complete")
	}
	if doc.Frontmatter["topic"] == nil {
		t.Error("topic field missing")
	}
	if !strings.HasPrefix(doc.Body, "\n# Design:") && !strings.HasPrefix(doc.Body, "# Design:") {
		t.Errorf("body should start with design heading, got: %q", doc.Body[:50])
	}
}

func TestParseFixtureWithoutFrontmatter(t *testing.T) {
	path := filepath.Join("testdata", "without-frontmatter.md")

	doc, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Frontmatter) != 0 {
		t.Errorf("should have no frontmatter, got %v", doc.Frontmatter)
	}
	if !strings.HasPrefix(doc.Body, "# Rename Commands") {
		t.Errorf("body should start with heading, got: %q", doc.Body[:50])
	}
}

func TestSerializeRoundTrip(t *testing.T) {
	input := "---\nstatus: draft\ntitle: Test\n---\n# Body\n\nContent here.\n"
	doc, err := ParseBytes([]byte(input), "test.md")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	data, err := Serialize(doc)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	// Re-parse the serialized output
	doc2, err := ParseBytes(data, "test.md")
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	if doc2.Frontmatter["status"] != "draft" {
		t.Errorf("status = %v, want %q", doc2.Frontmatter["status"], "draft")
	}
	if doc2.Frontmatter["title"] != "Test" {
		t.Errorf("title = %v, want %q", doc2.Frontmatter["title"], "Test")
	}
	if doc2.Body != doc.Body {
		t.Errorf("body changed after round-trip:\ngot:  %q\nwant: %q", doc2.Body, doc.Body)
	}
}

func TestWritePreservesBody(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	original := "---\nstatus: draft\n---\n# My Document\n\nThis body must not change.\n"
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	doc, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	// Modify frontmatter
	doc.Frontmatter["new_field"] = "new_value"
	if err := Write(doc); err != nil {
		t.Fatal(err)
	}

	// Re-parse and check body
	doc2, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	if doc2.Body != doc.Body {
		t.Errorf("body changed after write:\ngot:  %q\nwant: %q", doc2.Body, doc.Body)
	}
	if doc2.Frontmatter["new_field"] != "new_value" {
		t.Error("new_field not persisted")
	}
}

func TestTransitionValid(t *testing.T) {
	cases := []struct {
		from, to string
	}{
		{"draft", "active"},
		{"draft", "superseded"},
		{"active", "complete"},
		{"active", "superseded"},
		{"complete", "archived"},
		{"complete", "superseded"},
	}

	for _, tc := range cases {
		t.Run(tc.from+"→"+tc.to, func(t *testing.T) {
			doc := &Document{
				Frontmatter: map[string]interface{}{"status": tc.from},
			}
			if err := Transition(doc, tc.to); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if doc.Frontmatter["status"] != tc.to {
				t.Errorf("status = %v, want %q", doc.Frontmatter["status"], tc.to)
			}
		})
	}
}

func TestTransitionInvalid(t *testing.T) {
	cases := []struct {
		from, to string
	}{
		{"draft", "complete"},
		{"draft", "archived"},
		{"active", "draft"},
		{"active", "archived"},
		{"complete", "draft"},
		{"complete", "active"},
	}

	for _, tc := range cases {
		t.Run(tc.from+"→"+tc.to, func(t *testing.T) {
			doc := &Document{
				Frontmatter: map[string]interface{}{"status": tc.from},
			}
			err := Transition(doc, tc.to)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		})
	}
}

func TestTransitionMissingStatus(t *testing.T) {
	doc := &Document{
		Frontmatter: map[string]interface{}{},
	}
	// Missing status treated as draft → active should work
	if err := Transition(doc, "active"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if doc.Frontmatter["status"] != "active" {
		t.Errorf("status = %v, want %q", doc.Frontmatter["status"], "active")
	}
}

func TestTransitionFromArchived(t *testing.T) {
	doc := &Document{
		Frontmatter: map[string]interface{}{"status": "archived"},
	}
	err := Transition(doc, "active")
	if err == nil {
		t.Fatal("expected error transitioning from archived")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestSpecialCharactersInYAML(t *testing.T) {
	input := "---\ntitle: \"Colons: everywhere\"\ndescription: \"Quotes \\\"inside\\\" here\"\npath: \".thoughts/designs/test.md\"\n---\n# Body\n"
	doc, err := ParseBytes([]byte(input), "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter["title"] != "Colons: everywhere" {
		t.Errorf("title = %v, want %q", doc.Frontmatter["title"], "Colons: everywhere")
	}
	if doc.Frontmatter["path"] != ".thoughts/designs/test.md" {
		t.Errorf("path = %v, want %q", doc.Frontmatter["path"], ".thoughts/designs/test.md")
	}
}

func TestSerializeNoFrontmatter(t *testing.T) {
	doc := &Document{
		Frontmatter: map[string]interface{}{},
		Body:        "# Just body\n",
	}
	data, err := Serialize(doc)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Just body\n" {
		t.Errorf("got %q, want %q", string(data), "# Just body\n")
	}
}
