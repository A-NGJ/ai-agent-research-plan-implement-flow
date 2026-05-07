package split

import (
	"os"
	"strings"
	"testing"
)

func TestComputeBytes_SimpleDesignNoSplit(t *testing.T) {
	design := `---
topic: simple
---

# Design: simple

## Summary

Tiny one-component change.

## Components

A focused tweak; no numbered subheadings.

## File Structure

- ` + "`internal/foo/bar.go`" + ` — modify
`
	r, err := ComputeBytes("test.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	if r.ShouldSplit {
		t.Errorf("simple design should not trigger split, got score=%d", r.Score.Total)
	}
	if r.Threshold != DefaultThreshold {
		t.Errorf("threshold should default to %d, got %d", DefaultThreshold, r.Threshold)
	}
}

func TestComputeBytes_ComplexDesignSplits(t *testing.T) {
	design := `---
topic: complex
---

# Design: complex

## Components

### 1. Storage layout

Layout description.

### 2. Search tool

Search tool.

### 3. Capture invariant

Capture description.

### 4. Manual commands

Manual commands.

### 5. Read-side integration

Read integration.

## File Structure

- ` + "`cmd/rpi/foo.go`" + ` — new
- ` + "`internal/search/wiki.go`" + ` — new
- ` + "`.rpi/specs/feature.md`" + ` — new
- ` + "`CLAUDE.md`" + ` — append
`
	r, err := ComputeBytes("complex.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	if !r.ShouldSplit {
		t.Errorf("complex design should trigger split, got score=%d (components=%d dirs=%d specs=%d)",
			r.Score.Total, r.Components.Count, r.Dirs.Count, r.Specs.Count)
	}
	if r.Components.Count != 5 {
		t.Errorf("expected 5 components, got %d (headings=%v)", r.Components.Count, r.Components.Headings)
	}
	if r.Components.Source != "components" {
		t.Errorf("expected components source, got %q", r.Components.Source)
	}
}

func TestComputeBytes_FallsBackToFileStructureForComponents(t *testing.T) {
	design := `## File Structure

### 1. cmd
` + "- `cmd/foo.go`" + `

### 2. internal
` + "- `internal/foo.go`" + `

### 3. spec
` + "- `.rpi/specs/x.md`" + `
`
	r, err := ComputeBytes("fallback.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	if r.Components.Source != "file_structure" {
		t.Errorf("expected file_structure source, got %q", r.Components.Source)
	}
	if r.Components.Count != 3 {
		t.Errorf("expected 3 components from File Structure fallback, got %d", r.Components.Count)
	}
}

func TestComputeBytes_NoComponentsNoFileStructure(t *testing.T) {
	design := `## Summary

Just prose.
`
	r, err := ComputeBytes("bare.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	if r.Components.Count != 0 {
		t.Errorf("expected 0 components, got %d", r.Components.Count)
	}
	if r.Components.Source != "" {
		t.Errorf("expected empty source, got %q", r.Components.Source)
	}
	if r.ShouldSplit {
		t.Errorf("bare design should not trigger split")
	}
}

