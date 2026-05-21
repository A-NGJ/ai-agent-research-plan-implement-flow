// Package specdrift computes deterministic structural drift signals over
// `.rpi/specs/`. It powers the `rpi spec-drift scan` CLI and the
// `rpi_spec_drift_scan` MCP tool. All signals are pure functions of the
// filesystem and git history — no LLM, no qmd, no network.
package specdrift

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/A-NGJ/rpi/internal/frontmatter"
)

// Signal is one drift indicator fired for a spec.
type Signal struct {
	Name    string         `json:"name"`
	Details map[string]any `json:"details"`
}

// SpecRecord is the scan result for a single spec file.
type SpecRecord struct {
	Path    string   `json:"path"`
	Signals []Signal `json:"signals"`
}

// ScanOptions tunes signal thresholds and filesystem roots.
type ScanOptions struct {
	StaleDays    int
	RatioLow     float64
	RatioHigh    float64
	SpecsDir     string
	ArtifactsDir string // if empty, defaults to filepath.Dir(SpecsDir)
}

// DefaultOptions returns the thresholds documented in the design.
func DefaultOptions() ScanOptions {
	return ScanOptions{
		StaleDays: 30,
		RatioLow:  0.5,
		RatioHigh: 3.0,
		SpecsDir:  ".rpi/specs",
	}
}

func applyDefaults(opts ScanOptions) ScanOptions {
	// Numeric thresholds pass through as-is so callers can pick 0 (e.g.
	// --stale-days=0 to flag every spec). Callers that want sensible
	// defaults should start from DefaultOptions().
	if opts.SpecsDir == "" {
		opts.SpecsDir = ".rpi/specs"
	}
	if opts.ArtifactsDir == "" {
		opts.ArtifactsDir = filepath.Dir(opts.SpecsDir)
	}
	return opts
}

// Scan walks opts.SpecsDir for spec files, runs each detector, and returns
// records in path-sorted order so output is byte-deterministic across runs.
func Scan(opts ScanOptions) ([]SpecRecord, error) {
	opts = applyDefaults(opts)

	specPaths, err := listSpecs(opts.SpecsDir)
	if err != nil {
		return nil, err
	}

	specSet := make(map[string]bool, len(specPaths))
	for _, p := range specPaths {
		specSet[p] = true
	}
	incoming, scannedArtifacts, err := scanIncomingReferences(opts.ArtifactsDir, specSet)
	if err != nil {
		return nil, err
	}

	records := make([]SpecRecord, 0, len(specPaths))
	for _, specPath := range specPaths {
		doc, err := frontmatter.Parse(specPath)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", specPath, err)
		}
		rec := SpecRecord{Path: specPath, Signals: []Signal{}}
		if s := detectStaleLastUpdated(doc, opts); s != nil {
			rec.Signals = append(rec.Signals, *s)
		}
		if s := detectScenarioCountMismatch(doc, opts); s != nil {
			rec.Signals = append(rec.Signals, *s)
		}
		if s := detectBrokenReferences(doc); s != nil {
			rec.Signals = append(rec.Signals, *s)
		}
		if s := detectNamingMismatch(specPath, doc); s != nil {
			rec.Signals = append(rec.Signals, *s)
		}
		if s := detectOrphaned(specPath, doc, incoming, scannedArtifacts); s != nil {
			rec.Signals = append(rec.Signals, *s)
		}
		records = append(records, rec)
	}
	return records, nil
}

func listSpecs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read specs dir %s: %w", dir, err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	sort.Strings(out)
	return out, nil
}

// --- Signal detectors ---

func detectStaleLastUpdated(doc *frontmatter.Document, opts ScanOptions) *Signal {
	t, ok := parseLastUpdated(doc.Frontmatter["last_updated"])
	if !ok {
		return nil
	}
	daysSince := int(time.Since(t).Hours() / 24)
	if daysSince < opts.StaleDays {
		return nil
	}
	if !gitAvailable() {
		return &Signal{
			Name: "stale_last_updated",
			Details: map[string]any{
				"git":        "unavailable",
				"days_since": daysSince,
			},
		}
	}
	refs := extractReferencedPaths(doc.Body)
	changed := 0
	for _, ref := range refs {
		n, err := gitLogCountSince(ref, t)
		if err == nil && n > 0 {
			changed++
		}
	}
	if changed == 0 {
		return nil
	}
	return &Signal{
		Name: "stale_last_updated",
		Details: map[string]any{
			"days_since":          daysSince,
			"files_changed_since": changed,
		},
	}
}

func detectScenarioCountMismatch(doc *frontmatter.Document, opts ScanOptions) *Signal {
	scenarios := countScenarios(doc.Body)
	refs := extractReferencedPaths(doc.Body)
	references := 0
	for _, r := range refs {
		// Heuristic: impl files are non-markdown. Spec-to-spec links don't
		// count toward the scenarios/impl ratio.
		if strings.HasSuffix(r, ".md") {
			continue
		}
		if _, err := os.Stat(r); err == nil {
			references++
		}
	}
	if scenarios == 0 && references == 0 {
		return nil
	}
	if references == 0 {
		// scenarios > 0 with no impl references — clear divergence.
		return &Signal{
			Name: "scenario_count_mismatch",
			Details: map[string]any{
				"scenarios":  scenarios,
				"references": references,
			},
		}
	}
	ratio := float64(scenarios) / float64(references)
	if ratio >= opts.RatioLow && ratio <= opts.RatioHigh {
		return nil
	}
	return &Signal{
		Name: "scenario_count_mismatch",
		Details: map[string]any{
			"scenarios":  scenarios,
			"references": references,
			"ratio":      ratio,
		},
	}
}

