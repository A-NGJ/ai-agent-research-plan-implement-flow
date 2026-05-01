package workflow

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Target identifies which AI coding tool the workflow assets are installed for.
type Target string

const (
	TargetClaude     Target = "claude"
	TargetOpenCode   Target = "opencode"
	TargetAgentsOnly Target = "agents-only"
)

// modelMap translates short model aliases used in Claude Code assets to full
// provider-qualified IDs used by OpenCode.
var modelMap = map[string]string{
	"opus":   "anthropic/claude-opus-4-6",
	"sonnet": "anthropic/claude-sonnet-4-6",
	"haiku":  "anthropic/claude-haiku-4-5-20251001",
}

// readOnlyBuiltinTools lists Claude Code built-in tools available to read-only skills.
const readOnlyBuiltinTools = "Read,Glob,Grep,Bash,Agent,LSP"

// mcpTools lists all RPI MCP tools with the mcp__rpi__ prefix.
const mcpTools = "mcp__rpi__rpi_git_context,mcp__rpi__rpi_git_changed_files," +
	"mcp__rpi__rpi_git_sensitive_check,mcp__rpi__rpi_archive_scan," +
	"mcp__rpi__rpi_scan,mcp__rpi__rpi_scaffold," +
	"mcp__rpi__rpi_frontmatter_get,mcp__rpi__rpi_frontmatter_set," +
	"mcp__rpi__rpi_frontmatter_transition,mcp__rpi__rpi_chain," +
	"mcp__rpi__rpi_extract,mcp__rpi__rpi_extract_list_sections," +
	"mcp__rpi__rpi_verify_completeness,mcp__rpi__rpi_verify_markers," +
	"mcp__rpi__rpi_verify_spec,mcp__rpi__rpi_context_essentials," +
	"mcp__rpi__rpi_session_resume,mcp__rpi__rpi_suggest_next," +
	"mcp__rpi__rpi_archive_check_refs,mcp__rpi__rpi_archive_move"

// readOnlyTools is the full allowed-tools value for read-only skills.
const readOnlyTools = readOnlyBuiltinTools + "," + mcpTools

// researchTools adds web access to the read-only tool set.
const researchTools = readOnlyTools + ",WebSearch,WebFetch"

// skillOverrides defines per-skill tool-specific frontmatter fields to inject
// when the target is not agents-only. Skills not listed here get no extra
// fields.
var skillOverrides = map[string]map[string]string{
	"rpi-archive":  {"model": "haiku", "disable-model-invocation": "true"},
	"rpi-commit":   {"model": "haiku"},
	"rpi-research": {"allowed-tools": researchTools, "context": "fork"},
	"rpi-verify":   {"allowed-tools": readOnlyTools},
	"rpi-explain":  {"allowed-tools": readOnlyTools},
}

//go:embed all:assets
var assets embed.FS

// ReadAsset reads an embedded file from the assets directory.
// The path should be relative to assets/ (e.g., "templates/CLAUDE.md.template").
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile("assets/" + path)
}

// backupAndWrite writes newData to dest. If dest already exists with different
// content, a .bak copy is created first. Returns whether the file was written
// and whether a backup was created. Skips writing if content is identical.
func backupAndWrite(dest string, newData []byte) (written, backedUp bool, err error) {
	existing, readErr := os.ReadFile(dest)
	if readErr == nil {
		if bytes.Equal(existing, newData) {
			return false, false, nil
		}
		if writeErr := os.WriteFile(dest+".bak", existing, 0644); writeErr != nil {
			return false, false, fmt.Errorf("backup %s: %w", dest, writeErr)
		}
		backedUp = true
	}
	if writeErr := os.WriteFile(dest, newData, 0644); writeErr != nil {
		return false, backedUp, fmt.Errorf("write %s: %w", dest, writeErr)
	}
	return true, backedUp, nil
}

