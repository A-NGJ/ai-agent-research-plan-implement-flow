package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServeCmd_RegisteredOnRoot(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "serve" {
			return
		}
	}
	t.Error("serveCmd not registered on rootCmd")
}

// extractText returns the text content from an MCP tool result.
func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// --- Phase 1: No-param tool tests ---

func TestHandleGitContext_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitContext(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if _, ok := m["branch"]; !ok {
		t.Error("expected 'branch' key in git context")
	}
}

func TestHandleGitChangedFiles_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitChangedFiles(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var files []string
	if err := json.Unmarshal([]byte(text), &files); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
}

func TestHandleGitSensitiveCheck_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitSensitiveCheck(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var matches []any
	if err := json.Unmarshal([]byte(text), &matches); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
}

func TestHandleGitGitignoreCheck_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitGitignoreCheck(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var matches []any
	if err := json.Unmarshal([]byte(text), &matches); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
}

func TestHandleArchiveScan_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	oldFlag := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldFlag }()

	result, _, err := handleArchiveScan(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var results []archiveScanResult
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results in empty dir, got %d", len(results))
	}
}

func TestHandleArchiveScan_WithArchivable(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldFlag }()

	result, _, err := handleArchiveScan(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var results []archiveScanResult
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 archivable artifacts, got %d", len(results))
	}
}

// --- Phase 2: Parameterized tool tests ---

