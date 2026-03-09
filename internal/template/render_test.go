package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Auth Refactor!", "auth-refactor"},
		{"Hello World", "hello-world"},
		{"  leading and trailing  ", "leading-and-trailing"},
		{"UPPER CASE", "upper-case"},
		{"special@#$chars", "special-chars"},
		{"multiple   spaces", "multiple-spaces"},
		{"already-slug", "already-slug"},
		{"with--double---hyphens", "with-double-hyphens"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	ctx := &RenderContext{
		Date:   "2026-03-08T10:00:00+01:00",
		Topic:  "templates scaffold",
		Ticket: "cli-002",
		Prefix: "cli",
		Number: 2,
	}

	tests := []struct {
		artifactType string
		ctx          *RenderContext
		want         string
	}{
		{"research", ctx, "2026-03-08-templates-scaffold.md"},
		{"design", ctx, "2026-03-08-templates-scaffold.md"},
		{"structure", ctx, "2026-03-08-templates-scaffold.md"},
		{"plan", ctx, "2026-03-08-cli-002-templates-scaffold.md"},
		{"ticket", ctx, "cli-002-templates-scaffold.md"},
		{"ticket-index", ctx, "cli-index.md"},
		{"verify-report", ctx, "2026-03-08-verify-templates-scaffold.md"},
		{"spec", ctx, "templates-scaffold.md"},
	}

	for _, tt := range tests {
		t.Run(tt.artifactType, func(t *testing.T) {
			got := GenerateFilename(tt.artifactType, tt.ctx)
			if got != tt.want {
				t.Errorf("GenerateFilename(%q) = %q, want %q", tt.artifactType, got, tt.want)
			}
		})
	}
}

func TestGenerateFilenamePlanWithoutTicket(t *testing.T) {
	ctx := &RenderContext{
		Date:  "2026-03-08T10:00:00+01:00",
		Topic: "templates scaffold",
	}
	got := GenerateFilename("plan", ctx)
	want := "2026-03-08-templates-scaffold.md"
	if got != want {
		t.Errorf("GenerateFilename(plan, no ticket) = %q, want %q", got, want)
	}
}

func TestResolveAutoVars(t *testing.T) {
	ctx := &RenderContext{}
	err := ResolveAutoVars(ctx)
	if err != nil {
		t.Fatalf("ResolveAutoVars() error: %v", err)
	}

	if ctx.Date == "" {
		t.Error("Date should be populated")
	}
	// In a git repo, these should be real values
	if ctx.GitCommit == "" {
		t.Error("GitCommit should be populated")
	}
	if ctx.Branch == "" {
		t.Error("Branch should be populated")
	}
	if ctx.Repository == "" {
		t.Error("Repository should be populated")
	}
}

func TestRenderTemplate(t *testing.T) {
	ctx := &RenderContext{
		Date:      "2026-03-08T10:00:00+01:00",
		Topic:     "Auth Refactor",
		Ticket:    "cli-007",
		TypeLabel: "Plan",
	}

	result, err := RenderTemplate("simple", ctx, "testdata")
	if err != nil {
		t.Fatalf("RenderTemplate() error: %v", err)
	}

	if !strings.Contains(result, "topic: \"Auth Refactor\"") {
		t.Error("output should contain topic")
	}
	if !strings.Contains(result, "ticket: \"cli-007\"") {
		t.Error("output should contain ticket")
	}
	if !strings.Contains(result, "# Plan: Auth Refactor") {
		t.Error("output should contain heading with type label")
	}
}

func TestRenderTemplateConditionals(t *testing.T) {
	ctx := &RenderContext{
		Date:      "2026-03-08T10:00:00+01:00",
		Topic:     "Simple Task",
		TypeLabel: "Plan",
		// Ticket intentionally empty
	}

	result, err := RenderTemplate("simple", ctx, "testdata")
	if err != nil {
		t.Fatalf("RenderTemplate() error: %v", err)
	}

	if strings.Contains(result, "ticket:") {
		t.Error("output should NOT contain ticket when Ticket is empty")
	}
}

