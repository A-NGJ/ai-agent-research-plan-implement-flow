package frontmatter

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Document represents a markdown file with optional YAML frontmatter.
type Document struct {
	Frontmatter map[string]interface{}
	Body        string
	Path        string
}

// Parse reads a markdown file and splits it into frontmatter and body.
// Files without frontmatter return an empty map (not an error).
func Parse(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(data, path)
}

// ParseBytes parses frontmatter from raw bytes.
func ParseBytes(data []byte, path string) (*Document, error) {
	content := string(data)
	doc := &Document{
		Frontmatter: make(map[string]interface{}),
		Path:        path,
	}

	// Must start with "---\n"
	if !strings.HasPrefix(content, "---\n") {
		doc.Body = content
		return doc, nil
	}

	// Find closing "---"
	rest := content[4:] // skip opening "---\n"
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		doc.Body = content
		return doc, nil
	}

	yamlBlock := rest[:idx]
	// Body starts after closing "---\n"
	body := rest[idx+4:]
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}

	if err := yaml.Unmarshal([]byte(yamlBlock), &doc.Frontmatter); err != nil {
		return nil, fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}
	if doc.Frontmatter == nil {
		doc.Frontmatter = make(map[string]interface{})
	}

	doc.Body = body
	return doc, nil
}