// InstallSkills copies embedded skills to skillsDir as Agent Skills-compliant
// SKILL.md files. For non-agents-only targets, tool-specific frontmatter
// (model, etc.) is injected into the installed SKILL.md copies. Any sibling
// files in a skill source dir (e.g. an upstream LICENSE for bundled
// third-party skills) are copied verbatim so attribution travels with each
// deployed copy.
// Existing files that differ are backed up to .bak before overwriting.
func InstallSkills(skillsDir string, target Target) (installed int, backedUp int, err error) {
	entries, readErr := fs.ReadDir(assets, "assets/skills")
	if readErr != nil {
		return 0, 0, fmt.Errorf("read embedded skills: %w", readErr)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		srcDir := "assets/skills/" + skillName

		files, readErr := fs.ReadDir(assets, srcDir)
		if readErr != nil {
			return installed, backedUp, fmt.Errorf("read %s: %w", srcDir, readErr)
		}

		destDir := filepath.Join(skillsDir, skillName)
		if mkErr := os.MkdirAll(destDir, 0755); mkErr != nil {
			return installed, backedUp, fmt.Errorf("create %s: %w", destDir, mkErr)
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			fileName := f.Name()
			srcPath := srcDir + "/" + fileName

			data, readErr := assets.ReadFile(srcPath)
			if readErr != nil {
				return installed, backedUp, fmt.Errorf("read %s: %w", srcPath, readErr)
			}

			if fileName == "SKILL.md" && target != TargetAgentsOnly {
				if overrides, ok := skillOverrides[skillName]; ok {
					data = injectFrontmatter(data, overrides, target)
				}
			}

			dest := filepath.Join(destDir, fileName)
			written, backed, writeErr := backupAndWrite(dest, data)
			if writeErr != nil {
				return installed, backedUp, writeErr
			}
			if written {
				installed++
			}
			if backed {
				backedUp++
			}
		}
	}

	return installed, backedUp, nil
}

// InstallTemplates copies embedded .tmpl files to templatesDir.
// Existing files that differ are backed up to .bak before overwriting.
func InstallTemplates(templatesDir string) (installed int, backedUp int, err error) {
	entries, readErr := fs.ReadDir(assets, "assets/templates")
	if readErr != nil {
		return 0, 0, fmt.Errorf("read embedded templates: %w", readErr)
	}

	if mkErr := os.MkdirAll(templatesDir, 0755); mkErr != nil {
		return 0, 0, fmt.Errorf("create %s: %w", templatesDir, mkErr)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		data, readErr := assets.ReadFile("assets/templates/" + entry.Name())
		if readErr != nil {
			return installed, backedUp, fmt.Errorf("read %s: %w", entry.Name(), readErr)
		}

		dest := filepath.Join(templatesDir, entry.Name())
		written, backed, writeErr := backupAndWrite(dest, data)
		if writeErr != nil {
			return installed, backedUp, writeErr
		}
		if written {
			installed++
		}
		if backed {
			backedUp++
		}
	}

	return installed, backedUp, nil
}

// InstallAgents copies embedded agent definitions to agentsDir.
// Existing files that differ are backed up to .bak before overwriting.
func InstallAgents(agentsDir string) (installed int, backedUp int, err error) {
	entries, readErr := fs.ReadDir(assets, "assets/agents")
	if readErr != nil {
		return 0, 0, fmt.Errorf("read embedded agents: %w", readErr)
	}

	if mkErr := os.MkdirAll(agentsDir, 0755); mkErr != nil {
		return 0, 0, fmt.Errorf("create %s: %w", agentsDir, mkErr)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, readErr := assets.ReadFile("assets/agents/" + entry.Name())
		if readErr != nil {
			return installed, backedUp, fmt.Errorf("read %s: %w", entry.Name(), readErr)
		}

		dest := filepath.Join(agentsDir, entry.Name())
		written, backed, writeErr := backupAndWrite(dest, data)
		if writeErr != nil {
			return installed, backedUp, writeErr
		}
		if written {
			installed++
		}
		if backed {
			backedUp++
		}
	}

	return installed, backedUp, nil
}

// injectFrontmatter inserts additional YAML fields into an existing SKILL.md
// frontmatter block. For OpenCode targets, model aliases are translated to
// full provider-qualified IDs.
func injectFrontmatter(content []byte, fields map[string]string, target Target) []byte {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return content
	}

	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return content
	}

	fmEnd := 4 + end // index of \n before closing ---
	existingFM := s[4:fmEnd]
	rest := s[fmEnd:] // "\n---\n..." rest of file

	var extra strings.Builder
	for key, val := range fields {
		if key == "model" && target == TargetOpenCode {
			if fullID, ok := modelMap[val]; ok {
				val = fullID
			}
		}
		extra.WriteString(key + ": " + val + "\n")
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(existingFM)
	buf.WriteString("\n")
	buf.WriteString(extra.String())
	buf.WriteString(rest[1:]) // skip the leading \n (already added above)
	return []byte(buf.String())
}
