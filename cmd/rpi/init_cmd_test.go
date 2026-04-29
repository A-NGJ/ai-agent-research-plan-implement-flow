package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetInitFlags() {
	initNoClaudeMD = false
	initNoTrack = false
	initTarget = "claude"
	initNoMCP = false
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

// TC-1: Fresh claude init
func TestInitCreatesAllDirs(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AS-14: no .claude/commands/
	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands")); err == nil {
		t.Error(".claude/commands should not be created (AS-14)")
	}

	// Verify .claude/ subdirs (skills, hooks, agents)
	for _, d := range []string{"skills", "hooks", "agents"} {
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

	// AS-2: .claude/skills/ has 9 skill dirs with SKILL.md files
	claudeSkills := filepath.Join(dir, ".claude", "skills")
	entries, err := os.ReadDir(claudeSkills)
	if err != nil {
		t.Fatalf(".claude/skills/ not created: %v", err)
	}
	if len(entries) != 10 {
		t.Errorf("expected 10 skill dirs in .claude/skills/, got %d", len(entries))
	}

	// No .agents/ directory for claude target
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err == nil {
		t.Error(".agents/ should not be created for claude target")
	}

	// Verify .rpi/ subdirs
	rpiSubdirs := []string{
		"research", "designs", "diagnoses",
		"plans", "specs", "reviews", "archive",
	}
	for _, d := range rpiSubdirs {
		path := filepath.Join(dir, ".rpi", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".rpi/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".rpi/%s is not a directory", d)
		}
	}

	// Verify CLAUDE.md created
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not created")
	}

	// Verify .gitignore entries
	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), ".claude/") {
		t.Error(".gitignore missing .claude/ entry")
	}
	// .rpi/* should be gitignored by default, with .rpi/specs/ tracked
	if !strings.Contains(string(gitignore), ".rpi/*\n") {
		t.Error(".rpi/* should be in .gitignore by default")
	}
	if !strings.Contains(string(gitignore), "!.rpi/specs/\n") {
		t.Error("!.rpi/specs/ negation should be in .gitignore by default")
	}
	if strings.Contains(string(gitignore), ".rpi/\n") {
		t.Error("bare .rpi/ should not be in .gitignore by default (would gitignore specs too)")
	}
	output := buf.String()
	if !strings.Contains(output, "Created .claude/skills/") {
		t.Error("output missing .claude/skills/ creation message")
	}
	if !strings.Contains(output, "Created .rpi/research/") {
		t.Error("output missing .rpi/research/ creation message")
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

	// Pre-create some .rpi/ dirs but not .claude/
	os.MkdirAll(filepath.Join(dir, ".rpi", "research"), 0755)
	os.MkdirAll(filepath.Join(dir, ".rpi", "plans"), 0755)

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All dirs should exist
	for _, d := range []string{"designs", "specs", "reviews", "archive"} {
		path := filepath.Join(dir, ".rpi", d)
		if _, err := os.Stat(path); err != nil {
			t.Errorf(".rpi/%s not created: %v", d, err)
		}
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
}

func TestInitNoTrack(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initNoTrack = true
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

	if !strings.Contains(string(gitignore), ".rpi/\n") {
		t.Error(".rpi/ should be in .gitignore with --no-track")
	}
	if strings.Contains(string(gitignore), "!.rpi/specs/") {
		t.Error("--no-track should not emit the specs negation entry")
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

	if _, err := os.Stat(filepath.Join(target, ".claude", "skills")); err != nil {
		t.Error(".claude/skills not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(target, ".claude", "skills")); err != nil {
		t.Error(".claude/skills not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(target, ".rpi", "plans")); err != nil {
		t.Error(".rpi/plans not created in target dir")
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
		t.Fatal("should error when .claude/ already exists")
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
	if !strings.Contains(content, ".rpi/*\n") {
		t.Error(".rpi/* should be in .gitignore by default")
	}
	if !strings.Contains(content, "!.rpi/specs/\n") {
		t.Error("!.rpi/specs/ negation should be in .gitignore by default")
	}
	if strings.Contains(content, ".rpi/\n") {
		t.Error("bare .rpi/ should not be in .gitignore by default")
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
}

func TestInitInstallsSkills(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify skills in .claude/skills/
	expectedSkills := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
	}
	for _, skill := range expectedSkills {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf(".claude/skills/%s/SKILL.md not installed", skill)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Installed") {
		t.Error("output missing install confirmation")
	}
}

func TestInitInstallsAgents(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify agent files in .claude/agents/
	agentsDir := filepath.Join(dir, ".claude", "agents")
	if _, err := os.Stat(filepath.Join(agentsDir, "rpi-verify.md")); err != nil {
		t.Errorf(".claude/agents/rpi-verify.md not installed")
	}

	output := buf.String()
	if !strings.Contains(output, "Installed 1 agent file") {
		t.Error("output missing agent install confirmation")
	}
}

func TestInitAgentsOnlyNoAgentDefs(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No agent definitions should exist for agents-only target
	agentsDir := filepath.Join(dir, ".agents", "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		t.Error("agents-only target should not have agent definitions")
	}
}

func TestInitSucceedsWithEmptyDir(t *testing.T) {
	dir := t.TempDir()

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func runInitOpenCode(t *testing.T, dir string) (*bytes.Buffer, error) {
	t.Helper()
	resetInitFlags()
	initTarget = "opencode"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	return buf, err
}

func TestInitOpenCode(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitOpenCode(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No commands/ or agents/ for opencode (agents are Claude-only)
	for _, d := range []string{"commands", "agents"} {
		if _, err := os.Stat(filepath.Join(dir, ".opencode", d)); err == nil {
			t.Errorf(".opencode/%s should not be created", d)
		}
	}

	// Verify .opencode/ subdirs (skills, hooks only)
	for _, d := range []string{"skills", "hooks"} {
		path := filepath.Join(dir, ".opencode", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".opencode/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".opencode/%s is not a directory", d)
		}
	}

	// AS-3: .opencode/skills/ has 9 dirs
	ocSkills := filepath.Join(dir, ".opencode", "skills")
	entries, err := os.ReadDir(ocSkills)
	if err != nil {
		t.Fatalf(".opencode/skills/ not created: %v", err)
	}
	if len(entries) != 10 {
		t.Errorf("expected 10 skill dirs in .opencode/skills/, got %d", len(entries))
	}

	// Verify AGENTS.md generated, CLAUDE.md absent
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Error("AGENTS.md not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should not be created for opencode target")
	}

	// Verify .opencode/ in .gitignore
	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), ".opencode/") {
		t.Error(".gitignore missing .opencode/ entry")
	}
	// .claude/ should NOT be in .gitignore
	if strings.Contains(string(gitignore), ".claude/") {
		t.Error(".gitignore should not contain .claude/ for opencode target")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .opencode/skills/") {
		t.Error("output missing .opencode/skills/ creation message")
	}
	if !strings.Contains(output, "Installed") {
		t.Error("output missing install confirmation")
	}
}

func TestInitOpenCodeAlreadyExists(t *testing.T) {
	dir := t.TempDir()

	// First run
	_, err := runInitOpenCode(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Second run should error
	_, err = runInitOpenCode(t, dir)
	if err == nil {
		t.Fatal("second run should return error")
	}
	if !strings.Contains(err.Error(), ".opencode/ already exists") {
		t.Errorf("expected '.opencode/ already exists' error, got: %v", err)
	}
}

func TestInitInvalidTarget(t *testing.T) {
	dir := t.TempDir()
	resetInitFlags()
	initTarget = "invalid"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("should error for invalid target")
	}
	if !strings.Contains(err.Error(), "unknown target") {
		t.Errorf("expected 'unknown target' error, got: %v", err)
	}
}

// TC-2: Fresh agents-only init
func TestInitAgentsOnly(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AS-4: .agents/skills/ has 9 dirs
	agentsSkills := filepath.Join(dir, ".agents", "skills")
	entries, err := os.ReadDir(agentsSkills)
	if err != nil {
		t.Fatalf(".agents/skills/ not created: %v", err)
	}
	if len(entries) != 10 {
		t.Errorf("expected 10 skill dirs, got %d", len(entries))
	}

	// No .claude/ or .opencode/ directories
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		t.Error(".claude/ should not exist for agents-only")
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); err == nil {
		t.Error(".opencode/ should not exist for agents-only")
	}

	// No CLAUDE.md or AGENTS.md
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should not exist for agents-only")
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Error("AGENTS.md should not exist for agents-only")
	}

	// .rpi/ should still exist
	if _, err := os.Stat(filepath.Join(dir, ".rpi", "plans")); err != nil {
		t.Error(".rpi/plans/ not created for agents-only")
	}
}

func TestInitAgentsOnlyIdempotent(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("first run: %v", err)
	}

	// Second run should error
	buf = new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("second run should error")
	}
	if !strings.Contains(err.Error(), ".agents/ already exists") {
		t.Errorf("expected '.agents/ already exists' error, got: %v", err)
	}
}