func TestRenderTemplateMissingFile(t *testing.T) {
	ctx := &RenderContext{}
	_, err := RenderTemplate("nonexistent", ctx, "testdata")
	if err == nil {
		t.Error("expected error for missing template file")
	}
}

func TestRenderTemplateInvalidSyntax(t *testing.T) {
	ctx := &RenderContext{}
	_, err := RenderTemplate("invalid", ctx, "testdata")
	if err == nil {
		t.Error("expected error for invalid template syntax")
	}
}

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

func TestRenderAllTemplates(t *testing.T) {
	repoRoot := findRepoRoot(t)
	templatesDir := filepath.Join(repoRoot, ".claude", "templates")

	// Full context with all fields populated
	fullCtx := &RenderContext{
		Date:       "2026-03-08T10:00:00+01:00",
		GitCommit:  "abc1234",
		Branch:     "main",
		Repository: "test-repo",
		Topic:      "Test Topic",
		Ticket:     "cli-007",
		Design:     ".thoughts/designs/2026-03-08-test.md",
		Research:   ".thoughts/research/2026-03-08-test.md",
		Structure:  ".thoughts/structures/2026-03-08-test.md",
		Tags:       "go, cli",
		Prefix:     "cli",
		Number:     7,
		TypeLabel:  "Plan",
	}

	templates := []struct {
		name         string
		wantInOutput []string
	}{
		{"research", []string{"# Research: Test Topic", "researcher: Claude", "git_commit: abc1234"}},
		{"design", []string{"# Design: Test Topic", "related_research:"}},
		{"plan", []string{"# cli-007: Test Topic", "ticket: \"cli-007\"", "design: \".thoughts/designs/"}},
		{"ticket", []string{"# cli-007: Test Topic", "ticket_id: cli-007", "depends_on: []"}},
		{"ticket-index", []string{"# Tickets: Test Topic", "prefix: cli"}},
		{"structure", []string{"# Structure: Test Topic", "## Directory Layout"}},
		{"verify-report", []string{"# Verification Report: Test Topic", "## Completeness"}},
		{"spec", []string{"domain: Test Topic", "## Purpose", "## Behavior"}},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTemplate(tt.name, fullCtx, templatesDir)
			if err != nil {
				t.Fatalf("RenderTemplate(%q) error: %v", tt.name, err)
			}

			for _, want := range tt.wantInOutput {
				if !strings.Contains(result, want) {
					t.Errorf("output missing %q\n\nGot:\n%s", want, result)
				}
			}

			// Verify output starts with frontmatter
			if !strings.HasPrefix(result, "---\n") {
				t.Errorf("output should start with frontmatter delimiter, got: %q", result[:min(50, len(result))])
			}
		})
	}
}

func TestRenderTemplatesWithoutOptionalVars(t *testing.T) {
	repoRoot := findRepoRoot(t)
	templatesDir := filepath.Join(repoRoot, ".claude", "templates")

	// Minimal context — only required fields
	minCtx := &RenderContext{
		Date:       "2026-03-08T10:00:00+01:00",
		GitCommit:  "abc1234",
		Branch:     "main",
		Repository: "test-repo",
		Topic:      "Minimal Test",
		TypeLabel:  "Plan",
		Prefix:     "test",
		Number:     1,
		Ticket:     "test-001", // needed for ticket type
	}

	// Templates that should omit optional sections when vars are empty
	tests := []struct {
		name        string
		notInOutput []string
	}{
		{"design", []string{"related_research:"}},
		{"plan", []string{"- **Research**:"}},
		{"structure", []string{"- Design:"}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_no_optionals", func(t *testing.T) {
			// Clear optional fields
			ctx := *minCtx
			ctx.Research = ""
			ctx.Design = ""
			ctx.Structure = ""
			ctx.Ticket = ""
			if tt.name == "plan" {
				// plan doesn't require ticket
			}

			result, err := RenderTemplate(tt.name, &ctx, templatesDir)
			if err != nil {
				t.Fatalf("RenderTemplate(%q) error: %v", tt.name, err)
			}

			for _, notWant := range tt.notInOutput {
				if strings.Contains(result, notWant) {
					t.Errorf("output should NOT contain %q when optional var is empty\n\nGot:\n%s", notWant, result)
				}
			}
		})
	}
}
