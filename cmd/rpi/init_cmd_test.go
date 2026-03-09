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
	initAll = false
	initUpdate = false
	initAgentsOnly = false
	initCommandsOnly = false
	initSkillsOnly = false
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
		"research", "designs", "structures", "tickets",
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
	if !strings.Contains(string(gitignore), ".claude/settings.local.json") {
		t.Error(".gitignore missing .claude/settings.local.json entry")
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
	for _, d := range []string{"designs", "structures", "tickets", "specs", "reviews", "archive", "prs"} {
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
	if !strings.Contains(string(gitignore), ".claude/settings.local.json") {
		t.Error(".claude/settings.local.json should still be in .gitignore")
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
	if !strings.Contains(content, ".claude/settings.local.json") {
		t.Error("missing .claude/settings.local.json entry")
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
	if strings.Count(content, ".claude/settings.local.json") != 1 {
		t.Errorf(".claude/settings.local.json appears %d times, want 1", strings.Count(content, ".claude/settings.local.json"))
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

// setupDotfiles creates a fake dotfiles directory with test files.
func setupDotfiles(t *testing.T) string {
	t.Helper()
	dotfiles := t.TempDir()
	for _, dir := range []string{"agents", "commands", "skills", "hooks"} {
		os.MkdirAll(filepath.Join(dotfiles, dir), 0755)
		os.WriteFile(filepath.Join(dotfiles, dir, "test.md"), []byte(dir+" content"), 0644)
	}
	// Add a subdirectory in agents to test recursive copy
	os.MkdirAll(filepath.Join(dotfiles, "agents", "subdir"), 0755)
	os.WriteFile(filepath.Join(dotfiles, "agents", "subdir", "nested.md"), []byte("nested"), 0644)
	return dotfiles
}

func TestInitAll(t *testing.T) {
	dotfiles := setupDotfiles(t)
	t.Setenv("DOTFILES_CLAUDE", dotfiles)

	dir := t.TempDir()
	resetInitFlags()
	initAll = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all components copied
	for _, comp := range []string{"agents", "commands", "skills", "hooks"} {
		path := filepath.Join(dir, ".claude", comp, "test.md")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s/test.md not copied: %v", comp, err)
			continue
		}
		if string(data) != comp+" content" {
			t.Errorf("%s/test.md has wrong content: %s", comp, data)
		}
	}

	// Verify recursive copy
	nested := filepath.Join(dir, ".claude", "agents", "subdir", "nested.md")
	data, err := os.ReadFile(nested)
	if err != nil {
		t.Fatalf("nested file not copied: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("nested file has wrong content: %s", data)
	}

	output := buf.String()
	if !strings.Contains(output, "Copied") {
		t.Error("output missing copy confirmation")
	}
}

func TestInitAllAgentsOnly(t *testing.T) {
	dotfiles := setupDotfiles(t)
	t.Setenv("DOTFILES_CLAUDE", dotfiles)

	dir := t.TempDir()
	resetInitFlags()
	initAll = true
	initAgentsOnly = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// agents should be copied
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "test.md")); err != nil {
		t.Error("agents/test.md should be copied")
	}

	// commands, skills, hooks should NOT have test.md from dotfiles
	for _, comp := range []string{"commands", "skills", "hooks"} {
		path := filepath.Join(dir, ".claude", comp, "test.md")
		if _, err := os.Stat(path); err == nil {
			t.Errorf("%s/test.md should not be copied with --agents-only", comp)
		}
	}
}

func TestInitAllCommandsOnly(t *testing.T) {
	dotfiles := setupDotfiles(t)
	t.Setenv("DOTFILES_CLAUDE", dotfiles)

	dir := t.TempDir()
	resetInitFlags()
	initAll = true
	initCommandsOnly = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "test.md")); err != nil {
		t.Error("commands/test.md should be copied")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "test.md")); err == nil {
		t.Error("agents/test.md should not be copied with --commands-only")
	}
}

func TestInitAllSkillsOnly(t *testing.T) {
	dotfiles := setupDotfiles(t)
	t.Setenv("DOTFILES_CLAUDE", dotfiles)

	dir := t.TempDir()
	resetInitFlags()
	initAll = true
	initSkillsOnly = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", "test.md")); err != nil {
		t.Error("skills/test.md should be copied")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "test.md")); err == nil {
		t.Error("agents/test.md should not be copied with --skills-only")
	}
}