func detectBrokenReferences(doc *frontmatter.Document) *Signal {
	refs := extractReferencedPaths(doc.Body)
	var missing []string
	seen := make(map[string]bool)
	for _, r := range refs {
		if seen[r] {
			continue
		}
		seen[r] = true
		if _, err := os.Stat(r); os.IsNotExist(err) {
			missing = append(missing, r)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return &Signal{
		Name: "broken_references",
		Details: map[string]any{
			"missing_paths": missing,
		},
	}
}

func detectNamingMismatch(specPath string, doc *frontmatter.Document) *Signal {
	feature, _ := doc.Frontmatter["feature"].(string)
	if feature == "" {
		return nil
	}
	expected := slugify(feature) + ".md"
	if expected == filepath.Base(specPath) {
		return nil
	}
	return &Signal{
		Name: "naming_mismatch",
		Details: map[string]any{
			"expected_filename": expected,
		},
	}
}

func detectOrphaned(specPath string, doc *frontmatter.Document, incoming map[string]int, scannedArtifacts int) *Signal {
	if v, ok := doc.Frontmatter["orphaned"].(bool); ok && !v {
		return nil
	}
	if incoming[specPath] > 0 {
		return nil
	}
	return &Signal{
		Name: "orphaned",
		Details: map[string]any{
			"scanned_artifacts": scannedArtifacts,
		},
	}
}

// parseLastUpdated normalizes the `last_updated` frontmatter value. The YAML
// loader yields a time.Time when the field is an ISO timestamp; older docs
// may carry a plain string.
func parseLastUpdated(v any) (time.Time, bool) {
	switch val := v.(type) {
	case time.Time:
		return val, !val.IsZero()
	case string:
		if val == "" {
			return time.Time{}, false
		}
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t, true
		}
		if t, err := time.Parse("2006-01-02", val); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// --- Helpers ---

var markdownLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// extractReferencedPaths returns markdown link targets that look like file
// paths. URLs, anchors, and mailto: links are skipped.
func extractReferencedPaths(body string) []string {
	var out []string
	seen := make(map[string]bool)
	for _, m := range markdownLinkRe.FindAllStringSubmatch(body, -1) {
		target := strings.TrimSpace(m[1])
		if i := strings.Index(target, " "); i >= 0 {
			// Strip title: [foo](bar.md "title")
			target = strings.TrimSpace(target[:i])
		}
		if target == "" {
			continue
		}
		if strings.HasPrefix(target, "http://") ||
			strings.HasPrefix(target, "https://") ||
			strings.HasPrefix(target, "mailto:") ||
			strings.HasPrefix(target, "#") {
			continue
		}
		if i := strings.Index(target, "#"); i > 0 {
			target = target[:i]
		}
		if seen[target] {
			continue
		}
		seen[target] = true
		out = append(out, target)
	}
	return out
}

func countScenarios(body string) int {
	section, ok := frontmatter.ExtractSection(body, "Scenarios")
	if !ok {
		return 0
	}
	count := 0
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			count++
		}
	}
	return count
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-' || r == '.' || r == '/':
			b.WriteRune('-')
		}
	}
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}

func gitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func gitLogCountSince(path string, since time.Time) (int, error) {
	cmd := exec.Command("git", "log", "--since="+since.Format(time.RFC3339), "--format=%H", "--", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return 0, nil
	}
	return len(strings.Split(s, "\n")), nil
}

var frontmatterFieldRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*:\s*(.+)$`)

// scanIncomingReferences walks artifactsDir for `.md` files and counts how
// many times each spec is referenced — either by a markdown link target whose
// basename matches the spec, or by a frontmatter field value with the same
// basename. Specs reference themselves are not counted. Anything under an
// `archive` directory is skipped.
func scanIncomingReferences(artifactsDir string, specSet map[string]bool) (map[string]int, int, error) {
	refs := make(map[string]int, len(specSet))
	for sp := range specSet {
		refs[sp] = 0
	}
	specBasenames := make(map[string]string, len(specSet))
	for sp := range specSet {
		specBasenames[filepath.Base(sp)] = sp
	}

	if _, err := os.Stat(artifactsDir); os.IsNotExist(err) {
		return refs, 0, nil
	}

	scanned := 0
	walkErr := filepath.WalkDir(artifactsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if filepath.Base(path) == "archive" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		scanned++
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)

		// 1) markdown link targets
		for _, m := range markdownLinkRe.FindAllStringSubmatch(content, -1) {
			target := strings.TrimSpace(m[1])
			if i := strings.Index(target, " "); i >= 0 {
				target = strings.TrimSpace(target[:i])
			}
			if i := strings.Index(target, "#"); i > 0 {
				target = target[:i]
			}
			base := filepath.Base(target)
			if sp, ok := specBasenames[base]; ok && sp != path {
				refs[sp]++
			}
		}

		// 2) frontmatter field values (key: value)
		doc, err := frontmatter.ParseBytes(data, path)
		if err == nil {
			for _, v := range doc.Frontmatter {
				walkFrontmatterValue(v, func(s string) {
					base := filepath.Base(strings.TrimSpace(s))
					if sp, ok := specBasenames[base]; ok && sp != path {
						refs[sp]++
					}
				})
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, 0, walkErr
	}
	return refs, scanned, nil
}

// walkFrontmatterValue invokes fn for every string leaf inside arbitrary
// frontmatter values (string, []any, map[string]any).
func walkFrontmatterValue(v any, fn func(string)) {
	switch val := v.(type) {
	case string:
		fn(val)
	case []any:
		for _, item := range val {
			walkFrontmatterValue(item, fn)
		}
	case map[string]any:
		for _, item := range val {
			walkFrontmatterValue(item, fn)
		}
	}
}
