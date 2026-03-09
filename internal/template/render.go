package template

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// RenderContext holds all variables available to templates.
type RenderContext struct {
	// Auto-populated (resolved by binary)
	Date       string // ISO 8601 with timezone
	GitCommit  string // git rev-parse --short HEAD
	Branch     string // git branch --show-current
	Repository string // basename of repo root
	Filename   string // generated from naming conventions
	Type       string // artifact type (research, design, plan, etc.)
	TypeLabel  string // display label ("Research", "Design", etc.)

	// User-provided (passed as flags)
	Topic     string
	Ticket    string
	Design    string
	Research  string
	Structure string
	Tags      string
	Prefix    string
	Number    int
}

// typeLabels maps artifact type to display label.
var typeLabels = map[string]string{
	"research":      "Research",
	"design":        "Design",
	"plan":          "Plan",
	"ticket":        "Ticket",
	"ticket-index":  "Ticket Index",
	"structure":     "Structure",
	"verify-report": "Verification Report",
	"spec":          "Spec",
}

// slugRe matches non-alphanumeric, non-hyphen characters.
var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

// multiHyphenRe matches consecutive hyphens.
var multiHyphenRe = regexp.MustCompile(`-{2,}`)

// Slugify converts a string to a URL-friendly slug.
// "Auth Refactor!" → "auth-refactor"
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = multiHyphenRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// dateFromISO extracts the YYYY-MM-DD portion from an ISO 8601 date string.
func dateFromISO(isoDate string) string {
	if len(isoDate) >= 10 {
		return isoDate[:10]
	}
	return isoDate
}

// GenerateFilename produces a filename following the naming conventions for each artifact type.
func GenerateFilename(artifactType string, ctx *RenderContext) string {
	slug := Slugify(ctx.Topic)
	datePart := dateFromISO(ctx.Date)

	switch artifactType {
	case "research", "design", "structure":
		return fmt.Sprintf("%s-%s.md", datePart, slug)
	case "plan":
		if ctx.Ticket != "" {
			return fmt.Sprintf("%s-%s-%s.md", datePart, ctx.Ticket, slug)
		}
		return fmt.Sprintf("%s-%s.md", datePart, slug)
	case "ticket":
		return fmt.Sprintf("%s-%03d-%s.md", ctx.Prefix, ctx.Number, slug)
	case "ticket-index":
		return fmt.Sprintf("%s-index.md", ctx.Prefix)
	case "verify-report":
		return fmt.Sprintf("%s-verify-%s.md", datePart, slug)
	case "spec":
		return fmt.Sprintf("%s.md", slug)
	default:
		return fmt.Sprintf("%s-%s.md", datePart, slug)
	}
}

// ResolveAutoVars populates Date, GitCommit, Branch, and Repository on the context.
func ResolveAutoVars(ctx *RenderContext) error {
	ctx.Date = time.Now().Format(time.RFC3339)

	ctx.GitCommit = gitCommand("rev-parse", "--short", "HEAD")
	ctx.Branch = gitCommand("branch", "--show-current")

	toplevel := gitCommand("rev-parse", "--show-toplevel")
	if toplevel != "unknown" {
		ctx.Repository = filepath.Base(toplevel)
	} else {
		ctx.Repository = "unknown"
	}

	return nil
}

// gitCommand runs a git command and returns trimmed stdout, or "unknown" on error.
func gitCommand(args ...string) string {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// RenderTemplate loads a .tmpl file from templatesDir and renders it with the given context.
func RenderTemplate(templateName string, ctx *RenderContext, templatesDir string) (string, error) {
	tmplPath := filepath.Join(templatesDir, templateName+".tmpl")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", tmplPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}
