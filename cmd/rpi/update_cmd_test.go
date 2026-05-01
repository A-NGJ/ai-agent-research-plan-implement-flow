package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetUpdateFlags() {
	updateNoClaudeMD = false
}

func TestUpdateRequiresExistingProject(t *testing.T) {
	dir := t.TempDir()

	resetUpdateFlags()
	buf := new(bytes.Buffer)
	cmd := updateCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("update should error without prior init")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected 'not initialized' error, got: %v", err)
	}
}

func TestUpdateCreatesMissingDirs(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove some dirs
	os.Remove(filepath.Join(dir, ".rpi", "reviews"))
	os.Remove(filepath.Join(dir, ".claude", "hooks"))

	// Update
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Verify dirs recreated
	if _, err := os.Stat(filepath.Join(dir, ".rpi", "reviews")); err != nil {
		t.Error(".rpi/reviews not recreated")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "hooks")); err != nil {
		t.Error(".claude/hooks not recreated")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .rpi/reviews/") {
		t.Error("output missing .rpi/reviews/ creation message")
	}
	if !strings.Contains(output, "Created .claude/hooks/") {
		t.Error("output missing .claude/hooks/ creation message")
	}
}

func TestUpdateBacksUpModifiedFiles(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify a skill file
	skillFile := filepath.Join(dir, ".claude", "skills", "rpi-plan", "SKILL.md")
	os.WriteFile(skillFile, []byte("custom content"), 0644)

	// Update — should overwrite and create backup
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// File should be overwritten
	data, _ := os.ReadFile(skillFile)
	if string(data) == "custom content" {
		t.Error("update should overwrite modified files")
	}

	// Backup should exist with original content
	bakData, err := os.ReadFile(skillFile + ".bak")
	if err != nil {
		t.Fatal("backup file not created")
	}
	if string(bakData) != "custom content" {
		t.Error("backup should contain original custom content")
	}
}

func TestUpdateSkipsIdenticalFiles(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Update with no modifications — should not create backups
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// No .bak files should exist
	skillBak := filepath.Join(dir, ".claude", "skills", "rpi-plan", "SKILL.md.bak")
	if _, err := os.Stat(skillBak); err == nil {
		t.Error("no backup should be created for identical files")
	}
}

func TestUpdateBacksUpModifiedRulesFile(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(claudeMD, []byte("custom content"), 0644)

	// Update — should overwrite and create backup
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// CLAUDE.md should be overwritten with template
	data, _ := os.ReadFile(claudeMD)
	if string(data) == "custom content" {
		t.Error("update should overwrite CLAUDE.md with template")
	}
	if !strings.Contains(string(data), "# CLAUDE.md") {
		t.Error("CLAUDE.md missing template header after update")
	}

	// Backup should exist
	bakData, err := os.ReadFile(claudeMD + ".bak")
	if err != nil {
		t.Fatal("CLAUDE.md.bak not created")
	}
	if string(bakData) != "custom content" {
		t.Error("backup should contain original custom content")
	}
}

func TestUpdateNoClaudeMDSkipsRulesFile(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(claudeMD, []byte("custom content"), 0644)

	// Update with --no-claude-md
	resetUpdateFlags()
	updateNoClaudeMD = true
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// CLAUDE.md should be untouched
	data, _ := os.ReadFile(claudeMD)
	if string(data) != "custom content" {
		t.Error("update --no-claude-md should not modify CLAUDE.md")
	}
}

func TestUpdateAddsSettingsPermission(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove the permission from settings.json
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	os.WriteFile(settingsPath, []byte(`{"permissions":{"allow":[]}}`), 0644)

	// Update should re-add it
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(data), "mcp__rpi__*") {
		t.Error("update should add mcp__rpi__* permission to settings.json")
	}
}

func TestUpdateDetectsOpenCodeTarget(t *testing.T) {
	dir := t.TempDir()

	// Init with opencode
	resetInitFlags()
	initTarget = "opencode"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove a subdir
	os.Remove(filepath.Join(dir, ".opencode", "hooks"))

	// Update (should auto-detect opencode)
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "hooks")); err != nil {
		t.Error(".opencode/hooks not recreated")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .opencode/hooks/") {
		t.Error("output missing .opencode/hooks/ creation message")
	}
}

// TC-6: Update with existing commands dir
func TestUpdatePreservesExistingCommandsDir(t *testing.T) {
	dir := t.TempDir()

	// Manually set up an old-style project with .claude/commands/
	os.MkdirAll(filepath.Join(dir, ".claude", "skills"), 0755)
	os.MkdirAll(filepath.Join(dir, ".claude", "hooks"), 0755)
	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0755)
	os.WriteFile(filepath.Join(dir, ".claude", "commands", "rpi-propose.md"), []byte("old command"), 0644)
	os.MkdirAll(filepath.Join(dir, ".rpi", "plans"), 0755)

	// Run update
	resetUpdateFlags()
	buf := new(bytes.Buffer)
	cmd := updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// AS-13: .claude/commands/ must not be deleted
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "rpi-propose.md"))
	if err != nil {
		t.Error(".claude/commands/rpi-propose.md was deleted (AS-13)")
	} else if string(data) != "old command" {
		t.Error(".claude/commands/rpi-propose.md was modified (AS-13)")
	}

	// .claude/skills/ should have new skills (10 first-party + grill-me).
	entries, err := os.ReadDir(filepath.Join(dir, ".claude", "skills"))
	if err != nil {
		t.Fatalf("read .claude/skills/: %v", err)
	}
	if len(entries) != 11 {
		t.Errorf("expected 11 skill dirs in .claude/skills/, got %d", len(entries))
	}

	// Bundled third-party LICENSE files survive update.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", "grill-me", "LICENSE")); err != nil {
		t.Errorf("grill-me/LICENSE not deployed by update: %v", err)
	}
}

func TestUpdateSyncsAgents(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Delete an agent file
	os.Remove(filepath.Join(dir, ".claude", "agents", "rpi-verify.md"))

	// Update should restore it
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "rpi-verify.md")); err != nil {
		t.Error("rpi-verify.md not restored by update")
	}
}

func TestUpdateAgentsOnlyNoAgents(t *testing.T) {
	dir := t.TempDir()

	// Init with agents-only
	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Update
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// No agent definitions should be created
	agentsDir := filepath.Join(dir, ".agents", "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		t.Error("agents-only update should not create agent definitions")
	}
}

func TestUpdateDetectsAgentsOnlyTarget(t *testing.T) {
	dir := t.TempDir()

	// Init with agents-only
	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Update (should auto-detect agents-only)
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Should still have 11 skills (10 first-party + grill-me).
	entries, err := os.ReadDir(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatalf("read .agents/skills/: %v", err)
	}
	if len(entries) != 11 {
		t.Errorf("expected 11 skill dirs, got %d", len(entries))
	}

	// No tool-specific dirs should be created
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		t.Error(".claude/ should not be created for agents-only update")
	}
}
