package frontmatter

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Serialize renders the document back to bytes (frontmatter + body).
func Serialize(doc *Document) ([]byte, error) {
	if len(doc.Frontmatter) == 0 {
		return []byte(doc.Body), nil
	}

	yamlBytes, err := yaml.Marshal(doc.Frontmatter)
	if err != nil {
		return nil, err
	}

	var out []byte
	out = append(out, "---\n"...)
	out = append(out, yamlBytes...)
	out = append(out, "---\n"...)
	out = append(out, doc.Body...)
	return out, nil
}

// Write serializes the document and writes it back to doc.Path.
func Write(doc *Document) error {
	data, err := Serialize(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(doc.Path, data, 0644)
}
