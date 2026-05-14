package workflow

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// --- Agent Skills format compliance tests ---

func TestAllSkillsPresent(t *testing.T) {
	// AS-12: All 12 skills must be present (11 first-party pipeline skills +
	// 1 bundled third-party skill: grill-me from mattpocock/skills under MIT).
	expected := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
		"rpi-spec-sync", "rpi-handoff",
		"grill-me",
	}

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	found := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			found[e.Name()] = true
		}
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("missing skill directory: %s", name)
		}
	}
	if len(found) != len(expected) {
		t.Errorf("expected %d skills, found %d", len(expected), len(found))
	}
}

func TestSkillNameMatchesDir(t *testing.T) {
	// AS-7: name field must match parent directory name
	// AS-8: name must match regex and ≤64 chars
	nameRegex := regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		data, err := assets.ReadFile("assets/skills/" + dirName + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s/SKILL.md: %v", dirName, err)
			continue
		}

		name := extractFrontmatterField(string(data), "name")
		if name != dirName {
			t.Errorf("skill %s: name field %q doesn't match directory", dirName, name)
		}
		if len(name) > 64 {
			t.Errorf("skill %s: name exceeds 64 chars (%d)", dirName, len(name))
		}
		if !nameRegex.MatchString(name) {
			t.Errorf("skill %s: name %q doesn't match naming regex", dirName, name)
		}
	}
}

func TestCanonicalSkillsHaveNoToolFields(t *testing.T) {
	// AS-10: canonical SKILL.md must not contain tool-specific fields
	toolFields := []string{"model:", "disable-model-invocation:", "tools:", "allowed-tools:", "context:"}

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := assets.ReadFile("assets/skills/" + e.Name() + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s/SKILL.md: %v", e.Name(), err)
			continue
		}
		fm := extractFrontmatter(string(data))
		for _, field := range toolFields {
			if strings.Contains(fm, field) {
				t.Errorf("skill %s: canonical SKILL.md contains tool-specific field %s", e.Name(), field)
			}
		}
	}
}

func TestSkillDescriptionValid(t *testing.T) {
	// AS-9: description must be 1-1024 chars
	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := assets.ReadFile("assets/skills/" + e.Name() + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s/SKILL.md: %v", e.Name(), err)
			continue
		}
		desc := extractFrontmatterField(string(data), "description")
		if len(desc) == 0 {
			t.Errorf("skill %s: description is empty", e.Name())
		}
		if len(desc) > 1024 {
			t.Errorf("skill %s: description exceeds 1024 chars (%d)", e.Name(), len(desc))
		}
	}
}

// --- InstallSkills tests ---

func TestInstallSkills(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")

	count, _, err := InstallSkills(skillsDir)
	if err != nil {
		t.Fatalf("InstallSkills error: %v", err)
	}
	// 11 first-party SKILL.md + grill-me's SKILL.md + grill-me's LICENSE = 13.
	if count != 13 {
		t.Errorf("expected 13 files installed, got %d", count)
	}

	// Verify all 12 skill dirs exist (11 first-party + grill-me).
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	if len(entries) != 12 {
		t.Errorf("expected 12 skill dirs, got %d", len(entries))
	}

	// Verify deployed files have no tool-specific fields.
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(skillsDir, e.Name(), "SKILL.md"))
		if err != nil {
			t.Errorf("read %s: %v", e.Name(), err)
			continue
		}
		fm := extractFrontmatter(string(data))
		for _, field := range []string{"model:", "disable-model-invocation:", "allowed-tools:", "context:", "tools:"} {
			if strings.Contains(fm, field) {
				t.Errorf("deployed %s contains %s field", e.Name(), field)
			}
		}
	}
}