func TestInitAllDotfilesEnv(t *testing.T) {
	dotfiles := setupDotfiles(t)
	t.Setenv("DOTFILES_CLAUDE", dotfiles)

	dir := t.TempDir()
	resetInitFlags()
	initAll = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it used the env var path (files exist)
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "test.md")); err != nil {
		t.Error("DOTFILES_CLAUDE env var not respected")
	}
}

func TestInitAllMissingSource(t *testing.T) {
	t.Setenv("DOTFILES_CLAUDE", "/nonexistent/path")

	dir := t.TempDir()
	resetInitFlags()
	initAll = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("expected error for missing dotfiles source")
	}
	if !strings.Contains(err.Error(), "dotfiles source not found") {
		t.Errorf("expected 'dotfiles source not found' error, got: %v", err)
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

// initAndSetupForUpdate initializes a project with --all, then returns the dir
// for subsequent --update tests.
func initAndSetupForUpdate(t *testing.T, dotfiles string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("DOTFILES_CLAUDE", dotfiles)

	resetInitFlags()
	initAll = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init --all failed: %v", err)
	}
	return dir
}

func TestInitUpdate(t *testing.T) {
	dotfiles := setupDotfiles(t)
	dir := initAndSetupForUpdate(t, dotfiles)

	// Add a new file to dotfiles
	os.WriteFile(filepath.Join(dotfiles, "agents", "new-agent.md"), []byte("new agent"), 0644)

	resetInitFlags()
	initUpdate = true
	t.Setenv("DOTFILES_CLAUDE", dotfiles)
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--update error: %v", err)
	}

	// New file should be copied
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "agents", "new-agent.md"))
	if err != nil {
		t.Fatal("new-agent.md not copied during update")
	}
	if string(data) != "new agent" {
		t.Errorf("new-agent.md wrong content: %s", data)
	}

	// Existing unchanged file should still be there
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "test.md")); err != nil {
		t.Error("existing test.md should still exist")
	}

	output := buf.String()
	if !strings.Contains(output, "agents:") {
		t.Error("output should contain update stats for agents")
	}
}

func TestInitUpdateDiffers(t *testing.T) {
	dotfiles := setupDotfiles(t)
	dir := initAndSetupForUpdate(t, dotfiles)

	// Modify local file to differ from dotfiles
	os.WriteFile(filepath.Join(dir, ".claude", "agents", "test.md"), []byte("local changes"), 0644)

	resetInitFlags()
	initUpdate = true
	t.Setenv("DOTFILES_CLAUDE", dotfiles)
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--update error: %v", err)
	}

	// File should NOT be overwritten
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "test.md"))
	if string(data) != "local changes" {
		t.Error("differing file should not be overwritten without --force")
	}

	output := buf.String()
	if !strings.Contains(output, "Skipped (differs)") {
		t.Error("should warn about differing file")
	}
}

func TestInitUpdateForce(t *testing.T) {
	dotfiles := setupDotfiles(t)
	dir := initAndSetupForUpdate(t, dotfiles)

	// Modify local file
	os.WriteFile(filepath.Join(dir, ".claude", "agents", "test.md"), []byte("local changes"), 0644)

	resetInitFlags()
	initUpdate = true
	initForce = true
	t.Setenv("DOTFILES_CLAUDE", dotfiles)
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--update --force error: %v", err)
	}

	// File SHOULD be overwritten
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "test.md"))
	if string(data) != "agents content" {
		t.Errorf("--force should overwrite, got: %s", data)
	}
}

func TestInitUpdateFilters(t *testing.T) {
	dotfiles := setupDotfiles(t)
	dir := initAndSetupForUpdate(t, dotfiles)

	// Add new files to agents and commands
	os.WriteFile(filepath.Join(dotfiles, "agents", "new.md"), []byte("new"), 0644)
	os.WriteFile(filepath.Join(dotfiles, "commands", "new.md"), []byte("new"), 0644)

	resetInitFlags()
	initUpdate = true
	initAgentsOnly = true
	t.Setenv("DOTFILES_CLAUDE", dotfiles)
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("--update --agents-only error: %v", err)
	}

	// agents/new.md should be copied
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "new.md")); err != nil {
		t.Error("agents/new.md should be copied with --agents-only")
	}

	// commands/new.md should NOT be copied
	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "new.md")); err == nil {
		t.Error("commands/new.md should not be copied with --agents-only")
	}
}

