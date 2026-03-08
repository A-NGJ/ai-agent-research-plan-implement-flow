package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
)

func setupArchiveTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeArchiveFile(t, dir, "plans/p1.md",
		"---\ntopic: \"Plan One\"\nstatus: complete\ndesign: designs/d1.md\n---\n# P1\n")
	writeArchiveFile(t, dir, "designs/d1.md",
		"---\ntopic: \"Design One\"\nstatus: draft\n---\n# D1\n")
	writeArchiveFile(t, dir, "research/r1.md",
		"---\ntopic: \"Research One\"\nstatus: superseded\n---\n# R1\nReferences plans/p1.md in body.\n")
	writeArchiveFile(t, dir, "tickets/t1.md",
		"---\ntopic: \"Ticket One\"\nstatus: active\n---\n# T1\n")

	return dir
}

func writeArchiveFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveScanReturnsRefCount(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := archiveScanCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []archiveScanResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	// p1 (complete) and r1 (superseded) are archivable
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	for _, r := range results {
		if r.Type == "plan" {
			// p1 is referenced by r1 in body
			if r.RefCount < 1 {
				t.Errorf("plan ref_count should be >= 1, got %d", r.RefCount)
			}
		}
	}
}

func TestArchiveCheckRefsReturnsDetails(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := archiveCheckRefsCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"designs/d1.md"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var refs []scanner.ReferenceDetail
	if err := json.Unmarshal(buf.Bytes(), &refs); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	// p1 references designs/d1.md via frontmatter design field
	if len(refs) < 1 {
		t.Errorf("expected at least 1 reference, got %d", len(refs))
	}
}

func TestArchiveCheckRefsEmpty(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := archiveCheckRefsCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"nonexistent/file.md"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var refs []scanner.ReferenceDetail
	if err := json.Unmarshal(buf.Bytes(), &refs); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

func TestArchiveMoveUpdatesFrontmatter(t *testing.T) {
	dir := setupArchiveTestDir(t)
	now := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	targetPath := filepath.Join(dir, "plans/p1.md")

	result, err := doArchiveMove(targetPath, dir, true, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read the moved file and check frontmatter
	doc, err := frontmatter.Parse(result.To)
	if err != nil {
		t.Fatalf("parse moved file: %v", err)
	}
	if doc.Frontmatter["status"] != "archived" {
		t.Errorf("expected status 'archived', got %v", doc.Frontmatter["status"])
	}
	if doc.Frontmatter["archived_date"] != "2026-03-08" {
		t.Errorf("expected archived_date '2026-03-08', got %v", doc.Frontmatter["archived_date"])
	}
	if !result.FrontmatterUpdated {
		t.Error("expected frontmatter_updated to be true")
	}
}

func TestArchiveMoveCorrectPath(t *testing.T) {
	dir := setupArchiveTestDir(t)
	now := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	targetPath := filepath.Join(dir, "plans/p1.md")

	result, err := doArchiveMove(targetPath, dir, true, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDest := filepath.Join(dir, "archive", "2026-03", "plan", "p1.md")
	if result.To != expectedDest {
		t.Errorf("expected dest %s, got %s", expectedDest, result.To)
	}
	if result.From != targetPath {
		t.Errorf("expected from %s, got %s", targetPath, result.From)
	}
}

func TestArchiveMoveFileAtDestination(t *testing.T) {
	dir := setupArchiveTestDir(t)
	now := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	targetPath := filepath.Join(dir, "plans/p1.md")

	result, err := doArchiveMove(targetPath, dir, true, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Original should not exist
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Error("original file should not exist after move")
	}
	// Destination should exist
	if _, err := os.Stat(result.To); err != nil {
		t.Errorf("destination file should exist: %v", err)
	}
}

func TestArchiveMoveRefsWithoutForce(t *testing.T) {
	dir := setupArchiveTestDir(t)
	now := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	// p1.md is referenced by r1.md in body
	targetPath := filepath.Join(dir, "plans/p1.md")

	_, err := doArchiveMove(targetPath, dir, false, now)
	if err != errHasReferences {
		t.Errorf("expected errHasReferences, got %v", err)
	}

	// File should still be in original location
	if _, statErr := os.Stat(targetPath); statErr != nil {
		t.Error("file should still exist when move is blocked")
	}
}

func TestArchiveMoveRefsWithForce(t *testing.T) {
	dir := setupArchiveTestDir(t)
	now := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	targetPath := filepath.Join(dir, "plans/p1.md")

	result, err := doArchiveMove(targetPath, dir, true, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.To, "archive") {
		t.Errorf("expected archive path, got %s", result.To)
	}
}

func TestArchiveMoveNonexistent(t *testing.T) {
	dir := setupArchiveTestDir(t)
	now := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)

	_, err := doArchiveMove(filepath.Join(dir, "nonexistent.md"), dir, false, now)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
