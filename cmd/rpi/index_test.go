package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/index"
)

func setupIndexTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeIndexFile(t, dir, "main.go", `package main

func main() {}

func HandleRequest() {}

type Server struct {}
`)
	writeIndexFile(t, dir, "util.go", `package main

func helperFunc() {}
`)
	return dir
}

func writeIndexFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func buildTestIndex(t *testing.T, dir string) {
	t.Helper()
	idx, err := index.Build(dir, index.BuildOptions{})
	if err != nil {
		t.Fatalf("build index: %v", err)
	}
	rpiDir := filepath.Join(dir, ".rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := index.Save(idx, filepath.Join(rpiDir, "index.json")); err != nil {
		t.Fatalf("save index: %v", err)
	}
}

func TestIndexBuild(t *testing.T) {
	dir := setupIndexTestDir(t)

	oldPath := indexPathFlag
	indexPathFlag = dir
	defer func() { indexPathFlag = oldPath }()

	buf := new(bytes.Buffer)
	cmd := indexBuildCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Indexed") {
		t.Errorf("expected 'Indexed' in output, got: %s", output)
	}
	if !strings.Contains(output, "symbols") {
		t.Errorf("expected 'symbols' in output, got: %s", output)
	}

	// Verify index file was created.
	indexPath := filepath.Join(dir, ".rpi", "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Errorf("index file not created: %v", err)
	}
}

func TestIndexQuery(t *testing.T) {
	dir := setupIndexTestDir(t)
	buildTestIndex(t, dir)

	// Change working directory so loadIndex finds .rpi/index.json.
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "json"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexQueryCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"Handle"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []index.Symbol
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Name != "HandleRequest" {
		t.Errorf("got %q, want HandleRequest", results[0].Name)
	}
}

func TestIndexQueryMarkdown(t *testing.T) {
	dir := setupIndexTestDir(t)
	buildTestIndex(t, dir)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "md"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexQueryCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"Handle"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "| Name |") {
		t.Errorf("expected markdown table header, got: %s", output)
	}
	if !strings.Contains(output, "HandleRequest") {
		t.Errorf("expected HandleRequest in output, got: %s", output)
	}
}

func TestIndexQueryNoResults(t *testing.T) {
	dir := setupIndexTestDir(t)
	buildTestIndex(t, dir)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "json"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexQueryCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"nonexistent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []index.Symbol
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestIndexFiles(t *testing.T) {
	dir := setupIndexTestDir(t)
	buildTestIndex(t, dir)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "json"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexFilesCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []struct {
		Path     string `json:"path"`
		Language string `json:"language"`
		Symbols  int    `json:"symbols"`
	}
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(results) != 2 {
		t.Fatalf("got %d files, want 2", len(results))
	}
}

func TestIndexStatusWithIndex(t *testing.T) {
	dir := setupIndexTestDir(t)
	buildTestIndex(t, dir)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "json"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexStatusCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result index.StatusResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if !result.Exists {
		t.Error("expected Exists = true")
	}
	if result.FileCount != 2 {
		t.Errorf("FileCount = %d, want 2", result.FileCount)
	}
}

func TestIndexStatusWithoutIndex(t *testing.T) {
	dir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "json"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexStatusCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"exists": false`) {
		t.Errorf("expected exists: false, got: %s", output)
	}
}

func TestIndexStatusText(t *testing.T) {
	dir := setupIndexTestDir(t)
	buildTestIndex(t, dir)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	oldFormat := indexFormatFlag
	indexFormatFlag = "text"
	defer func() { indexFormatFlag = oldFormat }()

	buf := new(bytes.Buffer)
	cmd := indexStatusCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Index:") {
		t.Errorf("expected 'Index:' in text output, got: %s", output)
	}
	if !strings.Contains(output, "Files:") {
		t.Errorf("expected 'Files:' in text output, got: %s", output)
	}
}
