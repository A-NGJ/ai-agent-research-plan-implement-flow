package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSyncProjectGlobalSkipsProjectBlocks verifies that syncProject with
// global: true installs skills/agents/settings into the target dir but
// does NOT create .rpi/, templates, the rules file, or .gitignore.
func TestSyncProjectGlobalSkipsProjectBlocks(t *testing.T) {
	tmp := t.TempDir()

	cfg, err := resolveTargetConfig("claude")
	if err != nil {
		t.Fatalf("resolveTargetConfig: %v", err)
	}

	// Create the tool subdirs upfront — runInit normally does this, and
	// syncProject expects them to exist for skill/agent install. Mirror
	// the global init code path which skips the "tool dir already
	// exists" guard.
	for _, d := range cfg.subdirs {
		if err := os.MkdirAll(filepath.Join(tmp, cfg.toolDir, d), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := syncProject(syncOptions{
		targetDir: tmp,
		cfg:       cfg,
		global:    true,
		w:         &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("syncProject(global=true) failed: %v", err)
	}

	// Project-side blocks must be skipped.
	for _, missing := range []string{
		".rpi",
		".rpi/templates",
		"CLAUDE.md",
		".gitignore",
	} {
		if _, err := os.Stat(filepath.Join(tmp, missing)); !os.IsNotExist(err) {
			t.Errorf("global mode created %s; expected it to be skipped", missing)
		}
	}

	// Skills install still runs.
	skillPath := filepath.Join(tmp, ".claude", "skills", "rpi-research", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("skills not installed under global mode: %v", err)
	}

	// Agents install still runs (Claude target only).
	agentPath := filepath.Join(tmp, ".claude", "agents", "rpi-verify.md")
	if _, err := os.Stat(agentPath); err != nil {
		t.Errorf("agents not installed under global mode: %v", err)
	}

	// settings.json still gets configured.
	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("settings.json not written under global mode: %v", err)
	}
}

func TestUpdateInsertsContractBlockIntoPreExistingProject(t *testing.T) {
	dir := t.TempDir()

	// Init, then strip the contract block from the rules file to simulate a
	// project that predates the feature.
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	claudeMD := filepath.Join(dir, "CLAUDE.md")
	prior := "# CLAUDE.md\n\nProject overview here.\n\n## Custom Section\n\nUser content.\n"
	if err := os.WriteFile(claudeMD, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	// Update should append the contract block and back up the prior file.
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "<!-- rpi:contract:begin") {
		t.Error("update did not insert contract begin marker")
	}
	if !strings.Contains(content, "<!-- rpi:contract:end -->") {
		t.Error("update did not insert contract end marker")
	}
	if !strings.Contains(content, "## RPI Skill Contract") {
		t.Error("update did not insert '## RPI Skill Contract' heading")
	}
}

func TestUpdateReplacesContractBlockContents(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Hand-craft a rules file with a stale contract block plus user content
	// before and after.
	before := "# CLAUDE.md\n\n## User Before\n\nUser line A.\n\n"
	staleBlock := "<!-- rpi:contract:begin v=1 -->\n## Stale Heading\n\nObsolete text.\n<!-- rpi:contract:end -->\n"
	after := "\n## User After\n\nUser line B.\n"
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudeMD, []byte(before+staleBlock+after), 0644); err != nil {
		t.Fatal(err)
	}

	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, "Stale Heading") {
		t.Error("stale block contents survived update")
	}
	if !strings.Contains(got, "## RPI Skill Contract") {
		t.Error("fresh contract heading missing after update")
	}
	if !strings.HasPrefix(got, before) {
		t.Errorf("user content before block was modified.\nwant prefix: %q\ngot:         %q", before, got)
	}
	if !strings.HasSuffix(got, after) {
		t.Errorf("user content after block was modified.\nwant suffix: %q\ngot:         %q", after, got)
	}
}

func TestUpdateIsIdempotentOnSecondRun(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// First update — establishes baseline.
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("first update failed: %v", err)
	}

	claudeMD := filepath.Join(dir, "CLAUDE.md")
	first, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}

	// Second update — must not mutate CLAUDE.md.
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("second update failed: %v", err)
	}
	second, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Error("second update call mutated CLAUDE.md")
	}
}

func TestUpdateNoClaudeMDSkipsContract(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Replace CLAUDE.md with content that has no contract block.
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	prior := "# CLAUDE.md\n\nHand-rolled, no fences here.\n"
	if err := os.WriteFile(claudeMD, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	resetUpdateFlags()
	updateNoClaudeMD = true
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != prior {
		t.Errorf("--no-claude-md should leave rules file untouched.\nwant: %q\ngot:  %q", prior, data)
	}
	if bytes.Contains(data, []byte("<!-- rpi:contract:begin")) {
		t.Error("--no-claude-md should not insert contract block")
	}
}