func TestInstallSkills_BacksUpModifiedFiles(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")

	// First install
	_, _, err := InstallSkills(skillsDir)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Modify a file
	modPath := filepath.Join(skillsDir, "rpi-research", "SKILL.md")
	if err := os.WriteFile(modPath, []byte("custom content"), 0644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	// Second install — should overwrite and create backup
	installed, backedUp, err := InstallSkills(skillsDir)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if installed != 1 {
		t.Errorf("expected 1 file installed (modified one), got %d", installed)
	}
	if backedUp != 1 {
		t.Errorf("expected 1 backup, got %d", backedUp)
	}

	// Original content should be in .bak
	bakData, err := os.ReadFile(modPath + ".bak")
	if err != nil {
		t.Fatal("backup file not created")
	}
	if string(bakData) != "custom content" {
		t.Error("backup should contain original custom content")
	}

	// File should now have embedded content
	data, _ := os.ReadFile(modPath)
	if string(data) == "custom content" {
		t.Error("file should be overwritten with embedded version")
	}
}

func TestInstallSkills_SkipsIdenticalFiles(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")

	// First install
	_, _, err := InstallSkills(skillsDir)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install with no modifications — should skip all
	installed, backedUp, err := InstallSkills(skillsDir)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if installed != 0 {
		t.Errorf("expected 0 files installed (all identical), got %d", installed)
	}
	if backedUp != 0 {
		t.Errorf("expected 0 backups (all identical), got %d", backedUp)
	}

	// No .bak files should exist
	bakPath := filepath.Join(skillsDir, "rpi-research", "SKILL.md.bak")
	if _, err := os.Stat(bakPath); err == nil {
		t.Error("no backup should be created for identical files")
	}
}

func TestInstallSkills_CopiesSiblingFiles(t *testing.T) {
	// Bundled third-party skills (e.g. grill-me) ship a LICENSE file alongside
	// SKILL.md. InstallSkills must copy every regular file in each skill source
	// dir so the upstream attribution travels with each deployed copy.
	skillsDir := filepath.Join(t.TempDir(), "skills")
	if _, _, err := InstallSkills(skillsDir); err != nil {
		t.Fatalf("InstallSkills: %v", err)
	}

	licensePath := filepath.Join(skillsDir, "grill-me", "LICENSE")
	data, err := os.ReadFile(licensePath)
	if err != nil {
		t.Fatalf("grill-me/LICENSE not deployed: %v", err)
	}
	if !strings.Contains(string(data), "Matt Pocock") {
		t.Error("grill-me/LICENSE missing upstream attribution")
	}
	if !strings.Contains(string(data), "MIT License") {
		t.Error("grill-me/LICENSE missing MIT notice")
	}

	// Skills with no sibling files should not get extras (regression check).
	researchEntries, err := os.ReadDir(filepath.Join(skillsDir, "rpi-research"))
	if err != nil {
		t.Fatalf("read rpi-research dir: %v", err)
	}
	if len(researchEntries) != 1 {
		names := make([]string, 0, len(researchEntries))
		for _, e := range researchEntries {
			names = append(names, e.Name())
		}
		t.Errorf("rpi-research should only contain SKILL.md, got %d entries: %v", len(researchEntries), names)
	}
}

// --- Prompt structure tests ---

func TestPromptStructure_HasRequiredSections(t *testing.T) {
	pipelineSkills := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
		"rpi-spec-sync",
	}

	for _, skill := range pipelineSkills {
		data, err := ReadAsset("skills/" + skill + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s: %v", skill, err)
		}
		content := string(data)
		for _, section := range []string{"## Goal", "## Invariants", "## Principles"} {
			if !strings.Contains(content, section) {
				t.Errorf("%s missing required section %q", skill, section)
			}
		}
	}
}

func TestPromptStructure_LineCount(t *testing.T) {
	pipelineSkills := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
		"rpi-spec-sync",
	}

	for _, skill := range pipelineSkills {
		data, err := ReadAsset("skills/" + skill + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s: %v", skill, err)
		}
		body := extractBody(string(data))
		lines := strings.Count(strings.TrimSpace(body), "\n") + 1
		if lines > 50 {
			t.Errorf("%s has %d lines (excluding frontmatter), max is 50", skill, lines)
		}
	}
}

func TestPromptStructure_NoToolReferences(t *testing.T) {
	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := assets.ReadFile("assets/skills/" + e.Name() + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s: %v", e.Name(), err)
			continue
		}
		content := stripBootstrapPreamble(string(data))
		if strings.Contains(content, "rpi_") {
			t.Errorf("%s contains MCP tool name reference (rpi_)", e.Name())
		}
		if strings.Contains(content, "`rpi ") {
			t.Errorf("%s contains backtick-quoted rpi CLI invocation", e.Name())
		}
	}
}

