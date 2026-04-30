package search

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// runner is the exec indirection used by all qmd shell-outs. Tests substitute
// a stub via WithRunner; production code uses defaultRunner.
type runner func(ctx context.Context, name string, args ...string) ([]byte, error)

func defaultRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// Client wraps qmd command execution with a configurable runner so tests can
// substitute stubs without touching PATH.
type Client struct {
	run runner
}

// NewClient returns a Client that uses the real qmd binary on PATH.
func NewClient() *Client {
	return &Client{run: defaultRunner}
}

// WithRunner returns a copy of c with the runner replaced. Used by tests.
func (c *Client) WithRunner(r runner) *Client {
	return &Client{run: r}
}

// availabilityCache is a process-level cache of IsAvailable's result so the
// PATH lookup runs at most once per process.
type availabilityCache struct {
	once   sync.Once
	result bool
}

var availability = &availabilityCache{}

// IsAvailable reports whether the qmd binary is on PATH. The result is cached
// for the process lifetime — callers that want to retry after install must
// restart the process.
func (c *Client) IsAvailable(ctx context.Context) bool {
	availability.once.Do(func() {
		_, err := c.run(ctx, "qmd", "--version")
		availability.result = err == nil
	})
	return availability.result
}

// resetAvailabilityCache is a test-only helper to clear the cached result
// between table-driven cases.
func resetAvailabilityCache() {
	availability = &availabilityCache{}
}

// BackendState captures the readiness of the qmd backend for a Query call.
// All three booleans must be true for a search to proceed end-to-end.
type BackendState struct {
	Installed     bool   `json:"installed"`
	DaemonRunning bool   `json:"daemon_running"`
	ModelsReady   bool   `json:"models_ready"`
	RawStatus     string `json:"raw_status,omitempty"`
}

var (
	mcpRunningRe  = regexp.MustCompile(`(?i)\bMCP\s*:\s*running\b`)
	modelsReadyRe = regexp.MustCompile(`(?i)\b(models?\s*:\s*(ready|loaded)|all\s+models\s+(present|ready|loaded))\b`)
)

// Status probes qmd for daemon and model readiness. The returned BackendState
// always carries RawStatus when qmd was reachable, so callers can surface
// diagnostic context even when parsing fails. An error is returned only when
// qmd itself is missing or the runner fails to invoke it; in that case
// Installed is false.
func (c *Client) Status(ctx context.Context) (BackendState, error) {
	if !c.IsAvailable(ctx) {
		return BackendState{Installed: false}, nil
	}

	out, err := c.run(ctx, "qmd", "status")
	state := BackendState{Installed: true, RawStatus: string(out)}
	if err != nil {
		// qmd exists on PATH but the status command failed — return what we
		// have so the caller can map it to backend_error{parse}.
		return state, err
	}

	state.DaemonRunning = mcpRunningRe.MatchString(state.RawStatus)
	state.ModelsReady = modelsReadyRe.MatchString(state.RawStatus) || modelsCachedOnDisk()
	return state, nil
}

// modelsCachedOnDisk is a defensive backstop for when qmd's status output
// doesn't explicitly enumerate model readiness — we check the documented
// cache location for any .gguf file.
func modelsCachedOnDisk() bool {
	cacheRoot := os.Getenv("XDG_CACHE_HOME")
	if cacheRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		cacheRoot = filepath.Join(home, ".cache")
	}
	modelsDir := filepath.Join(cacheRoot, "qmd", "models")
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".gguf") {
			return true
		}
	}
	return false
}
