package main

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/git"
)

func TestGitContextIntegration(t *testing.T) {
	// Verify we're in a git repo
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		t.Skip("not in a git repository")
	}

	ctx, err := git.GatherContext()
	if err != nil {
		t.Fatalf("GatherContext failed: %v", err)
	}

	// Verify JSON marshaling works
	data, err := json.Marshal(ctx)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Verify it unmarshals back to a valid structure
	var decoded git.Context
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Branch should not be empty (either a branch name or "HEAD")
	if decoded.Branch == "" {
		t.Error("branch should not be empty")
	}

	// Status arrays should be initialized (not nil)
	if decoded.Status.Staged == nil {
		t.Error("staged should be initialized, not nil")
	}
	if decoded.Status.Modified == nil {
		t.Error("modified should be initialized, not nil")
	}
	if decoded.Status.Untracked == nil {
		t.Error("untracked should be initialized, not nil")
	}
}

func TestChangedFilesIntegration(t *testing.T) {
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		t.Skip("not in a git repository")
	}

	files, err := git.ChangedFiles()
	if err != nil {
		t.Fatalf("ChangedFiles failed: %v", err)
	}

	// Should return a slice (possibly empty), never nil
	if files == nil {
		t.Error("ChangedFiles should return empty slice, not nil")
	}
}
