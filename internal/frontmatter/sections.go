package frontmatter

import (
	"strings"
)

// ExtractSection returns the content of the first ## heading that matches
// the given name via case-insensitive prefix match. Returns the content
// between the matched heading and the next ## heading (or EOF), and true
// if found. The heading line itself is included in the output.
func ExtractSection(body string, heading string) (string, bool) {
	result := ExtractSections(body, []string{heading})
	for _, content := range result {
		return content, true
	}
	return "", false
}

// ExtractSections extracts multiple sections by name. Returns a map of
// actual heading text → section content. Only matched sections appear in
// the map. When multiple headings match a single prefix, all are included.
func ExtractSections(body string, headings []string) map[string]string {
	if body == "" || len(headings) == 0 {
		return nil
	}

	// Normalize requested headings to lowercase for matching.
	lowerHeadings := make([]string, len(headings))
	for i, h := range headings {
		lowerHeadings[i] = strings.ToLower(strings.TrimSpace(h))
	}

	result := make(map[string]string)

	// Split into sections. The first element is content before any ## heading.
	parts := strings.Split(body, "\n## ")

	for i, part := range parts {
		if i == 0 {
			// Check if body starts with "## " (no leading newline).
			if !strings.HasPrefix(body, "## ") {
				continue
			}
			// Body starts with ## — strip the prefix we'll re-add.
			part = strings.TrimPrefix(body, "## ")
		}

		// The heading text is the first line of this part.
		headingEnd := strings.Index(part, "\n")
		var headingText string
		if headingEnd == -1 {
			headingText = part
		} else {
			headingText = part[:headingEnd]
		}

		lowerActual := strings.ToLower(headingText)

		for _, lh := range lowerHeadings {
			if strings.HasPrefix(lowerActual, lh) {
				result["## "+headingText] = "## " + part
				break
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// ListSections returns the ## heading lines from the given markdown body,
// in document order. Each entry is the full heading line (e.g. "## Summary").
// Returns nil if the body contains no ## headings.
func ListSections(body string) []string {
	if body == "" {
		return nil
	}

	var headings []string

	parts := strings.Split(body, "\n## ")

	for i, part := range parts {
		if i == 0 {
			if !strings.HasPrefix(body, "## ") {
				continue
			}
			part = strings.TrimPrefix(body, "## ")
		}

		headingEnd := strings.Index(part, "\n")
		var headingText string
		if headingEnd == -1 {
			headingText = part
		} else {
			headingText = part[:headingEnd]
		}

		headings = append(headings, "## "+headingText)
	}

	return headings
}
