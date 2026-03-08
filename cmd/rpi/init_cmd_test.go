package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesAllDirs(t *testing.T) {
	dir := t.TempDir()
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"research", "designs", "structures", "tickets",
		"plans", "specs", "reviews", "archive",
	}
	for _, d := range expected {
		path := filepath.Join(dir, d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("directory %s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}

	output := buf.String()
	if strings.Count(output, "created ") != 8 {
		t.Errorf("expected 8 'created' lines, got output:\n%s", output)
	}
}

func TestInitIdempotent(t *testing.T) {
	dir := t.TempDir()
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	// First run
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Second run
	buf.Reset()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("second run error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already initialized") {
		t.Errorf("expected 'already initialized', got:\n%s", output)
	}
}

func TestInitPartial(t *testing.T) {
	dir := t.TempDir()
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	// Pre-create some directories
	os.MkdirAll(filepath.Join(dir, "research"), 0755)
	os.MkdirAll(filepath.Join(dir, "plans"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should create 5 missing dirs (designs, structures, tickets, specs, reviews)
	if strings.Count(output, "created ") != 5 {
		t.Errorf("expected 5 'created' lines, got output:\n%s", output)
	}
	if strings.Contains(output, "research") {
		t.Error("should not have created already-existing 'research' dir")
	}
	if strings.Contains(output, "already initialized") {
		t.Error("should not print 'already initialized' when some dirs were created")
	}
}
