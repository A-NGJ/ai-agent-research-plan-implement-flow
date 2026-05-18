package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/A-NGJ/rpi/internal/templates"
	"github.com/A-NGJ/rpi/internal/workflow"
)

var rpiSubdirs = []string{
	"research", "designs", "diagnoses",
	"plans", "specs", "reviews", "archive",
}

type syncOptions struct {
	targetDir string
	cfg       targetConfig
	skipRules bool
	noTrack   bool
	global    bool
	w         io.Writer
}

// liteSyncProject installs only the per-project artifacts: .rpi/ subdirs,
// templates, the rules file, and .gitignore entries. Skills, agents, MCP,
// and settings.json are intentionally not handled here — those run in the
// caller (syncProject) for full project init or come from the user-level
// global install. Designed for reuse by `rpi bootstrap`.
func liteSyncProject(opts syncOptions) error {
	rpiDir := filepath.Join(opts.targetDir, ".rpi")

	// Ensure .rpi/ subdirs exist
	for _, d := range rpiSubdirs {
		path := filepath.Join(rpiDir, d)
		if _, statErr := os.Stat(path); statErr != nil {
			if mkErr := os.MkdirAll(path, 0755); mkErr != nil {
				return fmt.Errorf("create %s: %w", path, mkErr)
			}
			logSuccess(opts.w, fmt.Sprintf("Created .rpi/%s/", d))
		}
	}

	// Manage .gitignore for tool dir (skip for agents-only)
	if opts.cfg.toolDir != "" {
		if err := ensureGitignoreEntries(opts.w, opts.targetDir, opts.cfg.toolDir+"/"); err != nil {
			logWarning(opts.w, fmt.Sprintf("Failed to update .gitignore: %v", err))
		}
	}

	// Manage .gitignore for .rpi/. Default: ignore artifacts but keep specs
	// tracked. With noTrack: ignore the entire .rpi/ tree.
	rpiEntries := []string{".rpi/*", "!.rpi/specs/"}
	if opts.noTrack {
		rpiEntries = []string{".rpi/"}
	}
	if err := ensureGitignoreEntries(opts.w, opts.targetDir, rpiEntries...); err != nil {
		logWarning(opts.w, fmt.Sprintf("Failed to update .gitignore: %v", err))
	}

	// Install scaffold templates
	templatesDir := filepath.Join(rpiDir, "templates")
	tplCount, tplBackups, err := workflow.InstallTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("install templates: %w", err)
	}
	if tplCount > 0 {
		logSuccess(opts.w, fmt.Sprintf("Installed %d template files", tplCount))
	}
	if tplBackups > 0 {
		logSuccess(opts.w, fmt.Sprintf("Backed up %d modified template files", tplBackups))
	}

	// Rules file: when missing, write the rendered template (which already
	// contains the spliced contract block). When present, leave user content
	// outside the contract fences alone — reconcileRulesFileSections appends
	// any top-level template sections missing from the existing file, and
	// writeContractBlock refreshes the fenced region in place
	// (see .rpi/specs/rpi-skill-contract.md).
	if !opts.skipRules && opts.cfg.rulesFile != "" {
		rulesPath := filepath.Join(opts.targetDir, opts.cfg.rulesFile)
		if _, statErr := os.Stat(rulesPath); os.IsNotExist(statErr) {
			content, tplErr := templates.Get(opts.cfg.rulesFile)
			if tplErr != nil {
				logWarning(opts.w, fmt.Sprintf("get %s template: %v", opts.cfg.rulesFile, tplErr))
			} else if writeErr := os.WriteFile(rulesPath, []byte(content), 0644); writeErr != nil {
				logWarning(opts.w, fmt.Sprintf("write %s: %v", opts.cfg.rulesFile, writeErr))
			} else {
				logSuccess(opts.w, fmt.Sprintf("Installed %s", opts.cfg.rulesFile))
			}
		}

		// Append any top-level template sections missing from the existing
		// rules file. Append-only and idempotent — never edits or replaces
		// existing content. Runs before writeContractBlock so the contract
		// block ends up at EOF on a contract-less file.
		if err := reconcileRulesFileSections(opts.w, rulesPath, opts.cfg.rulesFile); err != nil {
			logWarning(opts.w, fmt.Sprintf("reconcile sections in %s: %v", opts.cfg.rulesFile, err))
		}

		// Refresh the RPI Skill Contract block in place. Idempotent — no
		// change when the block is current. Skipped for agents-only target
		// (rulesFile == "") and under --no-claude-md (skipRules).
		if err := writeContractBlock(opts.w, rulesPath); err != nil {
			logWarning(opts.w, fmt.Sprintf("refresh contract block in %s: %v", opts.cfg.rulesFile, err))
		}
	}

	return nil
}

func syncProject(opts syncOptions) error {
	if !opts.global {
		if err := liteSyncProject(opts); err != nil {
			return err
		}
	}

	// Ensure tool subdirs exist (skip for agents-only)
	if opts.cfg.toolDir != "" {
		toolDirPath := filepath.Join(opts.targetDir, opts.cfg.toolDir)
		for _, d := range opts.cfg.subdirs {
			path := filepath.Join(toolDirPath, d)
			if _, statErr := os.Stat(path); statErr != nil {
				if mkErr := os.MkdirAll(path, 0755); mkErr != nil {
					return fmt.Errorf("create %s: %w", path, mkErr)
				}
				logSuccess(opts.w, fmt.Sprintf("Created %s/%s/", opts.cfg.toolDir, d))
			}
		}
	}

	// Install skills
	var skillsDir string
	if opts.cfg.toolDir != "" {
		skillsDir = filepath.Join(opts.targetDir, opts.cfg.toolDir, "skills")
	} else {
		skillsDir = filepath.Join(opts.targetDir, ".agents", "skills")
	}
	skillCount, skillBackups, err := workflow.InstallSkills(skillsDir)
	if err != nil {
		return fmt.Errorf("install skills: %w", err)
	}
	if skillCount > 0 {
		logSuccess(opts.w, fmt.Sprintf("Installed %d skill files", skillCount))
	}
	if skillBackups > 0 {
		logSuccess(opts.w, fmt.Sprintf("Backed up %d modified skill files", skillBackups))
	}

	// Install agent definitions (Claude target only)
	if opts.cfg.target == workflow.TargetClaude {
		agentsDir := filepath.Join(opts.targetDir, opts.cfg.toolDir, "agents")
		agentCount, agentBackups, err := workflow.InstallAgents(agentsDir)
		if err != nil {
			return fmt.Errorf("install agents: %w", err)
		}
		if agentCount > 0 {
			logSuccess(opts.w, fmt.Sprintf("Installed %d agent files", agentCount))
		}
		if agentBackups > 0 {
			logSuccess(opts.w, fmt.Sprintf("Backed up %d modified agent files", agentBackups))
		}
	}

	// Configure settings.json (Claude only)
	if opts.cfg.target == workflow.TargetClaude {
		toolDirPath := filepath.Join(opts.targetDir, opts.cfg.toolDir)
		configureSettings(opts.w, toolDirPath)
		configureHooks(opts.w, toolDirPath)
	}

	return nil
}
