package chain

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
)

const maxDepth = 10

// linkFields are frontmatter fields that contain paths to other artifacts.
var linkFields = []string{"research", "ticket", "related_research", "proposal"}

// listLinkFields are frontmatter fields that contain lists of paths.
var listLinkFields = []string{"depends_on"}

// thoughtsPathRe matches .thoughts/ paths in markdown content (relative or absolute).
var thoughtsPathRe = regexp.MustCompile(`[^\s\(\[` + "`" + `"']*\.thoughts/[^\s\)\]` + "`" + `"']+\.md`)

// Artifact represents metadata about a single .thoughts/ file in a chain.
type Artifact struct {
	Path     string   `json:"path"`
	Type     string   `json:"type"`
	Status   *string  `json:"status"`
	Title    *string  `json:"title"`
	TicketID *string  `json:"ticket_id,omitempty"`
	LinksTo  []string `json:"links_to"`
}

// Result is the output of a chain resolution.
type Result struct {
	Root      string     `json:"root"`
	Artifacts []Artifact `json:"artifacts"`
	Warnings  []string   `json:"warnings,omitempty"`
}

// Resolve follows frontmatter links from the given root file recursively.
func Resolve(rootPath string) (*Result, error) {
	result := &Result{
		Root:      rootPath,
		Artifacts: []Artifact{},
	}

	visited := make(map[string]bool)
	if err := resolve(rootPath, 0, visited, result); err != nil {
		return nil, err
	}
	return result, nil
}

func resolve(path string, depth int, visited map[string]bool, result *Result) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", path, err)
	}

	if visited[absPath] {
		return nil
	}
	visited[absPath] = true

	if depth > maxDepth {
		result.Warnings = append(result.Warnings, fmt.Sprintf("max depth (%d) reached at %s", maxDepth, path))
		return nil
	}

	doc, err := frontmatter.Parse(path)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("cannot read %s: %v", path, err))
		return nil
	}

	links := extractLinks(doc)
	artifact := buildArtifact(doc, path, links)
	result.Artifacts = append(result.Artifacts, artifact)

	for _, link := range links {
		if err := resolve(link, depth+1, visited, result); err != nil {
			return err
		}
	}

	return nil
}

func extractLinks(doc *frontmatter.Document) []string {
	var links []string
	seen := make(map[string]bool)

	addLink := func(path string) {
		path = strings.TrimSpace(path)
		if path != "" && !seen[path] {
			seen[path] = true
			links = append(links, path)
		}
	}

	// Extract from single-value link fields
	for _, field := range linkFields {
		if val, ok := doc.Frontmatter[field]; ok {
			if s, ok := val.(string); ok {
				addLink(s)
			}
		}
	}

	// Extract from list link fields (depends_on, etc.)
	for _, field := range listLinkFields {
		if val, ok := doc.Frontmatter[field]; ok {
			switch v := val.(type) {
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok {
						addLink(s)
					}
				}
			case string:
				addLink(v)
			}
		}
	}

	// Fallback: if no frontmatter links found, try Source Documents section
	if len(links) == 0 && len(doc.Frontmatter) == 0 {
		links = extractSourceDocumentLinks(doc.Body)
	}

	return links
}

func extractSourceDocumentLinks(body string) []string {
	var links []string
	seen := make(map[string]bool)

	// Look for ## Source Documents section
	idx := strings.Index(body, "## Source Documents")
	if idx == -1 {
		idx = strings.Index(body, "## References")
		if idx == -1 {
			return nil
		}
	}

	section := body[idx:]
	// Stop at next ## heading
	if nextH2 := strings.Index(section[3:], "\n## "); nextH2 != -1 {
		section = section[:nextH2+3]
	}

	matches := thoughtsPathRe.FindAllString(section, -1)
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			links = append(links, m)
		}
	}
	return links
}

func buildArtifact(doc *frontmatter.Document, path string, links []string) Artifact {
	a := Artifact{
		Path:    path,
		Type:    inferType(path),
		LinksTo: links,
	}
	if a.LinksTo == nil {
		a.LinksTo = []string{}
	}

	if s, ok := getStringField(doc.Frontmatter, "status"); ok {
		a.Status = &s
	}
	if t, ok := getStringField(doc.Frontmatter, "topic"); ok {
		a.Title = &t
	} else if t, ok := getStringField(doc.Frontmatter, "title"); ok {
		a.Title = &t
	}
	if id, ok := getStringField(doc.Frontmatter, "ticket_id"); ok {
		a.TicketID = &id
	} else if id, ok := getStringField(doc.Frontmatter, "ticket"); ok {
		a.TicketID = &id
	}

	return a
}

func inferType(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		switch p {
		case "plans":
			return "plan"
		case "research":
			return "research"
		case "proposals":
			return "proposal"
		case "prs":
			return "pr"
		case "reviews":
			return "review"
		case "archive":
			return "archive"
		}
	}
	return "unknown"
}

func getStringField(fm map[string]interface{}, key string) (string, bool) {
	val, ok := fm[key]
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok
}
