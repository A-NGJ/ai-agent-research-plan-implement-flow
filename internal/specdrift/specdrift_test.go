package specdrift

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/A-NGJ/rpi/internal/frontmatter"
)

func TestScan_TableDriven(t *testing.T) {
	// Force the git-unavailable branch so detectStaleLastUpdated fires on the
	// stale fixture without needing real git history.
	t.Setenv("PATH", "")

	opts := DefaultOptions()
	opts.SpecsDir = "testdata"

	records, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	want := map[string][]string{
		"testdata/bad-ratio.md":       {"scenario_count_mismatch"},
		"testdata/broken-refs.md":     {"broken_references"},
		"testdata/clean.md":           {},
		"testdata/naming-mismatch.md": {"naming_mismatch"},
		"testdata/orphaned-optout.md": {},
		"testdata/orphaned.md":        {"orphaned"},
		"testdata/referencer.md":      {},
		"testdata/stale.md":           {"stale_last_updated"},
	}
	if len(records) != len(want) {
		t.Fatalf("got %d records, want %d", len(records), len(want))
	}
	for _, rec := range records {
		expected, ok := want[rec.Path]
		if !ok {
			t.Errorf("unexpected record for %s", rec.Path)
			continue
		}
		got := signalNames(rec.Signals)
		if !equalStringSlices(got, expected) {
			t.Errorf("%s signals: got %v, want %v", rec.Path, got, expected)
		}
	}
}

func TestScan_Deterministic(t *testing.T) {
	t.Setenv("PATH", "")
	opts := DefaultOptions()
	opts.SpecsDir = "testdata"

	first, err := Scan(opts)
	if err != nil {
		t.Fatalf("first Scan: %v", err)
	}
	second, err := Scan(opts)
	if err != nil {
		t.Fatalf("second Scan: %v", err)
	}

	a, err := json.MarshalIndent(first, "", "  ")
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	b, err := json.MarshalIndent(second, "", "  ")
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(a) != string(b) {
		t.Errorf("scans differ:\n--- first ---\n%s\n--- second ---\n%s", a, b)
	}
}

func TestScan_OrphanedOptOut(t *testing.T) {
	t.Setenv("PATH", "")
	opts := DefaultOptions()
	opts.SpecsDir = "testdata"

	records, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	for _, rec := range records {
		if rec.Path != "testdata/orphaned-optout.md" {
			continue
		}
		for _, sig := range rec.Signals {
			if sig.Name == "orphaned" {
				t.Errorf("orphaned-optout.md should suppress orphaned signal, got: %v", sig)
			}
		}
		return
	}
	t.Fatal("orphaned-optout.md not in scan results")
}

func TestDetectStaleLastUpdated_GitUnavailable(t *testing.T) {
	t.Setenv("PATH", "")
	doc := &frontmatter.Document{
		Frontmatter: map[string]any{
			"last_updated": "2025-01-01T10:00:00+02:00",
		},
		Body: "no references here",
	}
	sig := detectStaleLastUpdated(doc, DefaultOptions())
	if sig == nil {
		t.Fatal("expected signal under PATH='', got nil")
	}
	if sig.Name != "stale_last_updated" {
		t.Errorf("name: got %q, want stale_last_updated", sig.Name)
	}
	if got, _ := sig.Details["git"].(string); got != "unavailable" {
		t.Errorf("details.git: got %v, want \"unavailable\"", sig.Details["git"])
	}
	if _, ok := sig.Details["days_since"].(int); !ok {
		t.Errorf("details.days_since missing or wrong type: %v", sig.Details["days_since"])
	}
}

func TestDetectStaleLastUpdated_NotOldEnough(t *testing.T) {
	doc := &frontmatter.Document{
		Frontmatter: map[string]any{
			"last_updated": "2026-05-15T10:00:00+02:00",
		},
	}
	sig := detectStaleLastUpdated(doc, ScanOptions{StaleDays: 365})
	if sig != nil {
		t.Errorf("expected no signal for recent spec under high threshold, got %+v", sig)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"foo bar":      "foo-bar",
		"Foo_Bar":      "foo-bar",
		"a/b.c":        "a-b-c",
		"  spaced  ":   "spaced",
		"multi--dash":  "multi-dash",
		"already-slug": "already-slug",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtractReferencedPaths_SkipsURLs(t *testing.T) {
	body := "see [docs](https://example.com), [code](impl.go), [anchor](#section), [titled](file.go \"title\")"
	got := extractReferencedPaths(body)
	want := []string{"impl.go", "file.go"}
	if !equalStringSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// --- helpers ---

func signalNames(sigs []Signal) []string {
	out := make([]string, 0, len(sigs))
	for _, s := range sigs {
		out = append(out, s.Name)
	}
	sort.Strings(out)
	return out
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
