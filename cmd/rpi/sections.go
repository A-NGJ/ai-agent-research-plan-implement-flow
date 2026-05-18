package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/rpi/internal/templates"
)

// templateSection holds one top-level (`## `) section from the rendered rules
// template — the heading line and the full text (heading + body).
type templateSection struct {
	heading string // e.g. "## Project Overview" (no trailing newline)
	text    string // heading line + body, trailing blank lines trimmed, ends with "\n"
}

// reconcileRulesFileSections appends top-level `## Heading` sections from the
// rendered rules template that are absent in the existing rules file. Append-
// only — existing content (including drifted bodies and the contract block) is
// preserved byte-for-byte. No-op when the file does not exist or no sections
// are missing.
//
// The contract block is excluded from reconciliation; writeContractBlock owns
// that path. See `.rpi/specs/rpi-skill-contract.md`.
func reconcileRulesFileSections(w io.Writer, rulesPath, rulesName string) error {
	rendered, err := templates.Get(rulesName)
	if err != nil {
		return fmt.Errorf("get %s template: %w", rulesName, err)
	}

	existing, err := os.ReadFile(rulesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", rulesPath, err)
	}

	// Strip the contract block so we don't try to reconcile against headings
	// owned by writeContractBlock. parseTemplateSections then sees a template
	// whose section list reflects only "regular" rules-file sections.
	templateWithoutContract := stripContractBlock(rendered)
	sections := parseTemplateSections(templateWithoutContract)
	if len(sections) == 0 {
		return nil
	}

	present := headingLineSet(existing)
	var missing []templateSection
	for _, s := range sections {
		if !present[s.heading] {
			missing = append(missing, s)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	var buf bytes.Buffer
	trimmed := bytes.TrimRight(existing, "\n")
	buf.Write(trimmed)
	if len(trimmed) > 0 {
		buf.WriteString("\n\n")
	}
	for i, s := range missing {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(s.text)
	}

	if bytes.Equal(buf.Bytes(), existing) {
		return nil
	}

	mode := os.FileMode(0644)
	if info, statErr := os.Stat(rulesPath); statErr == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(rulesPath, buf.Bytes(), mode); err != nil {
		return fmt.Errorf("write %s: %w", rulesPath, err)
	}
	logSuccess(w, fmt.Sprintf("Appended %d missing template section(s) to %s", len(missing), filepath.Base(rulesPath)))
	return nil
}

// stripContractBlock removes the fenced RPI Skill Contract block (begin marker
// line through end marker line, inclusive) from s. Returns s unchanged when
// either marker is missing — the section parser tolerates dangling markers and
// the reconciler's heading-match logic ignores non-`## ` lines anyway.
func stripContractBlock(s string) string {
	beginIdx := strings.Index(s, contractBeginPrefix)
	if beginIdx < 0 {
		return s
	}
	// Walk back to start of the begin-marker's line.
	lineStart := beginIdx
	for lineStart > 0 && s[lineStart-1] != '\n' {
		lineStart--
	}
	endIdx := strings.Index(s[beginIdx:], contractEndMarker)
	if endIdx < 0 {
		return s
	}
	endIdx += beginIdx + len(contractEndMarker)
	// Consume the newline that terminates the end marker line, if present.
	if endIdx < len(s) && s[endIdx] == '\n' {
		endIdx++
	}
	return s[:lineStart] + s[endIdx:]
}

// parseTemplateSections splits the input into ordered top-level sections. A
// section starts at a line beginning with `## ` (two hashes + space) and
// extends through the line before the next `## ` heading (or EOF). Content
// preceding the first `## ` is treated as preamble and excluded.
//
// Each returned section's `text` includes the heading line and the body, with
// trailing blank lines stripped and a single terminating "\n". The `heading`
// field carries the heading line trimmed of trailing whitespace, suitable for
// exact-match comparison.
func parseTemplateSections(s string) []templateSection {
	lines := strings.Split(s, "\n")
	var starts []int
	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			starts = append(starts, i)
		}
	}
	if len(starts) == 0 {
		return nil
	}
	out := make([]templateSection, 0, len(starts))
	for i, start := range starts {
		end := len(lines)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		// Drop trailing blank lines from the section body so we control the
		// separator when appending.
		j := end
		for j > start+1 && strings.TrimSpace(lines[j-1]) == "" {
			j--
		}
		heading := strings.TrimRight(lines[start], " \t\r")
		// Write the trimmed heading so output never carries stray trailing
		// whitespace from the template source.
		bodyLines := make([]string, 0, j-start)
		bodyLines = append(bodyLines, heading)
		bodyLines = append(bodyLines, lines[start+1:j]...)
		body := strings.Join(bodyLines, "\n") + "\n"
		out = append(out, templateSection{
			heading: heading,
			text:    body,
		})
	}
	return out
}

// headingLineSet returns the set of trimmed `## ` heading lines present in
// the existing file. Detection is purely line-based (heading must appear at
// line start). Code fences are not parsed — a `## ` literal inside a fence is
// indistinguishable from a real heading by this check, but template headings
// don't appear inside fenced code in our embedded templates, so the absence of
// fence awareness is acceptable here.
func headingLineSet(content []byte) map[string]bool {
	out := make(map[string]bool)
	for _, raw := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(raw, "## ") {
			out[strings.TrimRight(raw, " \t\r")] = true
		}
	}
	return out
}
