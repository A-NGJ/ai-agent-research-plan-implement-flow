package search

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// fakePipeline drives Query through specific spec scenarios without touching
// qmd or filesystems. Each field can be customized per test.
type fakePipeline struct {
	available  bool
	state      BackendState
	stateErr   error
	collection string
	collectErr error
	refresh    RefreshResult
	refreshErr error
}

func (f *fakePipeline) IsAvailable(ctx context.Context) bool { return f.available }
func (f *fakePipeline) Status(ctx context.Context) (BackendState, error) {
	return f.state, f.stateErr
}
func (f *fakePipeline) EnsureCollection(ctx context.Context, rpiDir string) (string, error) {
	if f.collectErr != nil {
		return "", f.collectErr
	}
	if f.collection == "" {
		return "rpi-test-abcdef", nil
	}
	return f.collection, nil
}
func (f *fakePipeline) Refresh(ctx context.Context, name string) (RefreshResult, error) {
	return f.refresh, f.refreshErr
}

// fakeDaemon returns a canned response or error for one Call.
type fakeDaemon struct {
	raw json.RawMessage
	err error
}

func (f *fakeDaemon) Call(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	return f.raw, f.err
}

// readyPipeline returns a fakePipeline configured for a healthy backend.
func readyPipeline() *fakePipeline {
	return &fakePipeline{
		available: true,
		state:     BackendState{Installed: true, DaemonRunning: true, ModelsReady: true},
	}
}

func TestQueryRankedHits(t *testing.T) {
	daemon := &fakeDaemon{
		raw: json.RawMessage(`[
			{"path":".rpi/designs/auth.md","title":"auth","score":0.91,"snippet":"auth flow"},
			{"path":".rpi/designs/cache.md","title":"cache","score":0.62,"snippet":"caching"}
		]`),
	}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "auth"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if resp.Status != StatusOK {
		t.Fatalf("expected StatusOK, got %s (%+v)", resp.Status, resp)
	}
	if len(resp.Hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(resp.Hits))
	}
	if resp.Hits[0].Type != "design" {
		t.Errorf("expected type=design from path inference, got %q", resp.Hits[0].Type)
	}
	if resp.Hits[0].Score != 0.91 {
		t.Errorf("expected top score 0.91, got %v", resp.Hits[0].Score)
	}
}

func TestQueryEmptyResult(t *testing.T) {
	daemon := &fakeDaemon{raw: json.RawMessage("[]")}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "nothing"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if resp.Status != StatusEmpty {
		t.Fatalf("expected StatusEmpty, got %s", resp.Status)
	}
	if len(resp.Hits) != 0 {
		t.Errorf("expected zero hits, got %d", len(resp.Hits))
	}
}

func TestQueryBackendUnavailable(t *testing.T) {
	pipeline := &fakePipeline{available: false}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: pipeline, Daemon: &fakeDaemon{}})

	if resp.Status != StatusBackendUnavailable {
		t.Fatalf("expected StatusBackendUnavailable, got %s", resp.Status)
	}
	if resp.InstallHint == "" {
		t.Error("expected install_hint to be populated")
	}
	if resp.Fallback == "" {
		t.Error("expected fallback hint to be populated")
	}
}

func TestQueryDaemonNotRunning(t *testing.T) {
	pipeline := readyPipeline()
	pipeline.state.DaemonRunning = false
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: pipeline, Daemon: &fakeDaemon{}})

	if resp.Status != StatusBackendError {
		t.Fatalf("expected StatusBackendError, got %s", resp.Status)
	}
	if resp.Error == nil || resp.Error.Stage != StageDaemonNotRunning {
		t.Fatalf("expected StageDaemonNotRunning, got %+v", resp.Error)
	}
	if resp.Error.Hint == "" {
		t.Error("expected hint pointing at warmup")
	}
}

func TestQueryModelsNotReady(t *testing.T) {
	pipeline := readyPipeline()
	pipeline.state.ModelsReady = false
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: pipeline, Daemon: &fakeDaemon{}})

	if resp.Status != StatusBackendError {
		t.Fatalf("expected StatusBackendError, got %s", resp.Status)
	}
	if resp.Error == nil || resp.Error.Stage != StageModelsNotReady {
		t.Fatalf("expected StageModelsNotReady, got %+v", resp.Error)
	}
}

