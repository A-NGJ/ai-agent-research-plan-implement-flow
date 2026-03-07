package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFrontmatterGetAll(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: draft\ntitle: Test Doc\n---\n# Body\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter["status"] != "draft" {
		t.Errorf("status = %v, want %q", doc.Frontmatter["status"], "draft")
	}
	if doc.Frontmatter["title"] != "Test Doc" {
		t.Errorf("title = %v, want %q", doc.Frontmatter["title"], "Test Doc")
	}
}

func TestFrontmatterGetField(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: active\ntitle: Hello\n---\n# Body\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := doc.Frontmatter["status"]
	if !ok {
		t.Fatal("status field not found")
	}
	if val != "active" {
		t.Errorf("status = %v, want %q", val, "active")
	}
}

func TestFrontmatterGetNoFrontmatter(t *testing.T) {
	path := writeTempFile(t, "# Just a heading\n\nNo frontmatter.\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Frontmatter) != 0 {
		t.Errorf("expected empty frontmatter, got %v", doc.Frontmatter)
	}
}

func TestFrontmatterSet(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: draft\n---\n# Body\n\nOriginal content.\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	doc.Frontmatter["new_field"] = "new_value"
	if err := frontmatter.Write(doc); err != nil {
		t.Fatal(err)
	}

	// Re-parse and verify
	doc2, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	if doc2.Frontmatter["new_field"] != "new_value" {
		t.Errorf("new_field = %v, want %q", doc2.Frontmatter["new_field"], "new_value")
	}
	if doc2.Body != doc.Body {
		t.Errorf("body changed after set:\ngot:  %q\nwant: %q", doc2.Body, doc.Body)
	}
}

func TestFrontmatterTransitionValid(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: draft\n---\n# Body\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := frontmatter.Transition(doc, "active"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := frontmatter.Write(doc); err != nil {
		t.Fatal(err)
	}

	// Re-parse and verify
	doc2, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc2.Frontmatter["status"] != "active" {
		t.Errorf("status = %v, want %q", doc2.Frontmatter["status"], "active")
	}
}

func TestFrontmatterTransitionInvalidExitCode(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "rpi")
	build := exec.Command("go", "build", "-o", binPath, "./")
	build.Dir = filepath.Join(".", ".")
	// We need to find the project root for the build
	// Since we're in cmd/rpi/, we build from here
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	tmpFile := writeTempFile(t, "---\nstatus: draft\n---\n# Body\n")

	cmd := exec.Command(binPath, "frontmatter", "transition", tmpFile, "archived")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code for invalid transition")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Errorf("exit code = %d, want 2\noutput: %s", exitErr.ExitCode(), out)
	}

	if !strings.Contains(string(out), "invalid status transition") {
		t.Errorf("stderr should contain 'invalid status transition', got: %s", out)
	}
}