func TestComputeBytes_DirsCountsTopLevelOnly(t *testing.T) {
	design := `## File Structure

- ` + "`internal/workflow/assets/skills/rpi-plan/SKILL.md`" + `
- ` + "`internal/template/render.go`" + `
- ` + "`cmd/rpi/scaffold.go`" + `
- ` + "`cmd/rpi/serve.go`" + `
- ` + "`.rpi/specs/foo.md`" + `
- ` + "`CLAUDE.md`" + `
`
	r, err := ComputeBytes("dirs.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	// Top-level dirs: internal/, cmd/, .rpi/  → 3 dirs.
	// Plus root file CLAUDE.md → +1 → 4 entries total.
	if r.Dirs.Count != 4 {
		t.Errorf("expected 4 dir buckets (3 top-level + root), got %d (entries=%v)", r.Dirs.Count, r.Dirs.TopLevel)
	}
	hasRoot := false
	for _, d := range r.Dirs.TopLevel {
		if d == "<root>" {
			hasRoot = true
		}
	}
	if !hasRoot {
		t.Errorf("expected `<root>` entry for CLAUDE.md, got %v", r.Dirs.TopLevel)
	}
}

func TestComputeBytes_MultiSpecAddsTwo(t *testing.T) {
	design := `## File Structure

- ` + "`cmd/rpi/foo.go`" + `

## Out of Scope

References ` + "`.rpi/specs/feature-a.md`" + ` and ` + "`.rpi/specs/feature-b.md`" + ` and ` + "`.rpi/specs/feature-a.md`" + ` again.
`
	r, err := ComputeBytes("multispec.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	// 2 distinct specs → multi_spec contributes +2.
	if r.Specs.Count != 2 {
		t.Errorf("expected 2 distinct specs, got %d (paths=%v)", r.Specs.Count, r.Specs.Paths)
	}
	if r.Score.MultiSpecContrib != 2 {
		t.Errorf("expected multi_spec_contrib=2, got %d", r.Score.MultiSpecContrib)
	}
}

func TestComputeBytes_SingleSpecZeroContribution(t *testing.T) {
	design := `## File Structure

- ` + "`cmd/rpi/foo.go`" + `

References ` + "`.rpi/specs/only-one.md`" + ` repeatedly: ` + "`.rpi/specs/only-one.md`" + `.
`
	r, err := ComputeBytes("singlespec.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	if r.Specs.Count != 1 {
		t.Errorf("expected 1 distinct spec, got %d", r.Specs.Count)
	}
	if r.Score.MultiSpecContrib != 0 {
		t.Errorf("expected multi_spec_contrib=0, got %d", r.Score.MultiSpecContrib)
	}
}

func TestComputeBytes_ScoreFormula(t *testing.T) {
	design := `## Components

### 1. A
### 2. B
### 3. C

## File Structure

- ` + "`cmd/foo.go`" + `
- ` + "`internal/bar.go`" + `
- ` + "`.rpi/specs/x.md`" + `
- ` + "`.rpi/specs/y.md`" + `
`
	r, err := ComputeBytes("formula.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	// components = 3 → contrib = 2
	// dirs = 3 (cmd, internal, .rpi) → contrib = 2
	// specs = 2 → multi_spec contrib = 2
	// total = 6
	if r.Score.ComponentsContrib != 2 {
		t.Errorf("expected components_contrib=2, got %d", r.Score.ComponentsContrib)
	}
	if r.Score.DirsContrib != 2 {
		t.Errorf("expected dirs_contrib=2, got %d", r.Score.DirsContrib)
	}
	if r.Score.Total != 6 {
		t.Errorf("expected total=6, got %d", r.Score.Total)
	}
	if !r.ShouldSplit {
		t.Errorf("score 6 ≥ threshold 4 should split")
	}
}

func TestComputeBytes_CustomThreshold(t *testing.T) {
	design := `## Components

### 1. Only one
`
	r, err := ComputeBytes("custom.md", []byte(design), 1)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	if r.Threshold != 1 {
		t.Errorf("expected threshold=1, got %d", r.Threshold)
	}
	// components = 1 → contrib = 0; total = 0; 0 < 1 → no split
	if r.ShouldSplit {
		t.Errorf("score 0 < threshold 1 should not split")
	}
}

func TestComputeBytes_BoundaryAtThreshold(t *testing.T) {
	design := `## Components

### 1. A
### 2. B
### 3. C
### 4. D
### 5. E

## File Structure

- ` + "`cmd/foo.go`" + `
`
	r, err := ComputeBytes("boundary.md", []byte(design), 0)
	if err != nil {
		t.Fatalf("ComputeBytes: %v", err)
	}
	// components=5 → contrib=4; dirs=1 → contrib=0; specs=0 → 0; total=4
	if r.Score.Total != 4 {
		t.Errorf("expected total=4 at boundary, got %d", r.Score.Total)
	}
	if !r.ShouldSplit {
		t.Errorf("score == threshold should split")
	}
}

func TestCompute_ReadsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test-design.md"
	body := []byte(`## Components

### 1. A
### 2. B
### 3. C
### 4. D
### 5. E

## File Structure

- ` + "`cmd/foo.go`" + `
- ` + "`internal/bar.go`" + `
- ` + "`.rpi/specs/x.md`" + `
`)
	if err := os.WriteFile(path, body, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	r, err := Compute(path)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if r.Path != path {
		t.Errorf("expected path %q on result, got %q", path, r.Path)
	}
	if !r.ShouldSplit {
		t.Errorf("expected fixture to split (score=%d)", r.Score.Total)
	}
}

func TestCompute_MissingFile(t *testing.T) {
	_, err := Compute("/nonexistent/design.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read design") {
		t.Errorf("expected wrapped 'read design' error, got %v", err)
	}
}