func TestInitAgentsOnlyNotInGitignore(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		// No .gitignore is fine for agents-only (no tool dir to gitignore)
		return
	}

	// .agents/ should NOT be in .gitignore (skills should be shared)
	if strings.Contains(string(data), ".agents/") {
		t.Error(".agents/ should not be in .gitignore")
	}
}

// spec:MC-12
func TestInitCreatesSettingsJSON(t *testing.T) {
	dir := t.TempDir()
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings.json not created: %v", err)
	}
	if !strings.Contains(string(data), "mcp__rpi__*") {
		t.Error("settings.json missing mcp__rpi__* permission")
	}
}

func TestInitSettingsJSONMergesExisting(t *testing.T) {
	dir := t.TempDir()
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Add a custom key to settings.json
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	os.WriteFile(settingsPath, []byte(`{"permissions":{"allow":["mcp__rpi__*"]},"customKey":"value"}`), 0644)

	// Call configureSettings again (simulating update)
	buf := new(bytes.Buffer)
	configureSettings(buf, filepath.Join(dir, ".claude"))

	data, _ := os.ReadFile(settingsPath)
	content := string(data)
	if !strings.Contains(content, "mcp__rpi__*") {
		t.Error("mcp__rpi__* permission lost after merge")
	}
	if !strings.Contains(content, "customKey") {
		t.Error("existing customKey lost after merge")
	}
}

