package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/templates"
	"github.com/spf13/cobra"
)

var (
	initForce         bool
	initNoClaudeMD    bool
	initTrackThoughts bool
	initAll           bool
	initUpdate        bool
	initAgentsOnly    bool
	initCommandsOnly  bool
	initSkillsOnly    bool
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorReset  = "\033[0m"
)

func logSuccess(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s✓%s %s\n", colorGreen, colorReset, msg)
}

func logWarning(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s!%s %s\n", colorYellow, colorReset, msg)
}

func logError(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s✗%s %s\n", colorRed, colorReset, msg)
}

func logInfo(w io.Writer, msg string) {
	fmt.Fprintf(w, "  %s\n", msg)
}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize project with .claude/ and .thoughts/ directory structure",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing files and directories")
	initCmd.Flags().BoolVar(&initNoClaudeMD, "no-claude-md", false, "Skip CLAUDE.md generation")
	initCmd.Flags().BoolVar(&initTrackThoughts, "track-thoughts", false, "Do not add .thoughts/ to .gitignore")
	initCmd.Flags().BoolVar(&initAll, "all", false, "Copy agents, commands, skills, and hooks from dotfiles")
	initCmd.Flags().BoolVar(&initUpdate, "update", false, "Update/refresh configs from dotfiles (sync mode)")
	initCmd.Flags().BoolVar(&initAgentsOnly, "agents-only", false, "Only copy agents (use with --all or --update)")
	initCmd.Flags().BoolVar(&initCommandsOnly, "commands-only", false, "Only copy commands (use with --all or --update)")
	initCmd.Flags().BoolVar(&initSkillsOnly, "skills-only", false, "Only copy skills (use with --all or --update)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	w := cmd.OutOrStdout()

	claudeDir := filepath.Join(targetDir, ".claude")

	// Handle update mode
	if initUpdate {
		return runInitUpdate(w, targetDir, claudeDir)
	}

	// Check if already initialized
	if _, err := os.Stat(claudeDir); err == nil && !initForce {
		return fmt.Errorf(".claude/ already exists; use --force to reinitialize")
	}

	// Create .claude/ subdirs
	claudeSubdirs := []string{"agents", "commands", "skills", "hooks"}
	for _, d := range claudeSubdirs {
		path := filepath.Join(claudeDir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		logSuccess(w, fmt.Sprintf("Created .claude/%s/", d))
	}

	// Create .thoughts/ subdirs
	thoughtsDir := filepath.Join(targetDir, ".thoughts")
	thoughtsSubdirs := []string{
		"research", "designs", "structures", "tickets",
		"plans", "specs", "reviews", "archive", "prs",
	}
	for _, d := range thoughtsSubdirs {
		path := filepath.Join(thoughtsDir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		logSuccess(w, fmt.Sprintf("Created .thoughts/%s/", d))
	}

	// Generate CLAUDE.md
	if !initNoClaudeMD {
		claudeMDPath := filepath.Join(targetDir, "CLAUDE.md")
		if _, err := os.Stat(claudeMDPath); err == nil && !initForce {
			logWarning(w, "CLAUDE.md already exists (use --force to overwrite)")
		} else {
			content, err := templates.Get("CLAUDE.md")
			if err != nil {
				return fmt.Errorf("get CLAUDE.md template: %w", err)
			}
			if err := os.WriteFile(claudeMDPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("write CLAUDE.md: %w", err)
			}
			logSuccess(w, "Created CLAUDE.md")
		}
	}

	// Generate .thoughts/PIPELINE.md
	pipelinePath := filepath.Join(thoughtsDir, "PIPELINE.md")
	if _, err := os.Stat(pipelinePath); err == nil && !initForce {
		logWarning(w, ".thoughts/PIPELINE.md already exists (use --force to overwrite)")
	} else {
		content, err := templates.Get("PIPELINE.md")
		if err != nil {
			return fmt.Errorf("get PIPELINE.md template: %w", err)
		}
		if err := os.WriteFile(pipelinePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write PIPELINE.md: %w", err)
		}
		logSuccess(w, "Created .thoughts/PIPELINE.md")
	}

	// Manage .gitignore
	if err := ensureGitignoreEntry(w, targetDir, ".claude/settings.local.json"); err != nil {
		logWarning(w, fmt.Sprintf("Failed to update .gitignore: %v", err))
	}
	if !initTrackThoughts {
		if err := ensureGitignoreEntry(w, targetDir, ".thoughts/"); err != nil {
			logWarning(w, fmt.Sprintf("Failed to update .gitignore: %v", err))
		}
	}

	// Copy configs from dotfiles (only with --all)
	if initAll {
		source := dotfilesSource()
		if _, err := os.Stat(source); os.IsNotExist(err) {
			return fmt.Errorf("dotfiles source not found: %s", source)
		}
		for _, c := range filterComponents(source, claudeDir) {
			n, err := copyDirectory(c.src, c.dest)
			if err != nil {
				logWarning(w, fmt.Sprintf("Failed to copy %s: %v", c.name, err))
				continue
			}
			logSuccess(w, fmt.Sprintf("Copied %d %s", n, c.name))
		}
	}

	return nil
}

func ensureGitignoreEntry(w io.Writer, targetDir, entry string) error {
	gitignorePath := filepath.Join(targetDir, ".gitignore")

	// Check if entry already exists
	if data, err := os.ReadFile(gitignorePath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if line == entry {
				return nil // already present
			}
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n# RPI workflow\n%s\n", entry); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	logSuccess(w, fmt.Sprintf("Added %s to .gitignore", entry))
	return nil
}

func dotfilesSource() string {
	if v := os.Getenv("DOTFILES_CLAUDE"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

type component struct {
	name string
	src  string
	dest string
}

func filterComponents(source, claudeDir string) []component {
	all := []component{
		{"agents", filepath.Join(source, "agents"), filepath.Join(claudeDir, "agents")},
		{"commands", filepath.Join(source, "commands"), filepath.Join(claudeDir, "commands")},
		{"skills", filepath.Join(source, "skills"), filepath.Join(claudeDir, "skills")},
		{"hooks", filepath.Join(source, "hooks"), filepath.Join(claudeDir, "hooks")},
	}

	// If any --*-only flag is set, filter to just that component
	if initAgentsOnly || initCommandsOnly || initSkillsOnly {
		var filtered []component
		for _, c := range all {
			switch c.name {
			case "agents":
				if initAgentsOnly {
					filtered = append(filtered, c)
				}
			case "commands":
				if initCommandsOnly {
					filtered = append(filtered, c)
				}
			case "skills":
				if initSkillsOnly {
					filtered = append(filtered, c)
				}
			}
		}
		return filtered
	}

	return all
}

// copyDirectory copies all files and subdirectories from src to dest.
// Returns the number of top-level items copied.
func copyDirectory(src, dest string) (int, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, destPath); err != nil {
				return count, err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return count, err
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return count, err
			}
		}
		count++
	}
	return count, nil
}

func runInitUpdate(w io.Writer, targetDir, claudeDir string) error {
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf(".claude/ doesn't exist; run 'rpi init' first")
	}

	source := dotfilesSource()
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("dotfiles source not found: %s", source)
	}

	for _, c := range filterComponents(source, claudeDir) {
		stats, err := copyWithUpdate(w, c.src, c.dest, initForce)
		if err != nil {
			logWarning(w, fmt.Sprintf("Failed to update %s: %v", c.name, err))
			continue
		}
		logInfo(w, fmt.Sprintf("%s: %d new, %d updated, %d unchanged", c.name, stats.copied, stats.updated, stats.skipped))
	}

	if !initNoClaudeMD {
		if err := updateClaudeMD(w, targetDir); err != nil {
			logWarning(w, fmt.Sprintf("Failed to update CLAUDE.md: %v", err))
		}
	}

	return nil
}

