package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/A-NGJ/rpi/internal/search"
)

// resetSearchFlags returns a cleanup func that restores command flags and
// the searchQueryFn seam to their package defaults.
func resetSearchFlags(t *testing.T) func() {
	t.Helper()
	origType := searchType
	origLimit := searchLimit
	origInclude := searchExcludeArchive
	origMin := searchMinScore
	origWarmup := searchWarmup
	origQueryFn := searchQueryFn
	return func() {
		searchType = origType
		searchLimit = origLimit
		searchExcludeArchive = origInclude
		searchMinScore = origMin
		searchWarmup = origWarmup
		searchQueryFn = origQueryFn
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	defer resetSearchFlags(t)()
	searchWarmup = false

	err := runSearch(searchCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no query and no --warmup")
	}
	if !strings.Contains(err.Error(), "query required") {
		t.Errorf("expected 'query required' error, got %q", err.Error())
	}
}

func TestSearchFlagsMarshalIntoParams(t *testing.T) {
	defer resetSearchFlags(t)()

	var captured search.SearchParams
	searchQueryFn = func(ctx context.Context, rpiDir string, p search.SearchParams, opts search.QueryOptions) search.SearchResponse {
		captured = p
		return search.SearchResponse{Status: search.StatusEmpty}
	}

	searchType = "design"
	searchLimit = 10
	searchExcludeArchive = true
	searchMinScore = 0.6

	if err := runSearch(searchCmd, []string{"my query"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.Query != "my query" {
		t.Errorf("Query: got %q, want %q", captured.Query, "my query")
	}
	if captured.Type != "design" {
		t.Errorf("Type: got %q, want %q", captured.Type, "design")
	}
	if captured.Limit != 10 {
		t.Errorf("Limit: got %d, want 10", captured.Limit)
	}
	if !captured.ExcludeArchive {
		t.Error("ExcludeArchive: got false, want true")
	}
	if captured.MinScore != 0.6 {
		t.Errorf("MinScore: got %v, want 0.6", captured.MinScore)
	}
}

// recordingExec records each warmup command invocation so the test can
// assert that the daemon-start command was issued.
type recordingExec struct {
	calls []warmupCommand
}

func (r *recordingExec) fn(ctx context.Context, c warmupCommand, detached bool) error {
	r.calls = append(r.calls, c)
	return nil
}

func TestSearchWarmupSpawnsDaemonCommand(t *testing.T) {
	defer resetSearchFlags(t)()

	// Stub the exec seam and shorten the polling so the test exits fast.
	rec := &recordingExec{}
	origExec := warmupExecFn
	origTimeout := warmupPollTimeout
	origInterval := warmupPollInterval
	warmupExecFn = rec.fn
	warmupPollTimeout = 5 * time.Millisecond
	warmupPollInterval = 1 * time.Millisecond
	t.Cleanup(func() {
		warmupExecFn = origExec
		warmupPollTimeout = origTimeout
		warmupPollInterval = origInterval
	})

	searchWarmup = true
	// runSearch will attempt EnsureCollection after polling — that uses real
	// qmd via exec, which probably returns "not found" in CI. We don't
	// assert success; we only check that the daemon-start command was
	// invoked first. Use a very short context so the call returns quickly
	// even if EnsureCollection blocks.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	searchCmd.SetContext(ctx)

	_ = runSearch(searchCmd, nil) // ignore terminal error from EnsureCollection

	if len(rec.calls) == 0 {
		t.Fatal("expected at least one warmup exec call")
	}
	first := rec.calls[0]
	if first.name != "qmd" {
		t.Errorf("expected qmd command, got %q", first.name)
	}
	if len(first.args) < 3 || first.args[0] != "mcp" || first.args[1] != "--http" || first.args[2] != "--daemon" {
		t.Errorf("expected 'qmd mcp --http --daemon', got %q %v", first.name, first.args)
	}
}