func TestInitUpdateNoClaude(t *testing.T) {
	t.Setenv("DOTFILES_CLAUDE", t.TempDir())
	dir := t.TempDir()
	// Don't init — no .claude/ dir

	resetInitFlags()
	initUpdate = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("expected error when .claude/ doesn't exist")
	}
	if !strings.Contains(err.Error(), ".claude/ doesn't exist") {
		t.Errorf("expected '.claude/ doesn't exist' error, got: %v", err)
	}
}

func TestUpdateClaudeMD(t *testing.T) {
	dir := t.TempDir()

	// Create a CLAUDE.md with some sections removed
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(claudeMD, []byte("# CLAUDE.md\n\n## Project Overview\n\nMy project.\n\n## Git Workflow\n\nCustom workflow.\n"), 0644)

	buf := new(bytes.Buffer)
	err := updateClaudeMD(buf, dir)
	if err != nil {
		t.Fatalf("updateClaudeMD error: %v", err)
	}

	data, _ := os.ReadFile(claudeMD)
	content := string(data)

	// Should still have original sections
	if !strings.Contains(content, "My project.") {
		t.Error("original content should be preserved")
	}

	// Should have added missing sections from template
	output := buf.String()
	if !strings.Contains(output, "added section") {
		t.Error("should report added sections")
	}

	// The file should now have more content than before
	if !strings.Contains(content, "## Thoughts Directory") {
		t.Error("missing template section should be appended")
	}
}

func TestUpdateClaudeMDUpToDate(t *testing.T) {
	dir := t.TempDir()

	// First init to get full CLAUDE.md
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init error: %v", err)
	}

	// Now run updateClaudeMD — should be up to date
	buf.Reset()
	err := updateClaudeMD(buf, dir)
	if err != nil {
		t.Fatalf("updateClaudeMD error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "up to date") {
		t.Errorf("expected 'up to date' message, got: %s", output)
	}
}

func TestCopyWithUpdate(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create source files
	os.WriteFile(filepath.Join(src, "new.txt"), []byte("new"), 0644)
	os.WriteFile(filepath.Join(src, "same.txt"), []byte("same"), 0644)
	os.WriteFile(filepath.Join(src, "diff.txt"), []byte("source version"), 0644)

	// Create dest files (same and diff)
	os.WriteFile(filepath.Join(dest, "same.txt"), []byte("same"), 0644)
	os.WriteFile(filepath.Join(dest, "diff.txt"), []byte("local version"), 0644)

	buf := new(bytes.Buffer)
	stats, err := copyWithUpdate(buf, src, dest, false)
	if err != nil {
		t.Fatalf("copyWithUpdate error: %v", err)
	}

	if stats.copied != 1 {
		t.Errorf("expected 1 copied, got %d", stats.copied)
	}
	if stats.skipped != 2 { // same.txt + diff.txt (not forced)
		t.Errorf("expected 2 skipped, got %d", stats.skipped)
	}
	if stats.updated != 0 {
		t.Errorf("expected 0 updated, got %d", stats.updated)
	}

	// new.txt should exist
	data, _ := os.ReadFile(filepath.Join(dest, "new.txt"))
	if string(data) != "new" {
		t.Errorf("new.txt: got %q", data)
	}

	// diff.txt should be unchanged (no force)
	data, _ = os.ReadFile(filepath.Join(dest, "diff.txt"))
	if string(data) != "local version" {
		t.Error("diff.txt should not be overwritten without force")
	}

	// Now with force
	buf.Reset()
	stats, err = copyWithUpdate(buf, src, dest, true)
	if err != nil {
		t.Fatalf("copyWithUpdate force error: %v", err)
	}
	if stats.updated != 1 {
		t.Errorf("expected 1 updated with force, got %d", stats.updated)
	}
	data, _ = os.ReadFile(filepath.Join(dest, "diff.txt"))
	if string(data) != "source version" {
		t.Error("diff.txt should be overwritten with force")
	}
}
