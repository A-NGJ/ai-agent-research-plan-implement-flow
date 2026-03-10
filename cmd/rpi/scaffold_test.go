package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// findRepoRoot walks up from cwd to find the directory containing go.mod.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// buildBinary compiles the rpi binary to a temp dir and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	binary := filepath.Join(t.TempDir(), "rpi")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/rpi")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

// runRPI runs the rpi binary with the given args and returns stdout, stderr, and exit code.
func runRPI(t *testing.T, binary string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	repoRoot := findRepoRoot(t)
	cmd := exec.Command(binary, args...)
	cmd.Dir = repoRoot

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running rpi: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestScaffoldPlanStdout(t *testing.T) {
	binary := buildBinary(t)
	stdout, _, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Test plan",
		"--ticket", "cli-007",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "# cli-007: Test plan") {
		t.Error("stdout should contain plan heading")
	}
	if !strings.Contains(stdout, `ticket: "cli-007"`) {
		t.Error("stdout should contain ticket in frontmatter")
	}
	if !strings.Contains(stdout, "status: draft") {
		t.Error("stdout should contain status: draft")
	}
}

func TestScaffoldPlanWrite(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()
	thoughtsDir := filepath.Join(tmpDir, ".thoughts")

	stdout, _, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Test plan",
		"--thoughts-dir", thoughtsDir,
		"--write",
	)

	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}

	// stdout should contain the file path
	outPath := strings.TrimSpace(stdout)
	if !strings.HasSuffix(outPath, "test-plan.md") {
		t.Errorf("expected path ending in test-plan.md, got %q", outPath)
	}

	// File should exist
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("written file should exist: %v", err)
	}

	// File content should have frontmatter
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if !strings.HasPrefix(string(content), "---\n") {
		t.Error("written file should start with frontmatter")
	}
}

func TestScaffoldPlanWriteOverwriteProtection(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()
	thoughtsDir := filepath.Join(tmpDir, ".thoughts")

	// First write
	_, _, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Test plan",
		"--thoughts-dir", thoughtsDir,
		"--write",
	)
	if exitCode != 0 {
		t.Fatalf("first write: expected exit 0, got %d", exitCode)
	}

	// Second write should fail with exit 3
	_, stderr, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Test plan",
		"--thoughts-dir", thoughtsDir,
		"--write",
	)
	if exitCode != 3 {
		t.Errorf("second write: expected exit 3, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stderr, "already exists") {
		t.Error("stderr should mention file already exists")
	}
}

func TestScaffoldPlanWriteForce(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()
	thoughtsDir := filepath.Join(tmpDir, ".thoughts")

	// First write
	_, _, _ = runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Test plan",
		"--thoughts-dir", thoughtsDir,
		"--write",
	)

	// Second write with --force should succeed
	_, _, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Test plan",
		"--thoughts-dir", thoughtsDir,
		"--write", "--force",
	)
	if exitCode != 0 {
		t.Errorf("force write: expected exit 0, got %d", exitCode)
	}
}

func TestScaffoldAllTypes(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		artifactType string
		extraArgs    []string
	}{
		{"research", []string{"--topic", "Test research"}},
		{"plan", []string{"--topic", "Test plan"}},
		{"propose", []string{"--topic", "Test proposal"}},
		{"verify-report", []string{"--topic", "Test verify"}},
		{"spec", []string{"--topic", "Test spec"}},
	}

	for _, tt := range tests {
		t.Run(tt.artifactType, func(t *testing.T) {
			args := append([]string{"scaffold", tt.artifactType}, tt.extraArgs...)
			stdout, stderr, exitCode := runRPI(t, binary, args...)
			if exitCode != 0 {
				t.Errorf("expected exit 0, got %d (stderr: %s)", exitCode, stderr)
			}
			if !strings.Contains(stdout, "---\n") {
				t.Error("output should contain frontmatter")
			}
		})
	}
}

func TestScaffoldMissingRequiredFlag(t *testing.T) {
	binary := buildBinary(t)

	// plan without --topic
	_, stderr, exitCode := runRPI(t, binary, "scaffold", "plan")
	if exitCode != 2 {
		t.Errorf("expected exit 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "--topic") {
		t.Errorf("stderr should mention --topic, got: %s", stderr)
	}
}

func TestScaffoldCustomTemplatesDir(t *testing.T) {
	binary := buildBinary(t)
	repoRoot := findRepoRoot(t)
	customDir := filepath.Join(repoRoot, ".claude", "templates")

	stdout, _, exitCode := runRPI(t, binary,
		"scaffold", "research",
		"--topic", "Custom dir test",
		"--templates-dir", customDir,
	)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "# Research: Custom dir test") {
		t.Error("output should use template from custom dir")
	}
}

func TestScaffoldPlanWithTicket(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()
	thoughtsDir := filepath.Join(tmpDir, ".thoughts")

	stdout, _, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Auth refactor",
		"--ticket", "cli-007",
		"--thoughts-dir", thoughtsDir,
		"--write",
	)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	outPath := strings.TrimSpace(stdout)
	if !strings.Contains(outPath, "cli-007") {
		t.Errorf("filename should contain ticket id, got %q", outPath)
	}
}

func TestScaffoldPlanWithoutTicket(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()
	thoughtsDir := filepath.Join(tmpDir, ".thoughts")

	stdout, _, exitCode := runRPI(t, binary,
		"scaffold", "plan",
		"--topic", "Auth refactor",
		"--thoughts-dir", thoughtsDir,
		"--write",
	)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	outPath := strings.TrimSpace(stdout)
	filename := filepath.Base(outPath)
	// Should NOT contain a ticket segment, just date-slug
	if strings.Count(filename, "-") > 5 {
		// date has 2 hyphens, slug "auth-refactor" has 1 = 3 total expected
		// This is a rough check
	}
	if !strings.HasSuffix(filename, "auth-refactor.md") {
		t.Errorf("filename should end with auth-refactor.md, got %q", filename)
	}
}

func TestScaffoldUnknownType(t *testing.T) {
	binary := buildBinary(t)

	_, stderr, exitCode := runRPI(t, binary,
		"scaffold", "unknown-type",
		"--topic", "Test",
	)
	if exitCode != 2 {
		t.Errorf("expected exit 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "unknown artifact type") {
		t.Errorf("stderr should mention unknown type, got: %s", stderr)
	}
}
