# The `rpi init` Command

The `rpi init` command bootstraps the RPI workflow into any project. All workflow files (skills and scaffold templates) are embedded in the binary -- no external dotfiles or source repo needed.

Supports three targets: **Claude Code** (default), **OpenCode**, and **agents-only**.

```bash
# Initialize current directory (Claude Code)
rpi init

# Initialize for OpenCode
rpi init --target opencode

# Initialize a specific project directory
rpi init ~/projects/my-app
rpi init ~/projects/my-app --target opencode

# Options
rpi init --no-claude-md     # Skip rules file generation (CLAUDE.md or AGENTS.md)
rpi init --no-mcp           # Skip MCP server registration
rpi init --no-track         # Gitignore the entire .rpi/ tree (specs included)
```

## What it creates

### Claude Code target (default)

- `.claude/skills/` -- Agent Skills (rpi-research, rpi-propose, rpi-plan, rpi-implement, rpi-verify, rpi-commit, rpi-archive, rpi-diagnose, rpi-explain, rpi-spec-sync)
- `CLAUDE.md` -- Project-level instructions for Claude Code
- MCP server registered via `claude mcp add rpi -- rpi serve`
- `.claude/settings.json` -- Auto-allow permissions for RPI MCP tools

### OpenCode target

- `.opencode/skills/` -- Agent Skills (same set, with provider-qualified model IDs)
- `AGENTS.md` -- Project-level instructions for OpenCode

### Agents-only target

- `.agents/skills/` -- Cross-tool Agent Skills (no tool-specific directory, no MCP config)

### Shared (all targets)

- `.rpi/` -- Directory structure for pipeline artifacts. By default `.gitignore` is updated with `.rpi/*` and `!.rpi/specs/`, so specs are tracked while research/designs/plans/reviews/diagnoses stay local. Use `--no-track` to gitignore the entire `.rpi/` tree (specs included).
- `.rpi/templates/` -- Scaffold templates for plans, designs, research docs, specs, etc.

### MCP Server Configuration

When the target is `claude`, `rpi init` auto-registers an MCP server so the AI calls typed tools (`rpi_scaffold`, `rpi_scan`, etc.) instead of shelling out to the CLI.

- Requires both `rpi` and `claude` to be in PATH
- Skipped with `--no-mcp` or when the target is `opencode`
- Warns and continues (doesn't fail) if `rpi` or `claude` are not found, or if the server is already registered
- Uses `claude mcp add rpi -- rpi serve` to register (or `claude mcp add rpi --scope user -- rpi serve` under `--global`)
- Use `rpi update` to sync an existing project (see `rpi update --help`)

## Global setup (`--global`)

Pass `--global` to `rpi init` to install RPI's skills, agents, MCP server registration, and `settings.json` hooks/permissions into the user-level config directory instead of the current project. After a one-time global install, every Claude Code (or OpenCode) session has the RPI skills available without per-project setup.

```bash
# Claude Code → ~/.claude/
rpi init --global

# OpenCode → ~/.config/opencode/
rpi init --global --target opencode
```

What `--global` writes:

- `~/.claude/skills/` (or `~/.config/opencode/skills/`) — the full RPI skill set, with `.bak`-on-diff semantics for files you've customized.
- `~/.claude/agents/` — Claude target only.
- `~/.claude/settings.json` — `mcp__rpi__*` permission, the safe `Bash(rpi …)` allowlist, and the `PostCompact` / `SessionStart` / `Stop` hooks. Existing keys (yours and other tools') are preserved.
- MCP server registration via `claude mcp add rpi --scope user -- rpi serve`, so the registration is available from any working directory rather than just where the command was run.

What `--global` does **not** touch:

- No `~/.claude/CLAUDE.md` or `~/.config/opencode/AGENTS.md` is written — those are user-curated personal config.
- No `~/.rpi/` tree is created — `.rpi/` artifacts remain per-project.
- No `~/.gitignore` modifications.

Conflicts (each returns an error):

- `rpi init --global ./somewhere` — `--global` and a positional directory are mutually exclusive.
- `rpi init --global --no-track` — `--no-track` controls `.gitignore` policy, which `--global` never touches.
- `rpi init --global --target agents-only` — agents-only has no canonical user-level home; not supported in v1.

### Refreshing the global install

```bash
rpi update --global
rpi update --global --target opencode
```

Refreshes skills, agents (Claude target), and the `settings.json`
hooks/permissions in the user-level config dir. Same `.bak`-on-diff
semantics as the project-mode update — your customizations are preserved
to a sibling `.bak` file before the latest embedded version is written.
No project-level artifacts (`.rpi/`, rules file, `.gitignore`) are
created or modified.

## Auto-bootstrap on first skill use

Once you've run `rpi init --global`, every `rpi-*` skill auto-initializes
the project's `.rpi/` tree, rules file, and `.gitignore` policy at the
git root the first time it's invoked in a fresh repo. No `rpi init`
needed per project.

Each `rpi-*` SKILL.md prompt opens with a one-line preamble:

> Before doing anything else, run `rpi bootstrap` (silent and idempotent
> — initializes `.rpi/` if a global install is present and the project
> hasn't been bootstrapped yet).

The `rpi bootstrap` subcommand has four exit paths:

- **Silent no-op** when the project already has a `.rpi/` (anywhere up
  the directory tree).
- **Silent no-op** when no user-level install exists at
  `~/.claude/skills/rpi-research/` or `~/.config/opencode/skills/rpi-research/`.
- **Silent no-op** when the cwd is not inside a git repository.
- **Auto-init** otherwise: creates `.rpi/<subdirs>/`, the rules file
  (`CLAUDE.md` or `AGENTS.md` per detected target), and the standard
  `.rpi/*` + `!.rpi/specs/` `.gitignore` entries — all at the git root,
  regardless of which subdirectory the user is in. Prints exactly one
  line to stderr:
  `✓ Auto-initialized .rpi/ in <git-root> — skills inherited from <global-path>`

You can run `rpi bootstrap` directly any time — it's safe and idempotent.

### What auto-bootstrap does NOT do

It deliberately keeps the project's footprint minimal:

- No `./.claude/` or `./.opencode/` directory is created — skills,
  agents, hooks, and permissions are inherited from the global install.
- No MCP server registration (already user-scope from `rpi init --global`).
- No `settings.json` is written.

If you later want a fully project-local install (skills, agents, etc.
in `./.claude/`), run `rpi init` explicitly.

### OpenCode users

Auto-bootstrap is currently triggered by the same skill preamble used
for the Claude target — the SKILL.md content is shared across both. If
you're using OpenCode without a global install, simply run
`rpi bootstrap` manually in a project to get the same lite-init.

## Installation

Build and install the `rpi` binary:

```bash
make install
```

This builds the Go binary and copies it to `~/.local/bin/rpi`. Make sure `~/.local/bin` is in your PATH.

## Shell Completion

The `rpi` binary supports autocompletion for bash, zsh, fish, and powershell. Add one of the following lines to your shell configuration file to enable completions for every new session:

```bash
# Bash (~/.bashrc)
source <(rpi completion bash)

# Zsh (~/.zshrc)
source <(rpi completion zsh)

# Fish (~/.config/fish/config.fish)
rpi completion fish | source
```

Run `rpi completion <shell> --help` for more options.