func TestQueryUpdateFailure(t *testing.T) {
	pipeline := readyPipeline()
	pipeline.refreshErr = &RefreshError{Stage: StageUpdate, Message: "boom"}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: pipeline, Daemon: &fakeDaemon{}})

	if resp.Status != StatusBackendError {
		t.Fatalf("expected StatusBackendError, got %s", resp.Status)
	}
	if resp.Error == nil || resp.Error.Stage != StageUpdate {
		t.Fatalf("expected StageUpdate, got %+v", resp.Error)
	}
}

func TestQueryEmbedFailure(t *testing.T) {
	pipeline := readyPipeline()
	pipeline.refreshErr = &RefreshError{Stage: StageEmbed, Message: "model load failed"}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: pipeline, Daemon: &fakeDaemon{}})

	if resp.Error == nil || resp.Error.Stage != StageEmbed {
		t.Fatalf("expected StageEmbed, got %+v", resp.Error)
	}
}

func TestQueryDaemonCallParseFailure(t *testing.T) {
	daemon := &fakeDaemon{err: errParse}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if resp.Error == nil || resp.Error.Stage != StageParse {
		t.Fatalf("expected StageParse, got %+v", resp.Error)
	}
}

func TestQueryDaemonCallConnectionRefused(t *testing.T) {
	daemon := &fakeDaemon{err: errDaemonNotRunning}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if resp.Error == nil || resp.Error.Stage != StageDaemonNotRunning {
		t.Fatalf("expected StageDaemonNotRunning, got %+v", resp.Error)
	}
}

func TestQueryDaemonGenericError(t *testing.T) {
	daemon := &fakeDaemon{err: errors.New("some other failure")}
	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if resp.Error == nil || resp.Error.Stage != StageQuery {
		t.Fatalf("expected StageQuery, got %+v", resp.Error)
	}
}

