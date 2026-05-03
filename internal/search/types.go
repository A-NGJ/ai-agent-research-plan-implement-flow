// Package search provides semantic discovery of .rpi/ artifacts via the
// optional qmd backend. The package's public surface is the SearchResponse
// contract — every failure mode is encoded in Status, never as a Go error
// returned past the package boundary, so MCP/CLI callers always get a
// well-formed JSON response.
package search

// Status is the four-state response contract returned by Query.
type Status string

const (
	// StatusOK indicates a successful query with one or more hits.
	StatusOK Status = "ok"
	// StatusEmpty indicates a successful query with zero matching hits —
	// distinct from any error or unavailability state.
	StatusEmpty Status = "empty"
	// StatusBackendError indicates the backend is installed but failing.
	// The Error field carries an actionable hint.
	StatusBackendError Status = "backend_error"
	// StatusBackendUnavailable indicates the backend is not installed.
	// Callers should fall back to rpi_scan + keyword grep.
	StatusBackendUnavailable Status = "backend_unavailable"
)

// ErrorStage names which step of the query pipeline failed when Status is
// StatusBackendError. Callers can branch on this to surface specific hints.
type ErrorStage string

const (
	StageUpdate           ErrorStage = "update"
	StageEmbed            ErrorStage = "embed"
	StageQuery            ErrorStage = "query"
	StageParse            ErrorStage = "parse"
	StageModelsNotReady   ErrorStage = "models_not_ready"
	StageDaemonNotRunning ErrorStage = "daemon_not_running"
)

// SearchParams is the input contract for a search.
type SearchParams struct {
	Query          string  `json:"query"`
	Type           string  `json:"type,omitempty"`
	Limit          int     `json:"limit,omitempty"`
	ExcludeArchive bool    `json:"exclude_archive,omitempty"`
	MinScore       float64 `json:"min_score,omitempty"`
}

// Hit is a single ranked search result.
type Hit struct {
	Path    string  `json:"path"`
	Type    string  `json:"type"`
	Title   string  `json:"title"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"`
	Context string  `json:"context,omitempty"`
}

// SearchError describes a backend failure that callers can act on.
type SearchError struct {
	Stage   ErrorStage `json:"stage"`
	Message string     `json:"message,omitempty"`
	Hint    string     `json:"hint,omitempty"`
}

// SearchResponse is the unified return shape. Fields are populated based on
// Status; omitempty keeps each status's payload narrow on the wire.
type SearchResponse struct {
	Status      Status       `json:"status"`
	Hits        []Hit        `json:"hits,omitempty"`
	Warnings    []string     `json:"warnings,omitempty"`
	Error       *SearchError `json:"error,omitempty"`
	Reason      string       `json:"reason,omitempty"`
	InstallHint string       `json:"install_hint,omitempty"`
	Fallback    string       `json:"fallback,omitempty"`
}

const (
	// DefaultLimit is the default number of hits returned when SearchParams.Limit is 0.
	DefaultLimit = 5
	// MaxLimit caps SearchParams.Limit to prevent unbounded responses.
	MaxLimit = 20
)

// fallbackHint is the standard fallback message embedded in non-ok responses.
const fallbackHint = "Use rpi_scan with type filter + keyword terms from the query"

// installHint is the standard install instruction shown when qmd isn't on PATH.
const installHint = "npm install -g @tobilu/qmd, then run rpi search --warmup"