func TestHandleScan_WithTypeFilter(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldFlag }()

	result, _, err := handleScan(context.Background(), nil, scanInput{Type: "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var results []scanner.ArtifactInfo
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	for _, r := range results {
		if r.Type != "plan" {
			t.Errorf("expected type 'plan', got %q", r.Type)
		}
	}
}

func TestHandleFrontmatterGet_AllFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: draft\ntopic: test\n---\n# Test\n"), 0644)

	result, _, err := handleFrontmatterGet(context.Background(), nil, frontmatterGetInput{File: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if m["status"] != "draft" {
		t.Errorf("expected status 'draft', got %v", m["status"])
	}
	if m["topic"] != "test" {
		t.Errorf("expected topic 'test', got %v", m["topic"])
	}
}

func TestHandleFrontmatterGet_SingleField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: active\ntopic: foo\n---\n# Test\n"), 0644)

	result, _, err := handleFrontmatterGet(context.Background(), nil, frontmatterGetInput{File: path, Field: "status"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var val string
	if err := json.Unmarshal([]byte(text), &val); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if val != "active" {
		t.Errorf("expected 'active', got %q", val)
	}
}

func TestHandleFrontmatterTransition_Invalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: draft\n---\n# Test\n"), 0644)

	_, _, err := handleFrontmatterTransition(context.Background(), nil, frontmatterTransitionInput{
		File:   path,
		Status: "complete",
	})
	if err == nil {
		t.Fatal("expected error for invalid transition draft→complete")
	}
}

func TestHandleFrontmatterTransition_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: draft\n---\n# Test\n"), 0644)

	result, _, err := handleFrontmatterTransition(context.Background(), nil, frontmatterTransitionInput{
		File:   path,
		Status: "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]string
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["new_status"] != "active" {
		t.Errorf("expected new_status 'active', got %q", m["new_status"])
	}

	// Verify file was actually updated
	doc, _ := frontmatter.Parse(path)
	if doc.Frontmatter["status"] != "active" {
		t.Errorf("file not updated: status is %v", doc.Frontmatter["status"])
	}
}

// setupTestTemplates creates a minimal template directory for scaffold tests.
func setupTestTemplates(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "research.tmpl"), []byte(`---
date: {{.Date}}
topic: "{{.Topic}}"
status: draft
---

# Research: {{.Topic}}

## Summary
`), 0644)
	old := templatesDirFlag
	templatesDirFlag = dir
	return func() { templatesDirFlag = old }
}

func TestHandleScaffold_Preview(t *testing.T) {
	defer setupTestTemplates(t)()

	result, _, err := handleScaffold(context.Background(), nil, scaffoldInput{
		Type:  "research",
		Topic: "test topic",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]string
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if m["content"] == "" {
		t.Error("expected non-empty content")
	}
	if m["filename"] == "" {
		t.Error("expected non-empty filename")
	}
}

func TestHandleScaffold_Write(t *testing.T) {
	defer setupTestTemplates(t)()

	dir := t.TempDir()
	oldRpi := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldRpi }()

	result, _, err := handleScaffold(context.Background(), nil, scaffoldInput{
		Type:  "research",
		Topic: "test write",
		Write: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]string
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	path := m["path"]
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %s: %v", path, err)
	}
}

func TestHandleScaffold_InvalidType(t *testing.T) {
	_, _, err := handleScaffold(context.Background(), nil, scaffoldInput{
		Type:  "invalid",
		Topic: "test",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestHandleExtract_Section(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: draft\n---\n# Doc\n\n## Summary\nThis is the summary.\n\n## Details\nMore info.\n"), 0644)

	result, _, err := handleExtract(context.Background(), nil, extractInput{Path: path, Section: "Summary"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]string
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if m["content"] == "" {
		t.Error("expected non-empty content")
	}
	if m["section"] != "Summary" {
		t.Errorf("expected section 'Summary', got %q", m["section"])
	}
}

func TestHandleExtract_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: draft\n---\n# Doc\n\n## Summary\nContent.\n"), 0644)

	_, _, err := handleExtract(context.Background(), nil, extractInput{Path: path, Section: "Nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing section")
	}
}

func TestHandleExtractListSections(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\nstatus: draft\n---\n# Doc\n\n## Summary\nContent.\n\n## Details\nMore.\n"), 0644)

	result, _, err := handleExtractListSections(context.Background(), nil, extractListSectionsInput{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	sections, ok := m["sections"].([]any)
	if !ok {
		t.Fatal("expected sections array")
	}
	if len(sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(sections))
	}
}

// --- Phase 3: Integration test ---

func TestIntegration_AllToolsRegistered(t *testing.T) {
	ctx := context.Background()
	server := newRPIServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer clientSession.Close()

	res, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	expectedTools := []string{
		"rpi_git_context",
		"rpi_git_changed_files",
		"rpi_git_sensitive_check",
		"rpi_git_gitignore_check",
		"rpi_archive_scan",
		"rpi_scan",
		"rpi_scaffold",
		"rpi_frontmatter_get",
		"rpi_frontmatter_set",
		"rpi_frontmatter_transition",
		"rpi_chain",
		"rpi_extract",
		"rpi_extract_list_sections",
		"rpi_verify_completeness",
		"rpi_verify_markers",
		"rpi_verify_spec",
		"rpi_context_essentials",
		"rpi_session_resume",
		"rpi_suggest_next",
		"rpi_archive_check_refs",
		"rpi_archive_move",
		"rpi_search",
		"rpi_split_score",
		"rpi_spec_drift_scan",
	}

	if len(res.Tools) != len(expectedTools) {
		t.Fatalf("expected %d tools, got %d", len(expectedTools), len(res.Tools))
	}

	got := make(map[string]string) // name -> description
	for _, tool := range res.Tools {
		got[tool.Name] = tool.Description
	}

	for _, name := range expectedTools {
		desc, ok := got[name]
		if !ok {
			t.Errorf("missing tool: %s", name)
			continue
		}
		if desc == "" {
			t.Errorf("tool %s has empty description", name)
		}
		if !strings.HasPrefix(name, "rpi_") {
			t.Errorf("tool %s missing rpi_ prefix", name)
		}
	}
}

// --- MCP Description Tests (GG-1 through GG-4) ---

func TestMCPDescription_IncludesLongText(t *testing.T) {
	// GG-1: Every MCP tool description includes the Cobra command's Long field
	s := newRPIServer()
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	tools := make(map[string]string)
	for _, tool := range res.Tools {
		tools[tool.Name] = tool.Description
	}

	// scaffold's Long contains "Types and their subdirectories"
	if desc, ok := tools["rpi_scaffold"]; !ok {
		t.Error("missing rpi_scaffold tool")
	} else if !strings.Contains(desc, "Types and their subdirectories") {
		t.Errorf("rpi_scaffold description missing Long content, got: %s", desc[:min(len(desc), 100)])
	}
}

func TestMCPDescription_IncludesExamples(t *testing.T) {
	// GG-2: Every MCP tool description includes the Cobra command's Example field
	s := newRPIServer()
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	tools := make(map[string]string)
	for _, tool := range res.Tools {
		tools[tool.Name] = tool.Description
	}

	// chain's Example contains "rpi chain .rpi/plans/"
	if desc, ok := tools["rpi_chain"]; !ok {
		t.Error("missing rpi_chain tool")
	} else if !strings.Contains(desc, "rpi chain .rpi/plans/") {
		t.Errorf("rpi_chain description missing Example content, got: %s", desc[:min(len(desc), 100)])
	}
}

func TestMCPDescription_SingleSourceOfTruth(t *testing.T) {
	// GG-3: No inline description strings — all derived from Cobra commands (should be long)
	s := newRPIServer()
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range res.Tools {
		if len(tool.Description) < 50 {
			t.Errorf("tool %s has short description (%d chars) — likely not derived from Cobra Long: %q",
				tool.Name, len(tool.Description), tool.Description)
		}
	}
}

func TestMCPDescription_FrontmatterIncludesParent(t *testing.T) {
	// GG-4: Frontmatter subcommand tools include parent command's Long field
	s := newRPIServer()
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	tools := make(map[string]string)
	for _, tool := range res.Tools {
		tools[tool.Name] = tool.Description
	}

	// frontmatter_transition should contain state transition info from parent Long
	desc := tools["rpi_frontmatter_transition"]
	for _, keyword := range []string{"draft", "active", "complete"} {
		if !strings.Contains(desc, keyword) {
			t.Errorf("rpi_frontmatter_transition description missing %q from parent Long field", keyword)
		}
	}
}

func TestHandleVerifyCompleteness(t *testing.T) {
	dir := t.TempDir()
	plan := filepath.Join(dir, "plan.md")
	os.WriteFile(plan, []byte("## Phase 1\n- [x] Do A\n- [ ] Do B\n- [x] Do C\n- [ ] Do D\n"), 0644)

	result, _, err := handleVerifyCompleteness(context.Background(), nil, verifyCompletenessInput{PlanPath: plan})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if m["total_checkboxes"] != float64(4) {
		t.Errorf("expected 4 total, got %v", m["total_checkboxes"])
	}
	if m["checked"] != float64(2) {
		t.Errorf("expected 2 checked, got %v", m["checked"])
	}
	if m["unchecked"] != float64(2) {
		t.Errorf("expected 2 unchecked, got %v", m["unchecked"])
	}
}

func TestHandleSpecDriftScan_ReturnsJSON(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.Mkdir(specsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "---\nfeature: x\nlast_updated: 2026-05-15T10:00:00+02:00\n---\n\n# X\n"
	if err := os.WriteFile(filepath.Join(specsDir, "x.md"), []byte(body), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result, _, err := handleSpecDriftScan(context.Background(), nil, specDriftScanInput{SpecsDir: specsDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var records []map[string]any
	if err := json.Unmarshal([]byte(text), &records); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d: %s", len(records), text)
	}
	if _, ok := records[0]["path"].(string); !ok {
		t.Errorf("record missing path string: %v", records[0])
	}
	if _, ok := records[0]["signals"]; !ok {
		t.Errorf("record missing signals key: %v", records[0])
	}
}
