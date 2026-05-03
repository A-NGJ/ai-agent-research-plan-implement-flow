package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Pipeline is the seam Query orchestrates against. Production wires it to
// the real Client; tests substitute stubs to drive specific spec scenarios.
type Pipeline interface {
	IsAvailable(ctx context.Context) bool
	Status(ctx context.Context) (BackendState, error)
	EnsureCollection(ctx context.Context, rpiDir string) (string, error)
	Refresh(ctx context.Context, collectionName string) (RefreshResult, error)
}

// QueryOptions configures Query's pipeline + daemon. Either field may be nil
// to use the production default.
type QueryOptions struct {
	Pipeline Pipeline
	Daemon   Daemon
}

// Query executes a semantic search end-to-end. It never returns an error;
// every failure mode is encoded in the SearchResponse so MCP/CLI callers
// receive a single well-formed JSON object.
func Query(ctx context.Context, rpiDir string, params SearchParams, opts QueryOptions) SearchResponse {
	pipeline := opts.Pipeline
	if pipeline == nil {
		pipeline = NewClient()
	}
	daemon := opts.Daemon
	if daemon == nil {
		daemon = NewDaemonClient()
	}

	if !pipeline.IsAvailable(ctx) {
		return SearchResponse{
			Status:      StatusBackendUnavailable,
			Reason:      "qmd not installed",
			InstallHint: installHint,
			Fallback:    fallbackHint,
		}
	}

	state, err := pipeline.Status(ctx)
	if err != nil {
		return backendError(StageParse, fmt.Sprintf("qmd status failed: %v", err),
			"Run 'qmd status' manually to diagnose")
	}
	if !state.DaemonRunning {
		return backendError(StageDaemonNotRunning, "qmd MCP daemon is not running",
			"Run 'rpi search --warmup' to start the qmd daemon")
	}
	if !state.ModelsReady {
		return backendError(StageModelsNotReady, "qmd models are not yet downloaded",
			"Run 'rpi search --warmup' to download qmd models (one-time, ~2 GB)")
	}

	collectionName, err := pipeline.EnsureCollection(ctx, rpiDir)
	if err != nil {
		return backendError(StageUpdate, fmt.Sprintf("collection bootstrap failed: %v", err),
			"Check 'qmd collection list' and verify .rpi/ is accessible")
	}

	refresh, err := pipeline.Refresh(ctx, collectionName)
	if err != nil {
		var rErr *RefreshError
		if errors.As(err, &rErr) {
			return backendError(rErr.Stage, rErr.Message,
				"Re-run after fixing the underlying issue")
		}
		return backendError(StageUpdate, err.Error(), "")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// qmd's MCP `query` tool expects a `searches` array of typed sub-queries
	// (lex / vec / hyde) — see node_modules/@tobilu/qmd/dist/mcp/server.js. We
	// send both a `lex` and a `vec` sub-query for the same string: lex catches
	// exact keyword matches, vec catches semantic matches. The first sub-query
	// gets 2× weight per qmd's docs; we put `vec` first so semantic relevance
	// dominates ranking, with `lex` as a recall booster for exact keywords.
	args := map[string]any{
		"searches": []map[string]any{
			{"type": "vec", "query": params.Query},
			{"type": "lex", "query": params.Query},
		},
		"collections": []string{collectionName},
		"limit":       limit,
	}
	if params.MinScore > 0 {
		args["minScore"] = params.MinScore
	}

	raw, err := daemon.Call(ctx, "query", args)
	if err != nil {
		stage := StageQuery
		hint := "Check 'qmd status' and try 'rpi search --warmup' to restart the daemon"
		if errors.Is(err, errDaemonNotRunning) {
			stage = StageDaemonNotRunning
			hint = "Run 'rpi search --warmup' to start the qmd daemon"
		} else if errors.Is(err, errParse) {
			stage = StageParse
			hint = "qmd returned an unexpected response — file an issue if this persists"
		}
		return backendError(stage, err.Error(), hint)
	}

	hits, parseErr := parseQmdHits(raw)
	if parseErr != nil {
		return backendError(StageParse, fmt.Sprintf("parse qmd response: %v", parseErr),
			"qmd's --json output may have changed; check qmd version compatibility")
	}

	hits = filterHits(hits, params)

	resp := SearchResponse{Status: StatusOK, Hits: hits}
	if len(hits) == 0 {
		resp.Status = StatusEmpty
	}
	if len(refresh.Warnings) > 0 {
		resp.Warnings = refresh.Warnings
	}
	return resp
}

// backendError builds a uniform StatusBackendError response with fallback.
func backendError(stage ErrorStage, msg, hint string) SearchResponse {
	return SearchResponse{
		Status: StatusBackendError,
		Error: &SearchError{
			Stage:   stage,
			Message: msg,
			Hint:    hint,
		},
		Fallback: fallbackHint,
	}
}

// qmdHit is the shape we extract from qmd's tool result. qmd's MCP `query`
// tool returns hits with field names `file` (collection-relative path),
// `title`, `score`, `snippet`, `context`, `docid`. We accept aliases
// (`path`, `displayPath`) so callers driving the parser with different
// shapes (or future qmd versions) still work.
type qmdHit struct {
	File        string  `json:"file,omitempty"`
	Path        string  `json:"path,omitempty"`
	DisplayPath string  `json:"displayPath,omitempty"`
	Title       string  `json:"title"`
	Score       float64 `json:"score"`
	Snippet     string  `json:"snippet,omitempty"`
	Text        string  `json:"text,omitempty"`
	Context     string  `json:"context,omitempty"`
}

// parseQmdHits accepts either a bare JSON array of hits or a wrapper object
// with a `results` (or `hits`) array. qmd's exact shape varies by tool; we
// stay permissive so this layer survives schema drift.
func parseQmdHits(raw json.RawMessage) ([]Hit, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil
	}

	var arr []qmdHit
	if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
		// Try the wrapper-object shape.
		var wrapper struct {
			Results []qmdHit `json:"results"`
			Hits    []qmdHit `json:"hits"`
		}
		if werr := json.Unmarshal([]byte(trimmed), &wrapper); werr != nil {
			return nil, fmt.Errorf("neither array nor wrapper: %v / %v", err, werr)
		}
		if len(wrapper.Results) > 0 {
			arr = wrapper.Results
		} else {
			arr = wrapper.Hits
		}
	}

	out := make([]Hit, 0, len(arr))
	for _, h := range arr {
		path := h.Path
		if path == "" {
			path = h.File
		}
		if path == "" {
			path = h.DisplayPath
		}
		snippet := h.Snippet
		if snippet == "" {
			snippet = h.Text
		}
		out = append(out, Hit{
			Path:    path,
			Type:    inferType(path),
			Title:   h.Title,
			Score:   h.Score,
			Snippet: snippet,
			Context: h.Context,
		})
	}
	return out, nil
}