func TestInitSettingsJSONIdempotent(t *testing.T) {
	dir := t.TempDir()
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Call configureSettings again
	buf := new(bytes.Buffer)
	configureSettings(buf, filepath.Join(dir, ".claude"))

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	count := strings.Count(string(data), "mcp__rpi__*")
	if count != 1 {
		t.Errorf("expected 1 mcp__rpi__* entry, got %d", count)
	}
}

func TestInitOpenCodeNoSettingsJSON(t *testing.T) {
	dir := t.TempDir()
	_, err := runInitOpenCode(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); err == nil {
		t.Error("settings.json should not be created for opencode target")
	}
}

func TestInitNoMCPFlag(t *testing.T) {
	flag := initCmd.Flags().Lookup("no-mcp")
	if flag == nil {
		t.Fatal("--no-mcp flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("--no-mcp default = %q, want %q", flag.DefValue, "false")
	}
}

// spec:MC-1 spec:MC-6
func TestInitWritesMCPConfig(t *testing.T) {
	dir := t.TempDir()

	var calls [][]string
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		// First call: "claude mcp get rpi" → not found
		if len(calls) == 1 {
			return []byte("No MCP server found"), fmt.Errorf("exit 1")
		}
		// Second call: "claude mcp add rpi -- rpi serve" → success
		return []byte("Added"), nil
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 claude mcp calls, got %d: %v", len(calls), calls)
	}
	// Verify the add command shape
	addCall := calls[1]
	expected := []string{"claude", "mcp", "add", "rpi", "--", "rpi", "serve"}
	if len(addCall) != len(expected) {
		t.Fatalf("add call = %v, want %v", addCall, expected)
	}
	for i, v := range expected {
		if addCall[i] != v {
			t.Errorf("add call[%d] = %q, want %q", i, addCall[i], v)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Configured MCP server via claude mcp add") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// spec:MC-3
func TestInitCallsClaudeMCPAdd(t *testing.T) {
	dir := t.TempDir()

	var addCalled bool
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
			return nil, fmt.Errorf("not found")
		}
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "add" {
			addCalled = true
			return []byte("Added"), nil
		}
		return nil, fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !addCalled {
		t.Error("claude mcp add was not called")
	}
}

// spec:MC-4
func TestInitWarnsExistingMCPEntry(t *testing.T) {
	dir := t.TempDir()

	var addCalled bool
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
			// Server already exists
			return []byte("rpi: connected"), nil
		}
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "add" {
			addCalled = true
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already configured") {
		t.Error("expected warning about existing MCP entry")
	}
	if addCalled {
		t.Error("claude mcp add should not be called when server already exists")
	}
}

// spec:MC-5
func TestInitSkipsMCPWithFlag(t *testing.T) {
	dir := t.TempDir()

	var called bool
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	resetInitFlags()
	initNoMCP = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Error("claude mcp should not be called with --no-mcp")
	}
}