// TestRPISkillsContainBootstrapPreamble enforces that every rpi-* skill
// invokes rpi bootstrap from its prompt body so a global install can
// auto-init project artifacts on first use. grill-me is third-party and
// must remain untouched.
func TestRPISkillsContainBootstrapPreamble(t *testing.T) {
	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	const marker = "rpi bootstrap"
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		data, err := assets.ReadFile("assets/skills/" + name + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		body := extractBody(string(data))
		bodyLines := strings.Split(body, "\n")
		// Look in the first 5 non-empty body lines so the preamble is
		// surfaced before any other skill instructions.
		var early []string
		for _, line := range bodyLines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			early = append(early, trimmed)
			if len(early) == 5 {
				break
			}
		}
		joined := strings.Join(early, "\n")
		hasMarker := strings.Contains(joined, marker)

		if strings.HasPrefix(name, "rpi-") {
			if !hasMarker {
				t.Errorf("%s missing rpi bootstrap preamble in first 5 body lines:\n%s", name, joined)
			}
		} else {
			// Non-rpi-* skills (e.g. grill-me) must not pick up the
			// preamble accidentally.
			if hasMarker {
				t.Errorf("%s should not contain rpi bootstrap preamble (third-party skill)", name)
			}
		}
	}
}

// stripBootstrapPreamble removes the bootstrap preamble line so the
// no-tool-reference policy can stay strict for the rest of every skill
// prompt. The preamble is the one explicitly permitted CLI invocation.
func stripBootstrapPreamble(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "> Before doing anything else, run `rpi bootstrap`") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// --- InstallAgents tests ---

func TestInstallAgents_InstallsAllAgents(t *testing.T) {
	agentsDir := filepath.Join(t.TempDir(), ".claude", "agents")

	count, _, err := InstallAgents(agentsDir)
	if err != nil {
		t.Fatalf("InstallAgents error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 agent file installed, got %d", count)
	}

	// Verify agent file exists
	path := filepath.Join(agentsDir, "rpi-verify.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("agent file rpi-verify.md not installed: %v", err)
	}
}

func TestInstallAgents_BacksUpModifiedFiles(t *testing.T) {
	agentsDir := filepath.Join(t.TempDir(), "agents")

	// First install
	_, _, err := InstallAgents(agentsDir)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Modify a file
	modPath := filepath.Join(agentsDir, "rpi-verify.md")
	if err := os.WriteFile(modPath, []byte("custom content"), 0644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	// Second install — should overwrite and create backup
	installed, backedUp, err := InstallAgents(agentsDir)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if installed != 1 {
		t.Errorf("expected 1 file installed, got %d", installed)
	}
	if backedUp != 1 {
		t.Errorf("expected 1 backup, got %d", backedUp)
	}

	bakData, err := os.ReadFile(modPath + ".bak")
	if err != nil {
		t.Fatal("backup file not created")
	}
	if string(bakData) != "custom content" {
		t.Error("backup should contain original custom content")
	}

	data, _ := os.ReadFile(modPath)
	if string(data) == "custom content" {
		t.Error("file should be overwritten with embedded version")
	}
}

func TestInstallAgents_SkipsIdenticalFiles(t *testing.T) {
	agentsDir := filepath.Join(t.TempDir(), "agents")

	// First install
	_, _, err := InstallAgents(agentsDir)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install — all identical
	installed, backedUp, err := InstallAgents(agentsDir)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if installed != 0 {
		t.Errorf("expected 0 files installed, got %d", installed)
	}
	if backedUp != 0 {
		t.Errorf("expected 0 backups, got %d", backedUp)
	}
}

func TestAgentDefinitions_ValidFrontmatter(t *testing.T) {
	entries, err := fs.ReadDir(assets, "assets/agents")
	if err != nil {
		t.Fatalf("read assets/agents: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := assets.ReadFile("assets/agents/" + e.Name())
		if err != nil {
			t.Errorf("read %s: %v", e.Name(), err)
			continue
		}
		content := string(data)
		name := extractFrontmatterField(content, "name")
		if name == "" {
			t.Errorf("agent %s: missing name field", e.Name())
		}
		desc := extractFrontmatterField(content, "description")
		if desc == "" {
			t.Errorf("agent %s: missing description field", e.Name())
		}
	}
}

// --- helpers ---

func extractFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return ""
	}
	return content[4 : 4+end]
}

func extractBody(content string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(content)
}

func extractFrontmatterField(content, field string) string {
	fm := extractFrontmatter(content)
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(line, field+": ") {
			return strings.TrimPrefix(line, field+": ")
		}
	}
	return ""
}
