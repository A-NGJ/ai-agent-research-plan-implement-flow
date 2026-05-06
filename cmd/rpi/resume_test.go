package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runResumeCmd(t *testing.T, dir, format string) (string, error) {
	t.Helper()
	oldFlag := rpiDirFlag
	oldFormat := formatFlag
	rpiDirFlag = dir
	formatFlag = format
	t.Cleanup(func() {
		rpiDirFlag = oldFlag
		formatFlag = oldFormat
	})

	buf := new(bytes.Buffer)
	resumeCmd.SetOut(buf)
	err := runResume(resumeCmd, nil)
	return buf.String(), err
}

func TestResumeText_RendersSections(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-plan.md", `---
topic: "active plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Setup

- [x] Done
- [ ] Pending one
- [ ] Pending two
`)
	writeTempArtifact(t, dir, "designs", "2026-01-01-design.md", `---
topic: "active design"
status: active
date: 2026-01-01T00:00:00Z
---

# Design
`)

	out, err := runResumeCmd(t, dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"Artifacts", "Active Plan", "Suggestion", "active plan", "1/3 (33%)", "Next:"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n--- output ---\n%s", want, out)
		}
	}
}

func TestResumeText_EmptyState(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "plans"), 0755)

	out, err := runResumeCmd(t, dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"No active work", "/rpi-propose"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n--- output ---\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Artifacts\n", "Active Plan\n", "Suggestion\n"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("empty-state output should not contain section header %q\n--- output ---\n%s", unwanted, out)
		}
	}
}

func TestResumeJSON_PreservesShape(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-plan.md", `---
topic: "active plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Work

- [ ] Pending
`)
	writeTempArtifact(t, dir, "designs", "2026-01-01-design.md", `---
topic: "active design"
status: active
date: 2026-01-01T00:00:00Z
---

# Design
`)

	out, err := runResumeCmd(t, dir, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got ResumeResult
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n--- output ---\n%s", err, out)
	}

	want, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("assembleResume failed: %v", err)
	}

	if !reflect.DeepEqual(&got, want) {
		t.Errorf("JSON shape mismatch\ngot:  %+v\nwant: %+v", got, *want)
	}
}

func TestResumeFormat_Unknown(t *testing.T) {
	dir := t.TempDir()
	_, err := runResumeCmd(t, dir, "yaml")
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("error = %q, want to contain 'unknown format'", err.Error())
	}
}

func TestResumeActiveArtifacts(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-plan.md", `---
topic: "active plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Work

- [ ] Pending
`)
	writeTempArtifact(t, dir, "designs", "2026-01-01-design.md", `---
topic: "active design"
status: active
date: 2026-01-01T00:00:00Z
---

# Design
`)

	result, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Artifacts) != 2 {
		t.Fatalf("artifacts count = %d, want 2", len(result.Artifacts))
	}

	// Verify both artifacts have required fields
	for _, a := range result.Artifacts {
		if a.Path == "" || a.Type == "" || a.Status == "" {
			t.Errorf("artifact missing fields: %+v", a)
		}
	}
}

func TestResumePlanProgress(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-plan.md", `---
topic: "progress plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Setup

- [x] First done
- [x] Second done

## Phase 2: Core

- [ ] First pending
- [ ] Second pending
- [ ] Third pending
`)

	result, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ActivePlan == nil {
		t.Fatal("expected non-nil ActivePlan")
	}
	if result.ActivePlan.CurrentPhase != "Phase 2: Core" {
		t.Errorf("current_phase = %q, want %q", result.ActivePlan.CurrentPhase, "Phase 2: Core")
	}
	if result.ActivePlan.Progress.Checked != 2 || result.ActivePlan.Progress.Total != 5 {
		t.Errorf("progress = %d/%d, want 2/5", result.ActivePlan.Progress.Checked, result.ActivePlan.Progress.Total)
	}
	if len(result.ActivePlan.NextItems) != 3 {
		t.Errorf("next_items count = %d, want 3", len(result.ActivePlan.NextItems))
	}
}

func TestResumeSuggestionIncluded(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-plan.md", `---
topic: "suggestable plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Work

- [ ] Pending
`)

	result, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Suggestion == nil {
		t.Fatal("expected non-nil Suggestion")
	}
	if result.Suggestion.Action == "" {
		t.Error("expected non-empty suggestion action")
	}
	if result.Suggestion.Reasoning == "" {
		t.Error("expected non-empty suggestion reasoning")
	}
}

func TestResumeEmpty(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "plans"), 0755)

	result, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Artifacts) != 0 {
		t.Errorf("artifacts count = %d, want 0", len(result.Artifacts))
	}
	if result.ActivePlan != nil {
		t.Error("expected nil ActivePlan")
	}
	if result.Suggestion == nil {
		t.Fatal("expected non-nil Suggestion even when empty")
	}
	if !strings.Contains(result.Suggestion.Action, "/rpi-propose") {
		t.Errorf("suggestion action = %q, want to contain /rpi-propose", result.Suggestion.Action)
	}
}

func TestResumeExcludesArchived(t *testing.T) {
	dir := t.TempDir()

	// Active artifact should appear
	writeTempArtifact(t, dir, "plans", "2026-01-01-active.md", `---
topic: "active plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1

- [ ] Work
`)
	// Complete artifact should NOT appear
	writeTempArtifact(t, dir, "designs", "2026-01-01-complete.md", `---
topic: "complete design"
status: complete
date: 2026-01-01T00:00:00Z
---

# Done
`)

	result, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Artifacts) != 1 {
		t.Fatalf("artifacts count = %d, want 1 (only active)", len(result.Artifacts))
	}
	if result.Artifacts[0].Status != "active" {
		t.Errorf("status = %q, want 'active'", result.Artifacts[0].Status)
	}
}

func TestResumeMostRecentPlan(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-old.md", `---
topic: "old plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1

- [ ] Old work
`)
	recentPath := writeTempArtifact(t, dir, "plans", "2026-06-01-recent.md", `---
topic: "recent plan"
status: active
date: 2026-06-01T00:00:00Z
---

## Phase 1

- [ ] Recent work
`)

	result, err := assembleResume(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ActivePlan == nil {
		t.Fatal("expected non-nil ActivePlan")
	}
	if result.ActivePlan.Path != recentPath {
		t.Errorf("active_plan path = %q, want %q", result.ActivePlan.Path, recentPath)
	}
	if result.ActivePlan.Topic != "recent plan" {
		t.Errorf("active_plan topic = %q, want %q", result.ActivePlan.Topic, "recent plan")
	}
}
