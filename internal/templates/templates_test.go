package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGet_AGENTS(t *testing.T) {
	content, err := Get("AGENTS.md")
	if err != nil {
		t.Fatalf("Get(AGENTS.md) returned error: %v", err)
	}
	if content == "" {
		t.Fatal("Get(AGENTS.md) returned empty content")
	}
	if !strings.Contains(content, "# AGENTS.md") {
		t.Error("AGENTS.md template missing expected '# AGENTS.md' header")
	}
	if !strings.Contains(content, ".thoughts/") {
		t.Error("AGENTS.md template missing .thoughts/ directory reference")
	}
}

func TestGet_CLAUDE(t *testing.T) {
	content, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md) returned error: %v", err)
	}
	if content == "" {
		t.Fatal("Get(CLAUDE.md) returned empty content")
	}
	if !strings.Contains(content, "# CLAUDE.md") {
		t.Error("CLAUDE.md template missing expected '# CLAUDE.md' header")
	}
}

func TestGet_PIPELINE(t *testing.T) {
	content, err := Get("PIPELINE.md")
	if err != nil {
		t.Fatalf("Get(PIPELINE.md) returned error: %v", err)
	}
	if content == "" {
		t.Fatal("Get(PIPELINE.md) returned empty content")
	}
	if !strings.Contains(content, "# Development Pipeline") {
		t.Error("PIPELINE.md template missing expected '# Development Pipeline' header")
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("Get(nonexistent) should return error")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("error should mention 'unknown template', got: %v", err)
	}
}

func TestGet_MatchesWorkflowAsset(t *testing.T) {
	for name := range map[string]string{
		"AGENTS.md":   "templates/AGENTS.md.template",
		"CLAUDE.md":   "templates/CLAUDE.md.template",
		"PIPELINE.md": "templates/PIPELINE.md.template",
	} {
		t.Run(name, func(t *testing.T) {
			setupTestHome(t) // ensure no user overrides

			got, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%s): %v", name, err)
			}
			if got == "" {
				t.Fatalf("Get(%s) returned empty content", name)
			}
		})
	}
}

func setupTestHome(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	return tmpDir
}

func TestEnsureUserTemplates(t *testing.T) {
	home := setupTestHome(t)

	err := EnsureUserTemplates()
	if err != nil {
		t.Fatalf("EnsureUserTemplates() returned error: %v", err)
	}

	templatesDir := filepath.Join(home, ".rpi", "templates")
	for _, name := range Names() {
		path := filepath.Join(templatesDir, name+".template")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file %s to exist: %v", path, err)
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", path)
		}
	}
}

func TestEnsureUserTemplates_NoOverwrite(t *testing.T) {
	home := setupTestHome(t)

	templatesDir := filepath.Join(home, ".rpi", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	customContent := "# My Custom CLAUDE.md"
	customPath := filepath.Join(templatesDir, "CLAUDE.md.template")
	if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := EnsureUserTemplates()
	if err != nil {
		t.Fatalf("EnsureUserTemplates() returned error: %v", err)
	}

	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != customContent {
		t.Errorf("EnsureUserTemplates overwrote custom content: got %q, want %q", string(data), customContent)
	}
}

func TestGet_PrefersUserTemplate(t *testing.T) {
	home := setupTestHome(t)

	templatesDir := filepath.Join(home, ".rpi", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	customContent := "# Customized CLAUDE.md"
	if err := os.WriteFile(filepath.Join(templatesDir, "CLAUDE.md.template"), []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md) returned error: %v", err)
	}
	if got != customContent {
		t.Errorf("Get(CLAUDE.md) = %q, want user-customized %q", got, customContent)
	}
}

func TestGet_FallsBackToEmbedded(t *testing.T) {
	setupTestHome(t) // empty home, no ~/.rpi/templates/

	content, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md) returned error: %v", err)
	}
	if !strings.Contains(content, "# CLAUDE.md") {
		t.Error("expected embedded content with '# CLAUDE.md' header")
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d: %v", len(names), names)
	}

	expected := []string{"AGENTS.md", "CLAUDE.md", "PIPELINE.md"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}