// typeTokens maps every directory name we might see in a path segment
// (plural folder names for active artifacts, singular folder names for
// archived ones) to the canonical singular type the spec contract uses.
var typeTokens = map[string]string{
	"research":  "research",
	"designs":   "design",
	"design":    "design",
	"plans":     "plan",
	"plan":      "plan",
	"specs":     "spec",
	"spec":      "spec",
	"diagnoses": "diagnosis",
	"diagnosis": "diagnosis",
	"reviews":   "review",
	"review":    "review",
}

// inferType maps an artifact path to its canonical type by scanning each
// path segment for a known directory name. Works for both:
//
//   - Filesystem-relative paths like ".rpi/designs/foo.md" or
//     ".rpi/archive/designs/foo.md".
//   - qmd's collection-prefixed paths like
//     "rpi-myrepo-abc123/specs/foo.md" or
//     "rpi-myrepo-abc123/archive/2026-04/plan/foo.md" (qmd's archive
//     subdirectory uses the singular form).
//
// Returns empty string when no recognizable type segment appears.
func inferType(path string) string {
	clean := filepath.ToSlash(path)
	for _, seg := range strings.Split(clean, "/") {
		if t, ok := typeTokens[seg]; ok {
			return t
		}
	}
	return ""
}

// filterHits applies caller-side filters that qmd doesn't natively know
// about: per-project archive scope, RPI artifact-type taxonomy, the score
// floor (in case the daemon ignores min_score), and the result limit.
func filterHits(hits []Hit, params SearchParams) []Hit {
	out := make([]Hit, 0, len(hits))
	for _, h := range hits {
		if params.ExcludeArchive && isArchivePath(h.Path) {
			continue
		}
		if params.Type != "" && h.Type != params.Type {
			continue
		}
		if params.MinScore > 0 && h.Score < params.MinScore {
			continue
		}
		out = append(out, h)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// isArchivePath reports whether the path lives under any /archive/ segment.
// Matches both filesystem-relative ".rpi/archive/..." and qmd's collection-
// prefixed "<collection>/archive/..." shapes.
func isArchivePath(path string) bool {
	clean := filepath.ToSlash(path)
	for _, seg := range strings.Split(clean, "/") {
		if seg == "archive" {
			return true
		}
	}
	return false
}