func TestQueryTypeFilter(t *testing.T) {
	daemon := &fakeDaemon{
		raw: json.RawMessage(`[
			{"path":".rpi/designs/a.md","title":"a","score":0.9,"snippet":"x"},
			{"path":".rpi/plans/b.md","title":"b","score":0.85,"snippet":"y"},
			{"path":".rpi/specs/c.md","title":"c","score":0.8,"snippet":"z"}
		]`),
	}
	resp := Query(context.Background(), "/repo/.rpi",
		SearchParams{Query: "test", Type: "design"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if resp.Status != StatusOK {
		t.Fatalf("expected StatusOK, got %s", resp.Status)
	}
	if len(resp.Hits) != 1 {
		t.Fatalf("expected 1 hit after type filter, got %d", len(resp.Hits))
	}
	if resp.Hits[0].Type != "design" {
		t.Errorf("expected design hit, got %q", resp.Hits[0].Type)
	}
}

func TestQueryLimitHonored(t *testing.T) {
	daemon := &fakeDaemon{
		raw: json.RawMessage(`[
			{"path":".rpi/designs/a.md","title":"a","score":0.9,"snippet":"x"},
			{"path":".rpi/designs/b.md","title":"b","score":0.85,"snippet":"y"},
			{"path":".rpi/designs/c.md","title":"c","score":0.8,"snippet":"z"}
		]`),
	}
	resp := Query(context.Background(), "/repo/.rpi",
		SearchParams{Query: "test", Limit: 2},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if len(resp.Hits) != 2 {
		t.Fatalf("expected limit=2 to cap at 2 hits, got %d", len(resp.Hits))
	}
}

func TestQueryArchiveIncludedByDefault(t *testing.T) {
	daemon := &fakeDaemon{
		raw: json.RawMessage(`[
			{"path":".rpi/designs/active.md","title":"a","score":0.9,"snippet":"x"},
			{"path":".rpi/archive/designs/old.md","title":"o","score":0.85,"snippet":"y"}
		]`),
	}
	resp := Query(context.Background(), "/repo/.rpi",
		SearchParams{Query: "test"},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if len(resp.Hits) != 2 {
		t.Fatalf("expected both hits when archive included by default, got %d", len(resp.Hits))
	}
}

func TestQueryArchiveExcludedOnRequest(t *testing.T) {
	daemon := &fakeDaemon{
		raw: json.RawMessage(`[
			{"path":".rpi/designs/active.md","title":"a","score":0.9,"snippet":"x"},
			{"path":".rpi/archive/designs/old.md","title":"o","score":0.85,"snippet":"y"}
		]`),
	}
	resp := Query(context.Background(), "/repo/.rpi",
		SearchParams{Query: "test", ExcludeArchive: true},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if len(resp.Hits) != 1 {
		t.Fatalf("expected 1 hit (archive excluded), got %d", len(resp.Hits))
	}
	for _, h := range resp.Hits {
		if isArchivePath(h.Path) {
			t.Errorf("unexpected archive hit when ExcludeArchive=true: %s", h.Path)
		}
	}
}

func TestQueryMinScoreFilter(t *testing.T) {
	daemon := &fakeDaemon{
		raw: json.RawMessage(`[
			{"path":".rpi/designs/strong.md","title":"s","score":0.85,"snippet":"x"},
			{"path":".rpi/designs/weak.md","title":"w","score":0.25,"snippet":"y"}
		]`),
	}
	resp := Query(context.Background(), "/repo/.rpi",
		SearchParams{Query: "test", MinScore: 0.5},
		QueryOptions{Pipeline: readyPipeline(), Daemon: daemon})

	if len(resp.Hits) != 1 {
		t.Fatalf("expected sub-threshold hit dropped, got %d hits", len(resp.Hits))
	}
	if resp.Hits[0].Score < 0.5 {
		t.Errorf("kept hit below min_score: %v", resp.Hits[0].Score)
	}
}

func TestQueryWarningsCarryFromRefresh(t *testing.T) {
	pipeline := readyPipeline()
	pipeline.refresh = RefreshResult{Warnings: []string{"partial update: 1 file failed"}}
	daemon := &fakeDaemon{raw: json.RawMessage(`[{"path":".rpi/designs/x.md","title":"x","score":0.9,"snippet":"y"}]`)}

	resp := Query(context.Background(), "/repo/.rpi", SearchParams{Query: "x"},
		QueryOptions{Pipeline: pipeline, Daemon: daemon})

	if resp.Status != StatusOK {
		t.Fatalf("expected StatusOK, got %s", resp.Status)
	}
	if len(resp.Warnings) != 1 {
		t.Fatalf("expected warnings carried from refresh, got %v", resp.Warnings)
	}
}

func TestInferType(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		// Filesystem-relative .rpi/ paths
		{".rpi/designs/x.md", "design"},
		{".rpi/research/x.md", "research"},
		{".rpi/plans/x.md", "plan"},
		{".rpi/specs/x.md", "spec"},
		{".rpi/diagnoses/x.md", "diagnosis"},
		{".rpi/reviews/x.md", "review"},
		{".rpi/archive/designs/old.md", "design"},
		{"/abs/path/.rpi/designs/x.md", "design"},
		// qmd's collection-prefixed paths (active artifacts use plural folders)
		{"rpi-myrepo-abc123/specs/session-awareness.md", "spec"},
		{"rpi-myrepo-abc123/designs/2026-04-07-foo.md", "design"},
		{"rpi-myrepo-abc123/plans/2026-04-30-bar.md", "plan"},
		// qmd's archive structure uses singular type folder under YYYY-MM
		{"rpi-myrepo-abc123/archive/2026-04/plan/2026-04-07-x.md", "plan"},
		{"rpi-myrepo-abc123/archive/2026-04/design/2026-04-07-y.md", "design"},
		// Non-matches
		{".rpi/unknown/x.md", ""},
		{"unrelated/file.md", ""},
	}
	for _, c := range cases {
		if got := inferType(c.path); got != c.want {
			t.Errorf("inferType(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestIsArchivePath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{".rpi/archive/designs/old.md", true},
		{"rpi-myrepo-abc/archive/2026-04/plan/x.md", true},
		{".rpi/designs/active.md", false},
		{"rpi-myrepo-abc/specs/x.md", false},
		// "archive" only counts as a path segment, not a substring
		{"my-archive-research/notes.md", false},
	}
	for _, c := range cases {
		if got := isArchivePath(c.path); got != c.want {
			t.Errorf("isArchivePath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestParseQmdHitsWrapperShape(t *testing.T) {
	// Defensive: qmd may wrap the array in {"results":[...]} or {"hits":[...]}.
	raw := json.RawMessage(`{"results":[{"path":".rpi/designs/x.md","title":"x","score":0.9}]}`)
	hits, err := parseQmdHits(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit from wrapper, got %d", len(hits))
	}
}
