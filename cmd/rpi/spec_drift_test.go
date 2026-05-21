package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSpecDriftCmd_RegisteredOnRoot(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "spec-drift" {
			return
		}
	}
	t.Error("specDriftCmd not registered on rootCmd")
}

func TestSpecDriftScanCmd_RegisteredOnParent(t *testing.T) {
	for _, cmd := range specDriftCmd.Commands() {
		if cmd.Name() == "scan" {
			return
		}
	}
	t.Error("specDriftScanCmd not registered on specDriftCmd")
}

func TestRunSpecDriftScan_EmitsJSON(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.Mkdir(specsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeSpec(t, filepath.Join(specsDir, "a.md"), "a", "2026-05-15T10:00:00+02:00")
	writeSpec(t, filepath.Join(specsDir, "b.md"), "b", "2026-05-15T10:00:00+02:00")

	withSpecDriftFlags(t, 30, 0.5, 3.0, specsDir)
	var buf bytes.Buffer
	specDriftScanCmd.SetOut(&buf)
	defer specDriftScanCmd.SetOut(nil)

	if err := runSpecDriftScan(specDriftScanCmd, nil); err != nil {
		t.Fatalf("runSpecDriftScan: %v", err)
	}

	var records []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d: %s", len(records), buf.String())
	}
	for _, r := range records {
		if _, ok := r["path"].(string); !ok {
			t.Errorf("record missing 'path' string: %v", r)
		}
		if _, ok := r["signals"]; !ok {
			t.Errorf("record missing 'signals' key: %v", r)
		}
	}
}

func TestRunSpecDriftScan_FlagOverrides(t *testing.T) {
	t.Setenv("PATH", "")
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.Mkdir(specsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Recent spec — with default --stale-days=30 it should not be stale.
	writeSpec(t, filepath.Join(specsDir, "fresh.md"), "fresh", "2026-05-19T10:00:00+02:00")

	// Default thresholds: no stale signal.
	withSpecDriftFlags(t, 30, 0.5, 3.0, specsDir)
	var buf bytes.Buffer
	specDriftScanCmd.SetOut(&buf)
	defer specDriftScanCmd.SetOut(nil)

	if err := runSpecDriftScan(specDriftScanCmd, nil); err != nil {
		t.Fatalf("runSpecDriftScan default: %v", err)
	}
	if hasSignal(t, buf.Bytes(), "stale_last_updated") {
		t.Errorf("default thresholds should not flag fresh spec as stale: %s", buf.String())
	}

	// --stale-days=0 → every spec with a parseable last_updated is stale.
	buf.Reset()
	withSpecDriftFlags(t, 0, 0.5, 3.0, specsDir)
	if err := runSpecDriftScan(specDriftScanCmd, nil); err != nil {
		t.Fatalf("runSpecDriftScan stale-days=0: %v", err)
	}
	if !hasSignal(t, buf.Bytes(), "stale_last_updated") {
		t.Errorf("--stale-days=0 should flag fresh spec as stale (via git-unavailable): %s", buf.String())
	}
}

// --- helpers ---

func writeSpec(t *testing.T, path, feature, lastUpdated string) {
	t.Helper()
	body := "---\nfeature: " + feature + "\nlast_updated: " + lastUpdated + "\n---\n\n# " + feature + "\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func withSpecDriftFlags(t *testing.T, days int, low, high float64, specsDir string) {
	t.Helper()
	oldDays, oldLow, oldHigh, oldDir := staleDaysFlag, ratioLowFlag, ratioHighFlag, specsDirFlag
	staleDaysFlag, ratioLowFlag, ratioHighFlag, specsDirFlag = days, low, high, specsDir
	t.Cleanup(func() {
		staleDaysFlag, ratioLowFlag, ratioHighFlag, specsDirFlag = oldDays, oldLow, oldHigh, oldDir
	})
}

func hasSignal(t *testing.T, data []byte, name string) bool {
	t.Helper()
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("invalid JSON: %v\ndata: %s", err, data)
	}
	for _, r := range records {
		sigs, _ := r["signals"].([]any)
		for _, s := range sigs {
			m, _ := s.(map[string]any)
			if m["name"] == name {
				return true
			}
		}
	}
	return false
}