type updateStats struct {
	copied  int
	updated int
	skipped int
}

// copyWithUpdate syncs files from src to dest with diff-aware logic.
// New files are copied. Unchanged files are skipped. Differing files
// are only overwritten when force is true; otherwise a warning is logged.
func copyWithUpdate(w io.Writer, src, dest string, force bool) (updateStats, error) {
	var stats updateStats
	entries, err := os.ReadDir(src)
	if err != nil {
		return stats, err
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return stats, err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			sub, err := copyWithUpdate(w, srcPath, destPath, force)
			if err != nil {
				return stats, err
			}
			stats.copied += sub.copied
			stats.updated += sub.updated
			stats.skipped += sub.skipped
			continue
		}
		srcData, err := os.ReadFile(srcPath)
		if err != nil {
			return stats, err
		}
		destData, destErr := os.ReadFile(destPath)
		if destErr != nil {
			// New file
			if err := os.WriteFile(destPath, srcData, 0644); err != nil {
				return stats, err
			}
			stats.copied++
		} else if bytes.Equal(srcData, destData) {
			stats.skipped++
		} else if force {
			if err := os.WriteFile(destPath, srcData, 0644); err != nil {
				return stats, err
			}
			stats.updated++
		} else {
			logWarning(w, fmt.Sprintf("Skipped (differs): %s (use --force to overwrite)", entry.Name()))
			stats.skipped++
		}
	}
	return stats, nil
}

type headingSection struct {
	heading string
	section string // full text from ## heading to next ## heading
}

func parseHeadingSections(content string) []headingSection {
	lines := strings.Split(content, "\n")
	var sections []headingSection
	var current *headingSection
	var buf strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if current != nil {
				current.section = buf.String()
				sections = append(sections, *current)
				buf.Reset()
			}
			current = &headingSection{heading: line}
		}
		if current != nil {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}
	if current != nil {
		current.section = buf.String()
		sections = append(sections, *current)
	}
	return sections
}

func parseHeadingSet(content string) map[string]bool {
	set := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			set[line] = true
		}
	}
	return set
}

// updateClaudeMD finds ## sections in the template that are missing from
// the existing CLAUDE.md and appends them at the end.
func updateClaudeMD(w io.Writer, targetDir string) error {
	tmplContent, err := templates.Get("CLAUDE.md")
	if err != nil {
		return err
	}
	outputPath := filepath.Join(targetDir, "CLAUDE.md")
	existing, err := os.ReadFile(outputPath)
	if err != nil {
		// No existing file — write template directly
		if err := os.WriteFile(outputPath, []byte(tmplContent), 0644); err != nil {
			return err
		}
		logSuccess(w, "Created CLAUDE.md from template")
		return nil
	}

	tmplSections := parseHeadingSections(tmplContent)
	existingHeadings := parseHeadingSet(string(existing))

	added := 0
	var appendBuf strings.Builder
	for _, s := range tmplSections {
		if !existingHeadings[s.heading] {
			appendBuf.WriteString("\n")
			appendBuf.WriteString(s.section)
			added++
			logSuccess(w, fmt.Sprintf("CLAUDE.md: added section \"%s\"", strings.TrimPrefix(s.heading, "## ")))
		}
	}

	if added == 0 {
		logInfo(w, "CLAUDE.md: up to date")
		return nil
	}

	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(appendBuf.String())
	return err
}

func copyDirRecursive(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, destPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
