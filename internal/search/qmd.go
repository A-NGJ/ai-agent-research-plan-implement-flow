package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// CollectionContext is the description attached to RPI's qmd collection. qmd
// returns this alongside hits to help its reranker score relevance.
const CollectionContext = "RPI artifacts: research, designs, behavioral specs, plans, diagnoses"

// CollectionEntry is qmd's per-collection metadata. We accept both `pwd`
// (matching the SDK shape) and `path` (in case the CLI uses that key) when
// parsing list output, so a future qmd schema tweak doesn't break drift
// detection.
type CollectionEntry struct {
	Name string `json:"name"`
	Pwd  string `json:"pwd,omitempty"`
	Path string `json:"path,omitempty"`
}

// AbsPath returns whichever path field qmd populated.
func (e CollectionEntry) AbsPath() string {
	if e.Pwd != "" {
		return e.Pwd
	}
	return e.Path
}

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

// EnsureCollection registers the project's .rpi/ as a qmd collection,
// repairing path drift if the same name is registered to a stale path. The
// returned name is what should be passed to subsequent qmd calls
// (`-c <name>`).
func (c *Client) EnsureCollection(ctx context.Context, rpiDir string) (string, error) {
	name, err := CollectionName(rpiDir)
	if err != nil {
		return "", err
	}
	absRpi, err := filepath.Abs(rpiDir)
	if err != nil {
		return "", fmt.Errorf("resolve rpiDir: %w", err)
	}

	entries, err := c.listCollections(ctx)
	if err != nil {
		return "", fmt.Errorf("list collections: %w", err)
	}

	var existing *CollectionEntry
	for i := range entries {
		if entries[i].Name == name {
			existing = &entries[i]
			break
		}
	}

	switch {
	case existing == nil:
		// Not registered — create.
		if err := c.addCollection(ctx, absRpi, name); err != nil {
			return "", fmt.Errorf("add collection: %w", err)
		}
	case existing.AbsPath() != "" && !samePath(existing.AbsPath(), absRpi):
		// Path drift — repair by removing and re-adding.
		fmt.Fprintf(os.Stderr, "rpi-search: drift detected for collection %q (was %q, now %q); repairing\n",
			name, existing.AbsPath(), absRpi)
		if _, err := c.run(ctx, "qmd", "collection", "remove", name); err != nil {
			return "", fmt.Errorf("remove stale collection: %w", err)
		}
		if err := c.addCollection(ctx, absRpi, name); err != nil {
			return "", fmt.Errorf("re-add collection: %w", err)
		}
	default:
		// Already registered with the right path; nothing to do.
	}

	return name, nil
}

// listCollections invokes `qmd collection list --json` and returns the parsed
// entries. Empty output (no registered collections) is a valid case and
// returns an empty slice without error.
func (c *Client) listCollections(ctx context.Context) ([]CollectionEntry, error) {
	out, err := c.run(ctx, "qmd", "collection", "list", "--json")
	if err != nil {
		return nil, fmt.Errorf("qmd collection list: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "[]" || trimmed == "null" {
		return nil, nil
	}
	var entries []CollectionEntry
	if err := json.Unmarshal([]byte(trimmed), &entries); err != nil {
		return nil, fmt.Errorf("parse qmd collection list output: %w", err)
	}
	return entries, nil
}

// addCollection runs `qmd collection add` followed by `qmd context add` so
// every newly-registered collection carries the descriptive context qmd's
// reranker uses to bias relevance.
func (c *Client) addCollection(ctx context.Context, absRpi, name string) error {
	if out, err := c.run(ctx, "qmd", "collection", "add", absRpi, "--name", name); err != nil {
		return fmt.Errorf("qmd collection add: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	if out, err := c.run(ctx, "qmd", "context", "add", "qmd://"+name, CollectionContext); err != nil {
		return fmt.Errorf("qmd context add: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// samePath compares two paths as canonical absolute forms. Symlink resolution
// is best-effort — if either side fails to resolve we fall back to a literal
// comparison so we don't false-trigger drift repair.
func samePath(a, b string) bool {
	canon := func(p string) string {
		abs, err := filepath.Abs(p)
		if err != nil {
			return p
		}
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			return resolved
		}
		return abs
	}
	return canon(a) == canon(b)
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
