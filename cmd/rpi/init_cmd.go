package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/templates"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	initForce         bool
	initNoClaudeMD    bool
	initTrackThoughts bool
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
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	w := cmd.OutOrStdout()

	claudeDir := filepath.Join(targetDir, ".claude")

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

	// Install embedded workflow files (agents, commands, skills)
	n, err := workflow.Install(claudeDir, initForce)
	if err != nil {
		return fmt.Errorf("install workflow files: %w", err)
	}
	logSuccess(w, fmt.Sprintf("Installed %d workflow files (agents, commands, skills)", n))

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
