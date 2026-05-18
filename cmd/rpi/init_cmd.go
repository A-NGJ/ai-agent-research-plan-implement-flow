package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/rpi/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	initNoClaudeMD bool
	initNoTrack    bool
	initTarget     string
	initNoMCP      bool
	initGlobal     bool
)

const (
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

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize project with workflow skills and rules file",
	Long: `Initialize a project with the RPI workflow structure.

Targets:
  claude      Creates .claude/ with skills and hooks subdirectories
              and a CLAUDE.md rules file (default)
  opencode    Creates .opencode/ with the same structure and an AGENTS.md rules file
  agents-only Creates .agents/skills/ with cross-tool Agent Skills — no tool-specific
              directory, no MCP config

Also creates:
  .rpi/             Artifact directory hierarchy (research, designs, plans, etc.)

Use --no-claude-md to skip rules file generation. By default, .rpi/* is
added to .gitignore with a !.rpi/specs/ negation so behavioral specs are
tracked while research/designs/plans/reviews/diagnoses stay local. Use
--no-track to gitignore the entire .rpi/ tree (specs included). Use
"rpi update" to sync an existing project with the latest workflow files.`,
	Example: `  # Initialize for Claude Code (default)
  rpi init

  # Initialize for OpenCode
  rpi init --target opencode

  # Initialize cross-tool skills only (no tool-specific setup)
  rpi init --target agents-only

  # Initialize in a specific directory without rules file
  rpi init ./my-project --no-claude-md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initNoClaudeMD, "no-claude-md", false, "Skip rules file generation (CLAUDE.md or AGENTS.md)")
	initCmd.Flags().BoolVar(&initNoTrack, "no-track", false, "Gitignore the entire .rpi/ tree, including specs (default: ignore artifacts but track .rpi/specs/)")
	initCmd.Flags().StringVar(&initTarget, "target", "claude", `AI coding tool to initialize for: "claude", "opencode", or "agents-only"`)
	initCmd.Flags().BoolVar(&initNoMCP, "no-mcp", false, "Skip MCP server configuration")
	initCmd.Flags().BoolVar(&initGlobal, "global", false, "Install skills/agents/MCP into the user-level config dir (~/.claude or ~/.config/opencode) instead of a project")
	rootCmd.AddCommand(initCmd)
}

type targetConfig struct {
	toolDir   string // ".claude", ".opencode", or "" for agents-only
	subdirs   []string
	rulesFile string // "CLAUDE.md", "AGENTS.md", or "" for agents-only
	target    workflow.Target
}

// resolveGlobalTargetDir returns the user-level base directory for the given
// target plus an adjusted targetConfig whose toolDir is relative to that
// base, so the rest of the install pipeline composes paths correctly.
// agents-only is unsupported in global mode.
func resolveGlobalTargetDir(cfg targetConfig) (string, targetConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", cfg, fmt.Errorf("resolve home dir: %w", err)
	}
	globalCfg := cfg
	switch cfg.target {
	case workflow.TargetClaude:
		globalCfg.toolDir = ".claude"
		return home, globalCfg, nil
	case workflow.TargetOpenCode:
		globalCfg.toolDir = filepath.Join(".config", "opencode")
		return home, globalCfg, nil
	case workflow.TargetAgentsOnly:
		return "", cfg, fmt.Errorf("--global is not supported with --target agents-only")
	default:
		return "", cfg, fmt.Errorf("unknown target %q", cfg.target)
	}
}

func resolveTargetConfig(t string) (targetConfig, error) {
	switch t {
	case "claude":
		return targetConfig{
			toolDir:   ".claude",
			subdirs:   []string{"skills", "hooks", "agents"},
			rulesFile: "CLAUDE.md",
			target:    workflow.TargetClaude,
		}, nil
	case "opencode":
		return targetConfig{
			toolDir:   ".opencode",
			subdirs:   []string{"skills", "hooks"},
			rulesFile: "AGENTS.md",
			target:    workflow.TargetOpenCode,
		}, nil
	case "agents-only":
		return targetConfig{
			toolDir:   "",
			subdirs:   nil,
			rulesFile: "",
			target:    workflow.TargetAgentsOnly,
		}, nil
	default:
		return targetConfig{}, fmt.Errorf("unknown target %q: must be \"claude\", \"opencode\", or \"agents-only\"", t)
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	cfg, err := resolveTargetConfig(initTarget)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()

	var targetDir string
	if initGlobal {
		if len(args) > 0 {
			return fmt.Errorf("--global and a directory argument are mutually exclusive")
		}
		if initNoTrack {
			return fmt.Errorf("--global is not compatible with --no-track (no .gitignore is touched in global mode)")
		}
		targetDir, cfg, err = resolveGlobalTargetDir(cfg)
		if err != nil {
			return err
		}
	} else {
		targetDir = "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		// Guard: fail if tool dir already exists
		if cfg.target == workflow.TargetAgentsOnly {
			agentsPath := filepath.Join(targetDir, ".agents")
			if _, err := os.Stat(agentsPath); err == nil {
				return fmt.Errorf(".agents/ already exists; use rpi update to sync")
			}
		} else {
			toolDirPath := filepath.Join(targetDir, cfg.toolDir)
			if _, err := os.Stat(toolDirPath); err == nil {
				return fmt.Errorf("%s/ already exists; use rpi update to sync", cfg.toolDir)
			}
		}
	}

	// Create tool subdirs or .agents/skills/ (first-time creation)
	if cfg.toolDir != "" {
		toolDirPath := filepath.Join(targetDir, cfg.toolDir)
		for _, d := range cfg.subdirs {
			path := filepath.Join(toolDirPath, d)
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("create %s: %w", path, err)
			}
			logSuccess(w, fmt.Sprintf("Created %s/%s/", cfg.toolDir, d))
		}
	} else {
		agentsSkillsDir := filepath.Join(targetDir, ".agents", "skills")
		if err := os.MkdirAll(agentsSkillsDir, 0755); err != nil {
			return fmt.Errorf("create .agents/skills/: %w", err)
		}
		logSuccess(w, "Created .agents/skills/")
	}

	// Configure MCP server (Claude only)
	if !initNoMCP && cfg.target == workflow.TargetClaude {
		configureMCP(w, targetDir, initGlobal)
	}

	// Sync shared project structure (dirs, skills, templates, rules, settings,
	// .gitignore entries).
	return syncProject(syncOptions{
		targetDir: targetDir,
		cfg:       cfg,
		skipRules: initNoClaudeMD,
		noTrack:   initNoTrack,
		global:    initGlobal,
		w:         w,
	})
}

func ensureGitignoreEntries(w io.Writer, targetDir string, entries ...string) error {
	if len(entries) == 0 {
		return nil
	}
	gitignorePath := filepath.Join(targetDir, ".gitignore")

	existing := make(map[string]bool)
	if data, err := os.ReadFile(gitignorePath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			existing[line] = true
		}
	}

	var missing []string
	for _, e := range entries {
		if !existing[e] {
			missing = append(missing, e)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n# RPI workflow\n%s\n", strings.Join(missing, "\n")); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	for _, e := range missing {
		logSuccess(w, fmt.Sprintf("Added %s to .gitignore", e))
	}
	return nil
}

// mcpCommandRunner abstracts command execution for testing.
var mcpCommandRunner func(name string, args ...string) ([]byte, error) = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func configureMCP(w io.Writer, _ string, userScope bool) {
	if _, err := exec.LookPath("claude"); err != nil {
		logWarning(w, "claude not found in PATH — skipping MCP server configuration")
		return
	}
	if _, err := exec.LookPath("rpi"); err != nil {
		logWarning(w, "rpi not found in PATH — skipping MCP server configuration")
		return
	}

	// Check if rpi MCP server is already registered
	if out, err := mcpCommandRunner("claude", "mcp", "get", "rpi"); err == nil {
		_ = out
		logWarning(w, "MCP server 'rpi' already configured")
		return
	}

	// Register the MCP server via claude CLI
	args := []string{"mcp", "add", "rpi"}
	if userScope {
		args = append(args, "--scope", "user")
	}
	args = append(args, "--", "rpi", "serve")
	if out, err := mcpCommandRunner("claude", args...); err != nil {
		logWarning(w, fmt.Sprintf("Failed to register MCP server: %s", strings.TrimSpace(string(out))))
		return
	}
	logSuccess(w, "Configured MCP server via claude mcp add")
}

// safeBashPatterns lists Claude Code permission patterns for the safe
// rpi subcommands. Inclusion rule: safe = does not mutate host config
// (~/.claude/, ~/.config/), does not mutate the rpi binary, and does
// not run as a long-lived daemon. New `rpi` subcommands default to
// unsafe — add them here only if they satisfy the safety contract.
var safeBashPatterns = []string{
	"Bash(rpi scan:*)",
	"Bash(rpi search:*)",
	"Bash(rpi status:*)",
	"Bash(rpi version:*)",
	"Bash(rpi context:*)",
	"Bash(rpi resume:*)",
	"Bash(rpi suggest:*)",
	"Bash(rpi chain:*)",
	"Bash(rpi extract:*)",
	"Bash(rpi git-context:*)",
	"Bash(rpi verify:*)",
	"Bash(rpi frontmatter:*)",
	"Bash(rpi scaffold:*)",
	"Bash(rpi archive:*)",
	"Bash(rm /tmp/claude-handoff-*.md)",
}

// configureSettings ensures .claude/settings.json contains the mcp__rpi__*
// permission and the safe rpi Bash allowlist. It merges into any existing
// settings file rather than overwriting.
func configureSettings(w io.Writer, toolDirPath string) {
	settingsPath := filepath.Join(toolDirPath, "settings.json")

	// Read existing settings (if any)
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &raw); err != nil {
			logWarning(w, fmt.Sprintf("Failed to parse %s: %v", settingsPath, err))
			return
		}
	}

	// Parse permissions
	var allow []string
	if permsRaw, ok := raw["permissions"]; ok {
		var perms map[string][]string
		if err := json.Unmarshal(permsRaw, &perms); err == nil {
			allow = perms["allow"]
		}
	}

	existing := make(map[string]bool, len(allow))
	for _, entry := range allow {
		existing[entry] = true
	}

	const rpiPattern = "mcp__rpi__*"
	mcpAdded := false
	if !existing[rpiPattern] {
		allow = append(allow, rpiPattern)
		existing[rpiPattern] = true
		mcpAdded = true
	}

	bashAdded := false
	for _, pattern := range safeBashPatterns {
		if existing[pattern] {
			continue
		}
		allow = append(allow, pattern)
		existing[pattern] = true
		bashAdded = true
	}

	if !mcpAdded && !bashAdded {
		return
	}

	permsMap := map[string][]string{"allow": allow}
	permsJSON, _ := json.Marshal(permsMap)
	raw["permissions"] = json.RawMessage(permsJSON)

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		logWarning(w, fmt.Sprintf("Failed to marshal settings: %v", err))
		return
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		logWarning(w, fmt.Sprintf("Failed to write %s: %v", settingsPath, err))
		return
	}
	if mcpAdded {
		logSuccess(w, "Configured auto-allow for RPI MCP tools in settings.json")
	}
	if bashAdded {
		logSuccess(w, "Configured auto-allow for safe rpi Bash commands in settings.json")
	}
}

// hookDef defines a hook to configure in .claude/settings.json.
type hookDef struct {
	event   string // Hook event name (e.g., "PostCompact")
	marker  string // Unique string to detect if already configured
	command string // Shell command for the hook
	logMsg  string // Success log message
}

// rpiHooks defines all hooks that configureHooks will ensure are present.
var rpiHooks = []hookDef{
	{
		event:   "PostCompact",
		marker:  "rpi_context_essentials",
		command: "cat <<'HOOK_EOF'\nIMPORTANT: Context was compacted. Call the rpi_context_essentials MCP tool to restore your implementation context (active plan phase, spec scenarios, constraints).\nHOOK_EOF",
		logMsg:  "Configured PostCompact hook for context preservation",
	},
	{
		event:   "SessionStart",
		marker:  "rpi_session_resume",
		command: "cat <<'HOOK_EOF'\nCall the rpi_session_resume MCP tool to see active work and suggested next steps.\nHOOK_EOF",
		logMsg:  "Configured SessionStart hook for session resume",
	},
	{
		event:   "Stop",
		marker:  "rpi_suggest_next",
		command: "cat <<'HOOK_EOF'\nCall the rpi_suggest_next MCP tool to determine the appropriate next pipeline step.\nHOOK_EOF",
		logMsg:  "Configured Stop hook for pipeline suggestions",
	},
	// Load-bearing recipe — must stay byte-identical to the path computed in
	// internal/workflow/assets/skills/rpi-handoff/SKILL.md. The drift-guard
	// test TestSessionHandoffHookRecipePinned enforces the pinned shape.
	{
		event:  "SessionStart",
		marker: "claude-handoff",
		command: `HANDOFF="/tmp/claude-handoff-$(echo -n "$PWD" | shasum -a 256 | cut -c1-12).md"; if [ -f "$HANDOFF" ]; then cat <<HOOK_EOF
Pending handoff at $HANDOFF from a previous session. Read it, summarise its instructions, then rm it (consume-once).
HOOK_EOF
fi`,
		logMsg: "Configured SessionStart hook for handoff pickup",
	},
}

// configureHooks ensures .claude/settings.json contains RPI hooks
// (PostCompact, SessionStart, Stop). Merges into existing settings/hooks.
// Existing RPI hook entries are always replaced with the latest version.
func configureHooks(w io.Writer, toolDirPath string) {
	settingsPath := filepath.Join(toolDirPath, "settings.json")

	// Read existing settings (if any)
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &raw); err != nil {
			logWarning(w, fmt.Sprintf("Failed to parse %s: %v", settingsPath, err))
			return
		}
	}

	// Parse existing hooks
	hooks := make(map[string]json.RawMessage)
	if hooksRaw, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
			logWarning(w, fmt.Sprintf("Failed to parse hooks in %s: %v", settingsPath, err))
			return
		}
	}

	added := 0
	for _, h := range rpiHooks {

		// Build the matcher+hooks entry per Claude Code hook format:
		// {"matcher": "", "hooks": [{"type": "command", "command": "..."}]}
		matcherEntry := map[string]interface{}{
			"matcher": "",
			"hooks": []map[string]string{
				{
					"type":    "command",
					"command": h.command,
				},
			},
		}

		// Remove existing entries containing our marker before appending
		var entries []json.RawMessage
		if existing, ok := hooks[h.event]; ok {
			var all []json.RawMessage
			json.Unmarshal(existing, &all)
			for _, e := range all {
				if !strings.Contains(string(e), h.marker) {
					entries = append(entries, e)
				}
			}
		}
		entryJSON, _ := json.Marshal(matcherEntry)
		entries = append(entries, json.RawMessage(entryJSON))

		entriesJSON, _ := json.Marshal(entries)
		hooks[h.event] = json.RawMessage(entriesJSON)

		logSuccess(w, h.logMsg+" in settings.json")
		added++
	}

	if added == 0 {
		return // all hooks already configured
	}

	hooksJSON, _ := json.Marshal(hooks)
	raw["hooks"] = json.RawMessage(hooksJSON)

	// Write back with indentation
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		logWarning(w, fmt.Sprintf("Failed to marshal settings: %v", err))
		return
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		logWarning(w, fmt.Sprintf("Failed to write %s: %v", settingsPath, err))
		return
	}
}
