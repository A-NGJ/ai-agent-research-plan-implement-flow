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
	updateGlobal = false
	updateTarget = "claude"
}

// runUpdateGlobal invokes rpi update --global with HOME redirected.
func runUpdateGlobal(t *testing.T, home, target string) (*bytes.Buffer, error) {
	t.Helper()
	t.Setenv("HOME", home)
	resetUpdateFlags()
	updateGlobal = true
	if target != "" {
		updateTarget = target
	}
	buf := new(bytes.Buffer)
	cmd := updateCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, nil)
	return buf, err
}

func TestUpdateGlobalRefreshesUserInstall(t *testing.T) {
	home := t.TempDir()

	calls, cleanup := stubMCPRunner(t)
	t.Cleanup(cleanup)
	_ = calls

	if _, err := runInitGlobal(t, home, "claude"); err != nil {
		t.Fatalf("rpi init --global: %v", err)
	}

	// Modify a deployed skill.
	skillPath := filepath.Join(home, ".claude", "skills", "rpi-research", "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("custom content"), 0644); err != nil {
		t.Fatalf("modify skill: %v", err)
	}

	if _, err := runUpdateGlobal(t, home, "claude"); err != nil {
		t.Fatalf("rpi update --global: %v", err)
	}

	// Skill restored to embedded version.
	data, _ := os.ReadFile(skillPath)
	if string(data) == "custom content" {
		t.Error("update --global should overwrite modified skill")
	}

	// Backup carries the modified content.
	bakData, err := os.ReadFile(skillPath + ".bak")
	if err != nil {
		t.Fatalf("missing .bak after update --global: %v", err)
	}
	if string(bakData) != "custom content" {
		t.Error(".bak should contain pre-update content")
	}

	// No per-project artifacts at HOME.
	for _, missing := range []string{".rpi", "CLAUDE.md", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(home, missing)); !os.IsNotExist(err) {
			t.Errorf("update --global created %s under HOME; expected absent", missing)
		}
	}
}

func TestUpdateGlobalIdempotent(t *testing.T) {
	home := t.TempDir()
	_, cleanup := stubMCPRunner(t)
	t.Cleanup(cleanup)

	if _, err := runInitGlobal(t, home, "claude"); err != nil {
		t.Fatalf("rpi init --global: %v", err)
	}

	if _, err := runUpdateGlobal(t, home, "claude"); err != nil {
		t.Fatalf("first rpi update --global: %v", err)
	}

	// No .bak should appear in any skill dir after a clean second-run update.
	skillsDir := filepath.Join(home, ".claude", "skills")
	entries, _ := os.ReadDir(skillsDir)
	for _, entry := range entries {
		bakSkill := filepath.Join(skillsDir, entry.Name(), "SKILL.md.bak")
		if _, err := os.Stat(bakSkill); err == nil {
			t.Errorf("idempotent update produced .bak: %s", bakSkill)
		}
	}
}

func TestUpdateGlobalRejectsPositionalDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	resetUpdateFlags()
	updateGlobal = true
	buf := new(bytes.Buffer)
	cmd := updateCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{"./somewhere"})
	if err == nil {
		t.Fatal("expected error for --global with positional dir")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutually exclusive, got: %v", err)
	}
}

func TestUpdateGlobalOpenCodeTarget(t *testing.T) {
	home := t.TempDir()

	if _, err := runInitGlobal(t, home, "opencode"); err != nil {
		t.Fatalf("rpi init --global --target opencode: %v", err)
	}
	if _, err := runUpdateGlobal(t, home, "opencode"); err != nil {
		t.Fatalf("rpi update --global --target opencode: %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".config", "opencode", "skills", "rpi-research", "SKILL.md")); err != nil {
		t.Errorf("opencode skill missing after global update: %v", err)
	}
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

// TestUpdatePreservesModifiedRulesFile verifies that rpi update no longer
// overwrites a user-customized CLAUDE.md wholesale. Per the rpi-skill-contract
// spec, update preserves user content outside the contract fence and only
// refreshes the fenced block in place. The contract block is appended if
// absent — but no .bak is created for non-fenced content.
func TestUpdatePreservesModifiedRulesFile(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Replace CLAUDE.md with hand-rolled content (no contract block).
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	custom := "# CLAUDE.md\n\nMy custom overview.\n\n## Custom Section\n\nUser content.\n"
	if err := os.WriteFile(claudeMD, []byte(custom), 0644); err != nil {
		t.Fatal(err)
	}

	// Update — must preserve the custom content and append the contract block.
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte(custom)) {
		t.Errorf("update should preserve user content at file start.\nwant prefix: %q\ngot:         %q", custom, data)
	}
	if !bytes.Contains(data, []byte("<!-- rpi:contract:begin")) {
		t.Error("update should append contract block to a customized rules file")
	}

	// No .bak should be created for the rules file under the new behavior.
	if _, err := os.Stat(claudeMD + ".bak"); err == nil {
		t.Error("update should not create CLAUDE.md.bak under new fence-preserving behavior")
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

func TestUpdateAddsSafeBashAllowlist(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.WriteFile(settingsPath,
		[]byte(`{"permissions":{"allow":["mcp__rpi__*"]}}`), 0644); err != nil {
		t.Fatalf("rewrite settings.json: %v", err)
	}

	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	content := string(data)
	for _, pattern := range safeBashPatterns {
		if !strings.Contains(content, pattern) {
			t.Errorf("update did not add safe pattern %q", pattern)
		}
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

	// .claude/skills/ should have new skills (11 first-party + grill-me).
	entries, err := os.ReadDir(filepath.Join(dir, ".claude", "skills"))
	if err != nil {
		t.Fatalf("read .claude/skills/: %v", err)
	}
	if len(entries) != 12 {
		t.Errorf("expected 12 skill dirs in .claude/skills/, got %d", len(entries))
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

	// Should still have 12 skills (11 first-party + grill-me).
	entries, err := os.ReadDir(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatalf("read .agents/skills/: %v", err)
	}
	if len(entries) != 12 {
		t.Errorf("expected 12 skill dirs, got %d", len(entries))
	}

	// No tool-specific dirs should be created
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		t.Error(".claude/ should not be created for agents-only update")
	}
}

func TestUpdateWritesGitignorePolicy(t *testing.T) {
	dir := t.TempDir()

	// Init the project, then strip .gitignore to simulate a project that pre-dates
	// the gitignore policy.
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, ".gitignore")); err != nil {
		t.Fatalf("remove .gitignore: %v", err)
	}

	// Update should re-apply the policy.
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, want := range []string{".claude/\n", ".rpi/*\n", "!.rpi/specs/\n"} {
		if !strings.Contains(content, want) {
			t.Errorf(".gitignore missing %q after update; got:\n%s", want, content)
		}
	}
}

func TestUpdateGitignoreIdempotent(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Run update twice — entries should appear exactly once.
	for range 2 {
		resetUpdateFlags()
		buf = new(bytes.Buffer)
		cmd = updateCmd
		cmd.SetOut(buf)
		if err := cmd.RunE(cmd, []string{dir}); err != nil {
			t.Fatalf("update error: %v", err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, entry := range []string{".claude/", ".rpi/*", "!.rpi/specs/"} {
		if got := strings.Count(content, entry+"\n"); got != 1 {
			t.Errorf(".gitignore entry %q appears %d times, want 1; content:\n%s", entry, got, content)
		}
	}
}
