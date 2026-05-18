package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A-NGJ/rpi/internal/templates"
)

// renderedTemplateHeadings returns the ordered list of `## Heading` lines from
// the embedded rules-file template after the contract slot has been spliced.
// It uses the same parser the reconciler uses, so the test list reflects what
// the production code considers a "template section".
func renderedTemplateHeadings(t *testing.T, name string) []string {
	t.Helper()
	rendered, err := templates.Get(name)
	if err != nil {
		t.Fatalf("templates.Get(%q): %v", name, err)
	}
	stripped := stripContractBlock(rendered)
	var headings []string
	for _, s := range parseTemplateSections(stripped) {
		headings = append(headings, s.heading)
	}
	return headings
}

func TestReconcile_MissingFileNoOp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file should not exist after reconciler call on missing path, got err=%v", err)
	}
}

func TestReconcile_NoMissingSectionsNoWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	rendered, err := templates.Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("templates.Get: %v", err)
	}
	if err := os.WriteFile(path, []byte(rendered), 0644); err != nil {
		t.Fatal(err)
	}

	before, _ := os.ReadFile(path)
	beforeStat, _ := os.Stat(path)

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	after, _ := os.ReadFile(path)
	afterStat, _ := os.Stat(path)
	if !bytes.Equal(before, after) {
		t.Error("reconciler mutated file when no sections were missing")
	}
	if !beforeStat.ModTime().Equal(afterStat.ModTime()) {
		t.Error("reconciler rewrote file (mtime changed) when no sections were missing")
	}
}

func TestReconcile_AppendsSingleMissingSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// Start with just the H1 + Project Overview heading. Every other template
	// section is missing.
	prior := "# CLAUDE.md\n\n## Project Overview\n\nMine.\n"
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	got, _ := os.ReadFile(path)
	// Prior content preserved at start.
	if !bytes.HasPrefix(got, []byte(prior)) {
		t.Errorf("prior content not preserved at file start.\nwant prefix: %q\ngot:         %q", prior, got)
	}
	// Every other template heading must now be present, in template order.
	for _, h := range renderedTemplateHeadings(t, "CLAUDE.md") {
		if h == "## Project Overview" {
			continue // already present in prior
		}
		if !bytes.Contains(got, []byte("\n"+h+"\n")) && !bytes.HasPrefix(got, []byte(h+"\n")) {
			t.Errorf("missing template heading %q not appended.\nfile:\n%s", h, got)
		}
	}
	// Contract block must NOT be appended by the reconciler (writeContractBlock owns it).
	if bytes.Contains(got, []byte("<!-- rpi:contract:begin")) {
		t.Error("reconciler should not append the contract block — that path is owned by writeContractBlock")
	}
}

func TestReconcile_AppendsMultipleInTemplateOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// Include only the H1 — every template section is missing.
	prior := "# CLAUDE.md\n\nbare-bones file.\n"
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	got, _ := os.ReadFile(path)
	headings := renderedTemplateHeadings(t, "CLAUDE.md")
	// Each heading must appear after the previous one.
	prevIdx := -1
	for _, h := range headings {
		idx := bytes.Index(got, []byte("\n"+h+"\n"))
		if idx < 0 {
			t.Errorf("heading %q missing after reconcile", h)
			continue
		}
		if idx <= prevIdx {
			t.Errorf("heading %q at byte %d not after previous heading at %d (out of template order)", h, idx, prevIdx)
		}
		prevIdx = idx
	}
}

func TestReconcile_IdempotentOnRerun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	prior := "# CLAUDE.md\n\nbare-bones file.\n"
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("first reconcile: %v", err)
	}
	first, _ := os.ReadFile(path)
	firstStat, _ := os.Stat(path)

	buf.Reset()
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("second reconcile: %v", err)
	}
	second, _ := os.ReadFile(path)
	secondStat, _ := os.Stat(path)

	if !bytes.Equal(first, second) {
		t.Error("second reconcile call mutated file content")
	}
	if !firstStat.ModTime().Equal(secondStat.ModTime()) {
		t.Error("second reconcile call rewrote file (mtime changed)")
	}
}

