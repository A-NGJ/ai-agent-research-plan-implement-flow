// Package split computes the rpi-plan-split complexity score for a design.
//
// The score is a deterministic three-signal heuristic used by the rpi-plan
// skill to decide whether to propose splitting a design into multiple
// sibling plans. See .rpi/designs/2026-05-07-rpi-plan-split.md and the
// rpi-plan skill for the full rule.
package split

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// DefaultThreshold is the score at and above which a split should be proposed.
const DefaultThreshold = 4

// Result is the full output of a complexity-score evaluation.
type Result struct {
	Path        string     `json:"path"`
	Components  Components `json:"components"`
	Dirs        Dirs       `json:"dirs"`
	Specs       Specs      `json:"specs"`
	Score       Score      `json:"score"`
	Threshold   int        `json:"threshold"`
	ShouldSplit bool       `json:"should_split"`
}

// Components describes the design's `## Components` (or `## File Structure`) breakdown.
type Components struct {
	Count    int      `json:"count"`
	Headings []string `json:"headings"`
	// Source is "components" if counts came from `## Components`,
	// "file_structure" if Components was absent and we fell back, or "" if neither.
	Source string `json:"source"`
}

// Dirs describes the distinct top-level directories referenced in `## File Structure`.
type Dirs struct {
	Count    int      `json:"count"`
	TopLevel []string `json:"top_level"`
}

// Specs describes the spec files referenced anywhere in the design.
type Specs struct {
	Count int      `json:"count"`
	Paths []string `json:"paths"`
}

// Score breaks down each signal's contribution to the total.
type Score struct {
	ComponentsContrib int `json:"components_contrib"`
	DirsContrib       int `json:"dirs_contrib"`
	MultiSpecContrib  int `json:"multi_spec_contrib"`
	Total             int `json:"total"`
}

var (
	numberedHeadingRe = regexp.MustCompile(`(?m)^### \d+\.\s+(.+)$`)
	sectionHeaderRe   = regexp.MustCompile(`(?m)^## (.+)$`)
	pathLikeRe        = regexp.MustCompile(`(\.?[a-zA-Z_][\w.-]*)(/[\w./-]+)+`)
	specRefRe         = regexp.MustCompile(`\.rpi/specs/[a-z0-9][a-z0-9_-]*\.md`)
	rootFileRe        = regexp.MustCompile(`\b(CLAUDE\.md|README\.md|README|\.gitignore|Makefile|go\.mod|go\.sum)\b`)
)

// Compute reads designPath and returns the heuristic result using DefaultThreshold.
func Compute(designPath string) (*Result, error) {
	data, err := os.ReadFile(designPath)
	if err != nil {
		return nil, fmt.Errorf("read design: %w", err)
	}
	return ComputeBytes(designPath, data, DefaultThreshold)
}

// ComputeBytes scores an in-memory design body.
// path is recorded on the result for reporting; threshold lets callers override DefaultThreshold.
func ComputeBytes(path string, body []byte, threshold int) (*Result, error) {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	text := string(body)

	componentsCount, headings, source := extractComponents(text)
	dirsCount, topLevel := extractDirs(text)
	specPaths := extractSpecs(text)

	componentsContrib := positiveDelta(componentsCount, 1)
	dirsContrib := positiveDelta(dirsCount, 1)
	multiSpecContrib := 0
	if len(specPaths) > 1 {
		multiSpecContrib = 2
	}
	total := componentsContrib + dirsContrib + multiSpecContrib

	return &Result{
		Path: path,
		Components: Components{
			Count:    componentsCount,
			Headings: headings,
			Source:   source,
		},
		Dirs: Dirs{
			Count:    dirsCount,
			TopLevel: topLevel,
		},
		Specs: Specs{
			Count: len(specPaths),
			Paths: specPaths,
		},
		Score: Score{
			ComponentsContrib: componentsContrib,
			DirsContrib:       dirsContrib,
			MultiSpecContrib:  multiSpecContrib,
			Total:             total,
		},
		Threshold:   threshold,
		ShouldSplit: total >= threshold,
	}, nil
}

func positiveDelta(count, base int) int {
	if count <= base {
		return 0
	}
	return count - base
}

// extractComponents counts `### N. ...` numbered subheadings inside `## Components`,
// falling back to `## File Structure` when Components is absent or has none.
func extractComponents(text string) (count int, headings []string, source string) {
	if h := numberedHeadingsInSection(text, "Components"); len(h) > 0 {
		return len(h), h, "components"
	}
	if h := numberedHeadingsInSection(text, "File Structure"); len(h) > 0 {
		return len(h), h, "file_structure"
	}
	return 0, nil, ""
}

// numberedHeadingsInSection returns the numbered ### subheadings between the named
// `## <section>` heading and the next `## ` heading. Returns nil if section is absent.
func numberedHeadingsInSection(text, sectionName string) []string {
	body, ok := sliceSection(text, sectionName)
	if !ok {
		return nil
	}
	matches := numberedHeadingRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, strings.TrimSpace(m[1]))
	}
	return out
}

// sliceSection returns the body between `## <name>` and the next `## ` heading.
// Comparison on the section name is case-insensitive and matches by prefix.
func sliceSection(text, sectionName string) (string, bool) {
	indices := sectionHeaderRe.FindAllStringSubmatchIndex(text, -1)
	wanted := strings.ToLower(sectionName)
	for i, idx := range indices {
		// idx[2:4] is the capture group covering the heading text.
		title := strings.ToLower(strings.TrimSpace(text[idx[2]:idx[3]]))
		if !strings.HasPrefix(title, wanted) {
			continue
		}
		bodyStart := idx[1] // end of the heading line
		bodyEnd := len(text)
		if i+1 < len(indices) {
			bodyEnd = indices[i+1][0]
		}
		return text[bodyStart:bodyEnd], true
	}
	return "", false
}

// extractDirs returns distinct top-level directories referenced in `## File Structure`,
// plus +1 (and a "<root>" entry) if any root-level files (CLAUDE.md etc.) are listed.
func extractDirs(text string) (int, []string) {
	body, ok := sliceSection(text, "File Structure")
	if !ok {
		return 0, nil
	}

	tops := map[string]bool{}
	for _, m := range pathLikeRe.FindAllStringSubmatch(body, -1) {
		seg := m[1]
		// Skip path-like matches that look like sentence fragments (e.g., "cmd.go" → "cmd").
		// Require at least one slash after the first segment for it to count as a dir.
		// pathLikeRe already enforces (/[\w./-]+)+ after the first capture, so seg is
		// guaranteed to have a child path. Just normalize the trailing slash.
		tops[seg+"/"] = true
	}

	out := make([]string, 0, len(tops)+1)
	for k := range tops {
		out = append(out, k)
	}

	if rootFileRe.MatchString(body) {
		out = append(out, "<root>")
	}

	sort.Strings(out)
	return len(out), out
}

// extractSpecs returns distinct .rpi/specs/<slug>.md references found anywhere in the design.
func extractSpecs(text string) []string {
	seen := map[string]bool{}
	for _, m := range specRefRe.FindAllString(text, -1) {
		seen[m] = true
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