// spec:MC-6
func TestInitMCPAddCommandShape(t *testing.T) {
	dir := t.TempDir()

	var addArgs []string
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
			return nil, fmt.Errorf("not found")
		}
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "add" {
			addArgs = append([]string{name}, args...)
			return []byte("Added"), nil
		}
		return nil, fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify exact command: claude mcp add rpi -- rpi serve
	expected := []string{"claude", "mcp", "add", "rpi", "--", "rpi", "serve"}
	if len(addArgs) != len(expected) {
		t.Fatalf("add args = %v, want %v", addArgs, expected)
	}
	for i, v := range expected {
		if addArgs[i] != v {
			t.Errorf("add args[%d] = %q, want %q", i, addArgs[i], v)
		}
	}
}

func TestConfigureHooksAddsAllHooks(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Start with a settings.json that has permissions only
	os.WriteFile(filepath.Join(claudeDir, "settings.json"),
		[]byte(`{"permissions":{"allow":["mcp__rpi__*"]}}`), 0644)

	buf := new(bytes.Buffer)
	configureHooks(buf, claudeDir)

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	content := string(data)

	// All three hooks should be present
	for _, check := range []struct {
		event  string
		marker string
	}{
		{"PostCompact", "rpi_context_essentials"},
		{"SessionStart", "rpi_session_resume"},
		{"Stop", "rpi_suggest_next"},
	} {
		if !strings.Contains(content, check.event) {
			t.Errorf("settings.json missing %s hook", check.event)
		}
		if !strings.Contains(content, check.marker) {
			t.Errorf("%s hook missing %s reference", check.event, check.marker)
		}
	}

	// Verify permissions weren't clobbered
	if !strings.Contains(content, "mcp__rpi__*") {
		t.Error("permissions lost after configureHooks")
	}

	// Verify hook entries use matcher+hooks structure
	if !strings.Contains(content, `"matcher"`) {
		t.Error("hook entries should contain matcher field")
	}
	// Each hook entry wraps its command in a hooks array
	var parsed map[string]json.RawMessage
	json.Unmarshal(data, &parsed)
	var hooks map[string]json.RawMessage
	json.Unmarshal(parsed["hooks"], &hooks)
	for event, raw := range hooks {
		var entries []struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		}
		if err := json.Unmarshal(raw, &entries); err != nil {
			t.Errorf("%s: failed to parse as matcher+hooks structure: %v", event, err)
			continue
		}
		for _, entry := range entries {
			if len(entry.Hooks) == 0 {
				t.Errorf("%s: matcher entry has empty hooks array", event)
			}
		}
	}
}

func TestConfigureHooksIdempotent(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{}`), 0644)

	buf := new(bytes.Buffer)
	configureHooks(buf, claudeDir)
	configureHooks(buf, claudeDir)

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	content := string(data)

	// Each marker should appear exactly once
	for _, marker := range []string{"rpi_context_essentials", "rpi_session_resume", "rpi_suggest_next"} {
		count := strings.Count(content, marker)
		if count != 1 {
			t.Errorf("expected 1 %s reference, got %d", marker, count)
		}
	}
}

func TestConfigureHooksMergesExisting(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Start with existing hooks
	os.WriteFile(filepath.Join(claudeDir, "settings.json"),
		[]byte(`{"hooks":{"PreToolUse":[{"type":"command","command":"echo pre"}]}}`), 0644)

	buf := new(bytes.Buffer)
	configureHooks(buf, claudeDir)

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	content := string(data)
	if !strings.Contains(content, "PreToolUse") {
		t.Error("existing PreToolUse hook was clobbered")
	}
	if !strings.Contains(content, "PostCompact") {
		t.Error("PostCompact hook not added")
	}
	if !strings.Contains(content, "SessionStart") {
		t.Error("SessionStart hook not added")
	}
	if !strings.Contains(content, "Stop") {
		t.Error("Stop hook not added")
	}
}

func TestConfigureHooksReplacesOldFormat(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Start with old-format hooks (flat {type, command} without matcher)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"),
		[]byte(`{"hooks":{"PostCompact":[{"type":"command","command":"cat <<'HOOK_EOF'\nrpi_context_essentials\nHOOK_EOF"}]}}`), 0644)

	buf := new(bytes.Buffer)
	configureHooks(buf, claudeDir)

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	content := string(data)

	// Old entry should be replaced with correct structure
	if !strings.Contains(content, `"matcher"`) {
		t.Error("should replace old-format entry with matcher+hooks structure")
	}
	// Marker should appear exactly once (not duplicated)
	if count := strings.Count(content, "rpi_context_essentials"); count != 1 {
		t.Errorf("expected 1 rpi_context_essentials reference, got %d", count)
	}
}
