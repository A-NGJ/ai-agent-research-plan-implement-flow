package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
)

// ArtifactInfo represents metadata about a scanned artifact.
type ArtifactInfo struct {
	Path     string  `json:"path"`
	Type     string  `json:"type"`
	Status   *string `json:"status"`
	Title    *string `json:"title"`
	Date     *string `json:"date,omitempty"`
	TicketID *string `json:"ticket_id,omitempty"`
}

// Filters controls which artifacts are returned.
type Filters struct {
	Status     string
	Type       string
	Design     string
	References string
	Archivable bool
}

// Scan walks the thoughts directory and returns artifacts matching filters.
func Scan(thoughtsDir string, filters Filters) ([]ArtifactInfo, error) {
	var results []ArtifactInfo

	err := filepath.Walk(thoughtsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		// Skip archive directory
		if info.IsDir() && info.Name() == "archive" {
			return filepath.SkipDir
		}

		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		doc, parseErr := frontmatter.Parse(path)
		if parseErr != nil {
			return nil // skip unparseable files
		}

		artifact := buildInfo(doc, path)

		if matches(doc, artifact, filters) {
			results = append(results, artifact)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if results == nil {
		results = []ArtifactInfo{}
	}
	return results, nil
}

func buildInfo(doc *frontmatter.Document, path string) ArtifactInfo {
	a := ArtifactInfo{
		Path: path,
		Type: InferType(path),
	}

	if s, ok := getStr(doc.Frontmatter, "status"); ok {
		a.Status = &s
	}
	if t, ok := getStr(doc.Frontmatter, "topic"); ok {
		a.Title = &t
	} else if t, ok := getStr(doc.Frontmatter, "title"); ok {
		a.Title = &t
	}
	if d, ok := getStr(doc.Frontmatter, "date"); ok {
		a.Date = &d
	}
	if id, ok := getStr(doc.Frontmatter, "ticket_id"); ok {
		a.TicketID = &id
	} else if id, ok := getStr(doc.Frontmatter, "ticket"); ok {
		a.TicketID = &id
	}

	return a
}

func matches(doc *frontmatter.Document, info ArtifactInfo, f Filters) bool {
	if f.Status != "" {
		if info.Status == nil || *info.Status != f.Status {
			return false
		}
	}

	if f.Type != "" {
		if info.Type != f.Type {
			return false
		}
	}

	if f.Design != "" {
		val, ok := getStr(doc.Frontmatter, "design")
		if !ok || val != f.Design {
			return false
		}
	}

	if f.Archivable {
		if info.Status == nil {
			return false
		}
		s := *info.Status
		if s != "complete" && s != "superseded" {
			return false
		}
	}

	if f.References != "" {
		if !fileReferences(doc, f.References) {
			return false
		}
	}

	return true
}

// fileReferences checks if a document references the given path in frontmatter fields or body.
func fileReferences(doc *frontmatter.Document, target string) bool {
	// Check all frontmatter string values
	for _, val := range doc.Frontmatter {
		switch v := val.(type) {
		case string:
			if v == target {
				return true
			}
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok && s == target {
					return true
				}
			}
		}
	}

	// Check body text
	return strings.Contains(doc.Body, target)
}

// ReferenceDetail describes where a reference was found.
type ReferenceDetail struct {
	ReferencingFile string `json:"referencing_file"`
	FieldOrLine     string `json:"field_or_line"`
}

// FindReferences returns detailed info about all files referencing targetPath.
func FindReferences(thoughtsDir, targetPath string) ([]ReferenceDetail, error) {
	var results []ReferenceDetail

	err := filepath.Walk(thoughtsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == "archive" {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		doc, parseErr := frontmatter.Parse(path)
		if parseErr != nil {
			return nil
		}

		// Check frontmatter fields
		for key, val := range doc.Frontmatter {
			switch v := val.(type) {
			case string:
				if v == targetPath {
					results = append(results, ReferenceDetail{
						ReferencingFile: path,
						FieldOrLine:     key + ": " + v,
					})
				}
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok && s == targetPath {
						results = append(results, ReferenceDetail{
							ReferencingFile: path,
							FieldOrLine:     key + ": " + s,
						})
					}
				}
			}
		}

		// Check body lines
		for _, line := range strings.Split(doc.Body, "\n") {
			if strings.Contains(line, targetPath) {
				results = append(results, ReferenceDetail{
					ReferencingFile: path,
					FieldOrLine:     strings.TrimSpace(line),
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []ReferenceDetail{}
	}
	return results, nil
}

// CountReferences returns how many artifacts in thoughtsDir reference targetPath.
func CountReferences(thoughtsDir, targetPath string) (int, error) {
	refs, err := FindReferences(thoughtsDir, targetPath)
	return len(refs), err
}

// InferType determines the artifact type from the file path.
func InferType(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		switch p {
		case "plans":
			return "plan"
		case "tickets":
			return "ticket"
		case "designs":
			return "design"
		case "research":
			return "research"
		case "structures":
			return "structure"
		case "proposals":
			return "proposal"
		case "prs":
			return "pr"
		case "reviews":
			return "review"
		}
	}
	return "unknown"
}

// FindByTicketID locates a ticket file by its ticket_id frontmatter field.
// Returns the file path if found, or empty string if not found.
func FindByTicketID(thoughtsDir, ticketID string) (string, error) {
	ticketsDir := filepath.Join(thoughtsDir, "tickets")
	if _, err := os.Stat(ticketsDir); os.IsNotExist(err) {
		return "", nil
	}
	results, err := Scan(ticketsDir, Filters{})
	if err != nil {
		return "", err
	}
	for _, r := range results {
		if r.TicketID != nil && *r.TicketID == ticketID {
			return r.Path, nil
		}
	}
	return "", nil
}

func getStr(fm map[string]interface{}, key string) (string, bool) {
	val, ok := fm[key]
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok
}
