package workflow

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Target identifies which AI coding tool the workflow assets are installed for.
type Target string

const (
	TargetClaude   Target = "claude"
	TargetOpenCode Target = "opencode"
)

// modelMap translates short model aliases used in Claude Code assets to full
// provider-qualified IDs used by OpenCode.
var modelMap = map[string]string{
	"opus":   "anthropic/claude-opus-4-6",
	"sonnet": "anthropic/claude-sonnet-4-6",
	"haiku":  "anthropic/claude-haiku-4-5-20251001",
}

//go:embed all:assets
var assets embed.FS

// ReadAsset reads an embedded file from the assets directory.
// The path should be relative to assets/ (e.g., "templates/CLAUDE.md.template").
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile("assets/" + path)
}

// Install copies all embedded workflow files (agents, commands, skills)
// into the target .claude/ directory. Existing files are only overwritten
// when force is true.
func Install(claudeDir string, force bool) (int, error) {
	return InstallTo(claudeDir, TargetClaude, force)
}

// InstallTo copies all embedded workflow files into targetDir, applying
// frontmatter transforms for the given target. Existing files are only
// overwritten when force is true.
func InstallTo(targetDir string, target Target, force bool) (int, error) {
	count := 0
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("assets", path)
		dest := filepath.Join(targetDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		if _, err := os.Stat(dest); err == nil && !force {
			return nil
		}

		data, err := assets.ReadFile(path)
		if err != nil {
			return err
		}

		if target == TargetOpenCode {
			parts := strings.SplitN(filepath.ToSlash(rel), "/", 2)
			if len(parts) > 0 {
				switch parts[0] {
				case "commands":
					data = transformCommandFrontmatter(data)
					data = transformCommandBody(data)
				case "agents":
					transformed, transformErr := transformAgentFrontmatter(data)
					if transformErr != nil {
						return fmt.Errorf("transform agent %s: %w", path, transformErr)
					}
					data = transformed
				}
			}
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

// transformCommandFrontmatter rewrites the model: field in a command markdown
// file from a short alias ("opus") to the full provider-qualified ID for OpenCode.
func transformCommandFrontmatter(content []byte) []byte {
	lines := bytes.Split(content, []byte("\n"))
	inFrontmatter := false
	fmCount := 0
	for i, line := range lines {
		s := string(bytes.TrimRight(line, " \t"))
		if s == "---" {
			fmCount++
			if fmCount == 1 {
				inFrontmatter = true
				continue
			}
			break
		}
		if inFrontmatter && strings.HasPrefix(s, "model: ") {
			alias := strings.TrimPrefix(s, "model: ")
			if fullID, ok := modelMap[alias]; ok {
				lines[i] = []byte("model: " + fullID)
			}
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

// transformCommandBody rewrites Claude Code-specific tool invocation patterns
// in command markdown body content to OpenCode conventions.
func transformCommandBody(content []byte) []byte {
	// Replace specific pattern first (before the general one)
	result := bytes.ReplaceAll(content,
		[]byte("Sub-task (@codebase-analyzer):"),
		[]byte("Use @codebase-analyzer to"))
	result = bytes.ReplaceAll(result,
		[]byte("Sub-task:"),
		[]byte("Subtask:"))

	// Remove lines containing "using TodoWrite"
	lines := bytes.Split(result, []byte("\n"))
	filtered := make([][]byte, 0, len(lines))
	for _, line := range lines {
		if bytes.Contains(line, []byte("using TodoWrite")) {
			continue
		}
		filtered = append(filtered, line)
	}
	return bytes.Join(filtered, []byte("\n"))
}

// transformAgentFrontmatter converts a Claude Code agent markdown file's
// frontmatter to OpenCode format: adds mode: subagent, converts tools to
// a bool deny-map, drops model: inherit.
func transformAgentFrontmatter(content []byte) ([]byte, error) {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return content, nil
	}
	end := strings.Index(s[4:], "\n---\n")
	if end < 0 {
		// Check for --- at end of file (no trailing newline after closing ---)
		end = strings.Index(s[4:], "\n---")
		if end < 0 {
			return content, nil
		}
	}
	fmStr := s[4 : 4+end]
	// Find where the body starts: skip past the closing ---\n
	bodyStart := 4 + end + 4 // len("\n---")
	if bodyStart < len(s) && s[bodyStart] == '\n' {
		bodyStart++
	}
	body := ""
	if bodyStart < len(s) {
		body = s[bodyStart:]
	}

	var src struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(fmStr), &src); err != nil {
		return nil, fmt.Errorf("parse agent frontmatter: %w", err)
	}

	out := struct {
		Name        string          `yaml:"name"`
		Description string          `yaml:"description"`
		Mode        string          `yaml:"mode"`
		Tools       map[string]bool `yaml:"tools"`
	}{
		Name:        src.Name,
		Description: src.Description,
		Mode:        "subagent",
		Tools:       map[string]bool{"bash": false, "write": false, "edit": false},
	}

	fmBytes, err := yaml.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal agent frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n")
	buf.WriteString(body)
	return buf.Bytes(), nil
}
