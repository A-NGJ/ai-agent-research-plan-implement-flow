package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetInitFlags() {
	initForce = false
	initNoClaudeMD = false
	initTrackThoughts = false
}

func runInitInDir(t *testing.T, dir string) (*bytes.Buffer, error) {
	t.Helper()
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	return buf, err
}

func TestInitCreatesAllDirs(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .claude/ subdirs
	claudeSubdirs := []string{"agents", "commands", "skills", "hooks"}
	for _, d := range claudeSubdirs {
		path := filepath.Join(dir, ".claude", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".claude/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".claude/%s is not a directory", d)
		}
	}

	// Verify .thoughts/ subdirs
	thoughtsSubdirs := []string{
		"research", "proposals",
		"plans", "specs", "reviews", "archive", "prs",
	}
	for _, d := range thoughtsSubdirs {
		path := filepath.Join(dir, ".thoughts", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".thoughts/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".thoughts/%s is not a directory", d)
		}
	}

	// Verify CLAUDE.md and PIPELINE.md created
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".thoughts", "PIPELINE.md")); err != nil {
		t.Error(".thoughts/PIPELINE.md not created")
	}

	// Verify .gitignore entries
	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), ".claude/") {
		t.Error(".gitignore missing .claude/ entry")
	}
	if !strings.Contains(string(gitignore), ".thoughts/") {
		t.Error(".gitignore missing .thoughts/ entry")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .claude/agents/") {
		t.Error("output missing .claude/agents/ creation message")
	}
	if !strings.Contains(output, "Created .thoughts/research/") {
		t.Error("output missing .thoughts/research/ creation message")
	}
}

func TestInitIdempotent(t *testing.T) {
	dir := t.TempDir()

	// First run
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Second run should error (`.claude/` exists)
	_, err = runInitInDir(t, dir)
	if err == nil {
		t.Fatal("second run should return error")
	}
	if !strings.Contains(err.Error(), ".claude/ already exists") {
		t.Errorf("expected '.claude/ already exists' error, got: %v", err)
	}
}

func TestInitPartial(t *testing.T) {
	dir := t.TempDir()

	// Pre-create some .thoughts/ dirs but not .claude/
	os.MkdirAll(filepath.Join(dir, ".thoughts", "research"), 0755)
	os.MkdirAll(filepath.Join(dir, ".thoughts", "plans"), 0755)

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All dirs should exist
	for _, d := range []string{"proposals", "specs", "reviews", "archive", "prs"} {
		path := filepath.Join(dir, ".thoughts", d)
		if _, err := os.Stat(path); err != nil {
			t.Errorf(".thoughts/%s not created: %v", d, err)
		}
	}
}

func TestInitForce(t *testing.T) {
	dir := t.TempDir()

	// First run
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Write custom CLAUDE.md content
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	original, _ := os.ReadFile(claudeMD)
	os.WriteFile(claudeMD, []byte("custom content"), 0644)

	// Second run with --force
	resetInitFlags()
	initForce = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err = cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--force run error: %v", err)
	}

	// CLAUDE.md should be overwritten with template content
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if string(data) == "custom content" {
		t.Error("--force should have overwritten CLAUDE.md")
	}
	if string(data) != string(original) {
		t.Error("--force should have restored template content")
	}
}

func TestInitNoClaudeMD(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initNoClaudeMD = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CLAUDE.md should not exist
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should not be created with --no-claude-md")
	}

	// PIPELINE.md should still exist
	if _, err := os.Stat(filepath.Join(dir, ".thoughts", "PIPELINE.md")); err != nil {
		t.Error(".thoughts/PIPELINE.md should still be created")
	}
}

func TestInitTrackThoughts(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initTrackThoughts = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if strings.Contains(string(gitignore), ".thoughts/") {
		t.Error(".thoughts/ should not be in .gitignore with --track-thoughts")
	}
	if !strings.Contains(string(gitignore), ".claude/") {
		t.Error(".claude/ should still be in .gitignore")
	}
}

func TestInitTargetDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "myproject")
	os.MkdirAll(target, 0755)

	_, err := runInitInDir(t, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, ".claude", "commands")); err != nil {
		t.Error(".claude/commands not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(target, ".thoughts", "plans")); err != nil {
		t.Error(".thoughts/plans not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not created in target dir")
	}
}

func TestInitExistingClaudeDir(t *testing.T) {
	dir := t.TempDir()

	// Pre-create .claude/
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	_, err := runInitInDir(t, dir)
	if err == nil {
		t.Fatal("should error when .claude/ exists without --force")
	}
	if !strings.Contains(err.Error(), ".claude/ already exists") {
		t.Errorf("expected '.claude/ already exists' error, got: %v", err)
	}
}

func TestInitGitignore(t *testing.T) {
	dir := t.TempDir()

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, ".claude/") {
		t.Error("missing .claude/ entry")
	}
	if !strings.Contains(content, ".thoughts/") {
		t.Error("missing .thoughts/ entry")
	}
}

func TestInitGitignoreIdempotent(t *testing.T) {
	dir := t.TempDir()

	// First run
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Second run with --force
	resetInitFlags()
	initForce = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err = cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--force run error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	content := string(data)
	if strings.Count(content, ".claude/") != 1 {
		t.Errorf(".claude/ appears %d times, want 1", strings.Count(content, ".claude/"))
	}
	if strings.Count(content, ".thoughts/") != 1 {
		t.Errorf(".thoughts/ appears %d times, want 1", strings.Count(content, ".thoughts/"))
	}
}

func TestInitTemplateContent(t *testing.T) {
	dir := t.TempDir()

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify CLAUDE.md has template content
	claudeData, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(claudeData), "# CLAUDE.md") {
		t.Error("CLAUDE.md missing expected template header")
	}

	// Verify PIPELINE.md has template content
	pipelineData, err := os.ReadFile(filepath.Join(dir, ".thoughts", "PIPELINE.md"))
	if err != nil {
		t.Fatalf("read PIPELINE.md: %v", err)
	}
	if !strings.Contains(string(pipelineData), "# Development Pipeline") {
		t.Error("PIPELINE.md missing expected template header")
	}
}

func TestInitInstallsWorkflowFiles(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify embedded agents are installed
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "codebase-analyzer.md")); err != nil {
		t.Error("agents/codebase-analyzer.md not installed")
	}

	// Verify embedded commands are installed
	for _, cmd := range []string{"rpi-plan.md", "rpi-research.md", "rpi-design.md", "rpi-implement.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", cmd)); err != nil {
			t.Errorf("commands/%s not installed", cmd)
		}
	}

	// Verify embedded skills are installed
	for _, skill := range []string{"find-patterns", "analyze-thoughts", "locate-codebase", "locate-thoughts"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf("skills/%s/SKILL.md not installed", skill)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Installed") {
		t.Error("output missing install confirmation")
	}
}

func TestInitDoesNotOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Modify an installed file
	cmdFile := filepath.Join(dir, ".claude", "commands", "rpi-plan.md")
	os.WriteFile(cmdFile, []byte("custom content"), 0644)

	// Second run with --force
	resetInitFlags()
	initForce = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err = cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--force run error: %v", err)
	}

	// File should be overwritten
	data, _ := os.ReadFile(cmdFile)
	if string(data) == "custom content" {
		t.Error("--force should have overwritten customized file")
	}
}

func TestCopyDirectory(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	// Create test files
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(src, "b.txt"), []byte("bbb"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "c.txt"), []byte("ccc"), 0644)

	count, err := copyDirectory(src, dest)
	if err != nil {
		t.Fatalf("copyDirectory error: %v", err)
	}
	if count != 3 { // a.txt, b.txt, sub/
		t.Errorf("expected 3 items copied, got %d", count)
	}

	// Verify content
	data, _ := os.ReadFile(filepath.Join(dest, "a.txt"))
	if string(data) != "aaa" {
		t.Errorf("a.txt content: got %q, want %q", data, "aaa")
	}
	data, _ = os.ReadFile(filepath.Join(dest, "sub", "c.txt"))
	if string(data) != "ccc" {
		t.Errorf("sub/c.txt content: got %q, want %q", data, "ccc")
	}
}
