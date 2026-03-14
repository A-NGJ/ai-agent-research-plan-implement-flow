package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformCommandFrontmatter_Opus(t *testing.T) {
	input := []byte("---\ndescription: test\nmodel: opus\n---\n\n# Body\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "model: anthropic/claude-opus-4-6") {
		t.Errorf("expected model: anthropic/claude-opus-4-6, got:\n%s", got)
	}
	if strings.Contains(got, "model: opus") {
		t.Error("original model: opus should be replaced")
	}
}

func TestTransformCommandFrontmatter_Sonnet(t *testing.T) {
	input := []byte("---\ndescription: test\nmodel: sonnet\n---\n\n# Body\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "model: anthropic/claude-sonnet-4-6") {
		t.Errorf("expected model: anthropic/claude-sonnet-4-6, got:\n%s", got)
	}
}

func TestTransformCommandFrontmatter_UnknownModel(t *testing.T) {
	input := []byte("---\ndescription: test\nmodel: custom-model\n---\n\n# Body\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "model: custom-model") {
		t.Errorf("unknown model should pass through unchanged, got:\n%s", got)
	}
}

func TestTransformCommandFrontmatter_PreservesBody(t *testing.T) {
	input := []byte("---\nmodel: opus\n---\n\n# Body\nSome content here\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "# Body") {
		t.Error("body content should be preserved")
	}
	if !strings.Contains(got, "Some content here") {
		t.Error("body content should be preserved")
	}
}

func TestTransformAgentFrontmatter_Basic(t *testing.T) {
	input := "---\nname: codebase-analyzer\ndescription: Analyzes stuff\ntools: Read, Grep, Glob, LS\nmodel: inherit\n---\n\nBody content here.\n"
	got, err := transformAgentFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(got)

	if !strings.Contains(s, "mode: subagent") {
		t.Error("expected mode: subagent")
	}
	if !strings.Contains(s, "bash: false") {
		t.Error("expected bash: false in tools deny-map")
	}
	if !strings.Contains(s, "write: false") {
		t.Error("expected write: false in tools deny-map")
	}
	if !strings.Contains(s, "edit: false") {
		t.Error("expected edit: false in tools deny-map")
	}
	if strings.Contains(s, "model:") {
		t.Error("model: field should be removed")
	}
	if strings.Contains(s, "Read, Grep") {
		t.Error("original tools string should be replaced")
	}
	if !strings.Contains(s, "Body content here.") {
		t.Error("body content should be preserved")
	}
	if !strings.Contains(s, "name: codebase-analyzer") {
		t.Error("name should be preserved")
	}
}

func TestTransformAgentFrontmatter_NoFrontmatter(t *testing.T) {
	input := []byte("# No frontmatter\nJust body.\n")
	got, err := transformAgentFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(input) {
		t.Error("content without frontmatter should be returned as-is")
	}
}

func TestInstallTo_OpenCode(t *testing.T) {
	dir := t.TempDir()

	count, err := InstallTo(dir, TargetOpenCode, false)
	if err != nil {
		t.Fatalf("InstallTo error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected files to be installed")
	}

	// Verify command frontmatter transform
	cmdData, err := os.ReadFile(filepath.Join(dir, "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: anthropic/claude-opus-4-6") {
		t.Error("command model should be transformed to full ID")
	}
	if strings.Contains(string(cmdData), "model: opus") {
		t.Error("original model: opus should be replaced")
	}

	// Verify command body is tool-agnostic (no Sub-task syntax)
	if strings.Contains(string(cmdData), "Sub-task") {
		t.Error("command body should not contain tool-specific Sub-task syntax")
	}

	// Verify agent frontmatter transform
	agentData, err := os.ReadFile(filepath.Join(dir, "agents", "codebase-analyzer.md"))
	if err != nil {
		t.Fatalf("read codebase-analyzer.md: %v", err)
	}
	s := string(agentData)
	if !strings.Contains(s, "mode: subagent") {
		t.Error("agent should have mode: subagent")
	}
	if !strings.Contains(s, "bash: false") {
		t.Error("agent should have bash: false")
	}
	if strings.Contains(s, "tools: Read, Grep") {
		t.Error("original tools string should be replaced")
	}

	// Verify skills are copied as-is (no transform needed)
	skillData, err := os.ReadFile(filepath.Join(dir, "skills", "locate-codebase", "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	if !strings.Contains(string(skillData), "name: locate-codebase") {
		t.Error("skill should be copied unchanged")
	}
}

func TestInstallTo_Claude_Unchanged(t *testing.T) {
	dir := t.TempDir()

	_, err := InstallTo(dir, TargetClaude, false)
	if err != nil {
		t.Fatalf("InstallTo error: %v", err)
	}

	// Verify command keeps short model alias
	cmdData, err := os.ReadFile(filepath.Join(dir, "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: opus") {
		t.Error("Claude target should preserve original model: opus")
	}

	// Verify agent keeps original format
	agentData, err := os.ReadFile(filepath.Join(dir, "agents", "codebase-analyzer.md"))
	if err != nil {
		t.Fatalf("read codebase-analyzer.md: %v", err)
	}
	if !strings.Contains(string(agentData), "tools: Read, Grep, Glob, LS") {
		t.Error("Claude target should preserve original tools format")
	}
}

func TestInstall_BackwardCompatible(t *testing.T) {
	dir := t.TempDir()

	count, err := Install(dir, false)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected files to be installed")
	}

	// Should behave identically to InstallTo with TargetClaude
	cmdData, err := os.ReadFile(filepath.Join(dir, "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: opus") {
		t.Error("Install() should preserve Claude Code format")
	}
}
