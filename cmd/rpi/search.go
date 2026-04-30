package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/A-NGJ/rpi/internal/search"
	"github.com/spf13/cobra"
)

var (
	searchType           string
	searchLimit          int
	searchIncludeArchive bool
	searchMinScore       float64
	searchWarmup         bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Semantically search .rpi/ artifacts via the optional qmd backend",
	Long: `Semantically search .rpi/ artifacts for prior work related to a
natural-language query. Returns a four-state status (ok / empty /
backend_error / backend_unavailable) with ranked hits or an actionable
hint.

Requires qmd to be installed and warmed up. When qmd is absent or its
models are not yet downloaded, the response carries a hint so callers
can either install qmd or run --warmup before retrying.

Run 'rpi search --warmup' once after installing qmd to spawn the qmd
HTTP MCP daemon and trigger the one-time ~2 GB GGUF model download.`,
	Example: `  # Run a query (qmd must be running and warmed up)
  rpi search "session resume strategies"

  # Filter by artifact type and minimum relevance
  rpi search --type design --min-score 0.5 "auth flow"

  # First-time setup: spawn the qmd daemon and download models
  rpi search --warmup`,
	RunE: runSearch,
}

func init() {
	addRpiDirFlag(searchCmd)
	searchCmd.Flags().StringVar(&searchType, "type", "", "Filter by artifact type (research, design, plan, spec, diagnosis, review)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 0, "Maximum number of hits (default 5, max 20)")
	searchCmd.Flags().BoolVar(&searchIncludeArchive, "include-archive", false, "Include archived artifacts in results")
	searchCmd.Flags().Float64Var(&searchMinScore, "min-score", 0, "Minimum relevance score (0.0-1.0)")
	searchCmd.Flags().BoolVar(&searchWarmup, "warmup", false, "Spawn the qmd MCP daemon and download models (one-time setup)")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if searchWarmup {
		return runWarmup(ctx)
	}

	if len(args) == 0 {
		return fmt.Errorf("query required (or pass --warmup for first-time setup)")
	}

	resp := searchQueryFn(ctx, rpiDirFlag, search.SearchParams{
		Query:          args[0],
		Type:           searchType,
		Limit:          searchLimit,
		IncludeArchive: searchIncludeArchive,
		MinScore:       searchMinScore,
	}, search.QueryOptions{})

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

// searchQueryFn is a package-level seam so tests can substitute a stub for
// search.Query without spinning up qmd.
var searchQueryFn = search.Query

// warmupPollTimeout and warmupPollInterval are package vars (not constants)
// so tests can shorten them; production keeps the documented 30s budget.
var (
	warmupPollTimeout  = 30 * time.Second
	warmupPollInterval = 500 * time.Millisecond
)

// warmupExecFn is the os/exec seam for runWarmup. Tests substitute a stub
// to assert the warmup flow without actually launching qmd.
var warmupExecFn = realWarmupExec

type warmupCommand struct {
	name string
	args []string
}

// realWarmupExec runs the daemon-start and warmup-query commands against
// the real qmd binary. The daemon is spawned with Start (not Run) so it
// outlives the rpi process; the warmup query is run synchronously so the
// user knows when models are ready.
func realWarmupExec(ctx context.Context, c warmupCommand, detached bool) error {
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if detached {
		// Don't wait — let the daemon outlive this process.
		return cmd.Start()
	}
	return cmd.Run()
}

func runWarmup(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "rpi-search: starting qmd MCP daemon...")
	if err := warmupExecFn(ctx, warmupCommand{
		name: "qmd",
		args: []string{"mcp", "--http", "--daemon"},
	}, true); err != nil {
		return fmt.Errorf("spawn qmd daemon: %w", err)
	}

	// Give the daemon a moment to bind its socket. qmd's docs say the HTTP
	// listener comes up quickly; we poll qmd status to confirm.
	deadline := time.Now().Add(warmupPollTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(warmupPollInterval):
		}
		c := search.NewClient()
		state, err := c.Status(ctx)
		if err == nil && state.DaemonRunning {
			break
		}
	}

	fmt.Fprintln(os.Stderr, "rpi-search: triggering model download (this may take several minutes on first run)...")
	collectionName, err := search.NewClient().EnsureCollection(ctx, rpiDirFlag)
	if err != nil {
		return fmt.Errorf("bootstrap collection: %w", err)
	}
	if err := warmupExecFn(ctx, warmupCommand{
		name: "qmd",
		args: []string{"query", "warmup", "-c", collectionName, "-n", "1"},
	}, false); err != nil {
		return fmt.Errorf("warmup query: %w", err)
	}
	fmt.Fprintln(os.Stderr, "rpi-search: warmup complete")
	return nil
}