func TestReconcile_DriftedBodyLeftAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// File has every template heading but the Project Overview body has been
	// rewritten by the user. The reconciler must not touch the body.
	headings := renderedTemplateHeadings(t, "CLAUDE.md")
	var body strings.Builder
	body.WriteString("# CLAUDE.md\n\n")
	for _, h := range headings {
		body.WriteString(h)
		body.WriteString("\n\n")
		if h == "## Project Overview" {
			body.WriteString("USER-OWNED OVERVIEW — keep this.\n\n")
		} else {
			body.WriteString("placeholder\n\n")
		}
	}
	prior := body.String()
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	got, _ := os.ReadFile(path)
	if !bytes.Contains(got, []byte("USER-OWNED OVERVIEW — keep this.")) {
		t.Error("user-edited body was overwritten by reconciler")
	}
	if !bytes.Equal(got, []byte(prior)) {
		t.Errorf("file should be unchanged when all headings present.\nwant: %q\ngot:  %q", prior, got)
	}
}

func TestReconcile_PreservesExistingContractBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// File contains the contract block in the middle. Missing template
	// headings should appear at EOF (after the contract block), but the
	// contract block bytes must be preserved untouched.
	contractMarker := "<!-- rpi:contract:begin v=1 -->\n## RPI Skill Contract\n\nbody.\n<!-- rpi:contract:end -->\n"
	prior := "# CLAUDE.md\n\n## Project Overview\n\nMine.\n\n" + contractMarker
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	got, _ := os.ReadFile(path)
	if !bytes.Contains(got, []byte(contractMarker)) {
		t.Errorf("contract block bytes not preserved verbatim.\nfile:\n%s", got)
	}
	// Confirm a template section that was missing now sits *after* the contract block.
	headings := renderedTemplateHeadings(t, "CLAUDE.md")
	contractEndIdx := bytes.Index(got, []byte("<!-- rpi:contract:end -->"))
	if contractEndIdx < 0 {
		t.Fatal("contract end marker missing after reconcile")
	}
	var appendedHeading string
	for _, h := range headings {
		if h == "## Project Overview" {
			continue
		}
		appendedHeading = h
		break
	}
	if appendedHeading == "" {
		t.Skip("no other template headings to verify positioning")
	}
	headingIdx := bytes.Index(got, []byte("\n"+appendedHeading+"\n"))
	if headingIdx < 0 {
		t.Fatalf("expected appended heading %q not found", appendedHeading)
	}
	if headingIdx < contractEndIdx {
		t.Errorf("appended heading %q at %d should come after contract end at %d", appendedHeading, headingIdx, contractEndIdx)
	}
}

func TestReconcile_AGENTSmd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	prior := "# AGENTS.md\n\nbare-bones AGENTS file.\n"
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "AGENTS.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	got, _ := os.ReadFile(path)
	if !bytes.HasPrefix(got, []byte(prior)) {
		t.Error("prior AGENTS.md content not preserved at file start")
	}
	for _, h := range renderedTemplateHeadings(t, "AGENTS.md") {
		if !bytes.Contains(got, []byte("\n"+h+"\n")) && !bytes.HasPrefix(got, []byte(h+"\n")) {
			t.Errorf("AGENTS.md template heading %q not appended.\nfile:\n%s", h, got)
		}
	}
	if bytes.Contains(got, []byte("<!-- rpi:contract:begin")) {
		t.Error("reconciler should not append the contract block to AGENTS.md")
	}
}

func TestReconcile_HeadingInsideCodeFenceIgnored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// User content includes a code fence with a `## ` line inside. The line
	// is not a real heading; the reconciler must still append the missing
	// template sections and leave the fenced line intact.
	prior := "# CLAUDE.md\n\n```bash\n# example\n## an arbitrary comment\n```\n"
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := reconcileRulesFileSections(buf, path, "CLAUDE.md"); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	got, _ := os.ReadFile(path)
	if !bytes.HasPrefix(got, []byte(prior)) {
		t.Errorf("prior fenced content not preserved.\nwant prefix: %q\ngot:         %q", prior, got)
	}
	// All real template headings should still have been appended.
	for _, h := range renderedTemplateHeadings(t, "CLAUDE.md") {
		if !bytes.Contains(got, []byte("\n"+h+"\n")) && !bytes.HasPrefix(got, []byte(h+"\n")) {
			t.Errorf("template heading %q not appended despite fenced ## decoy", h)
		}
	}
}
