---
domain: rpi distribution via claude code plugin
feature: rpi-claude-plugin
last_updated: 2026-05-20T18:30:00+02:00
updated_by: .rpi/plans/2026-05-20-rpi-as-claude-code-plugin-package.md
---

# rpi-claude-plugin

## Purpose

Distribute RPI as a Claude Code plugin so users can install it from the marketplace in one click and run an explicit setup skill to fetch the binary. The plugin owns all Claude Code wiring (skills, hooks, MCP server, permissions) inside its own directory and writes nothing to user-owned configuration files; the existing standalone `rpi init` install path remains available unchanged for non-plugin environments.

## Scenarios

### First-time plugin install and setup
Given the user has installed the RPI plugin from the marketplace and has no prior standalone RPI install
When the user runs `/rpi:rpi-setup`
Then the platform-matching RPI binary is downloaded from GitHub Releases, its SHA256 is verified against the release's checksums file, the binary is placed at `~/.rpi/bin/rpi` with executable permissions, the installed version is reported back, and the user sees a hint about adding `~/.rpi/bin` to their PATH for CLI use

### Setup refuses to overwrite an existing standalone install
Given the user has previously run `rpi init --global`, leaving RPI skills under `~/.claude/skills/rpi-*` and an `rpi` MCP server registered with Claude Code
When the user installs the plugin and runs `/rpi:rpi-setup`
Then setup does not download or install any binary, the existing standalone install is left untouched, and the user is instructed to run `rpi uninstall --global` before re-running `/rpi:rpi-setup`

### Re-running setup upgrades an already-installed binary
Given `~/.rpi/bin/rpi` exists and runs, and a newer release is available on GitHub
When the user runs `/rpi:rpi-setup` a second time
Then setup delegates to the existing upgrade flow, the binary is replaced with the newer version, and no other files in `~/.rpi/` or the plugin directory are modified

### Setup aborts when the checksum does not match
Given the user runs `/rpi:rpi-setup` and the downloaded archive's SHA256 does not match the value in `checksums.txt`
When verification fails
Then no binary is installed at `~/.rpi/bin/rpi`, any partially downloaded files in the temporary location are removed, and the user sees an error naming the expected and observed checksums

### Workflow context is injected by SessionStart without modifying user files
Given the user has installed the plugin and completed `/rpi:rpi-setup`
When the user starts a new Claude Code session
Then Claude is provided context describing the RPI pipeline (Research → Propose → Plan → Implement → Verify), when to invoke each skill, and the cross-skill flag contract — and no content has been written to `~/.claude/CLAUDE.md`, `~/.claude/settings.json`, or any project-level `CLAUDE.md`/`AGENTS.md`

### Uninstalling the plugin without the binary leaves the system in a known state
Given the user has installed the plugin and run `/rpi:rpi-setup`, then uninstalls the plugin via Claude Code
When the plugin is removed
Then the plugin's skills, hooks, MCP server registration, and permissions are no longer active in Claude Code, the binary at `~/.rpi/bin/rpi` remains on disk, and running `rpi uninstall --global` afterwards cleanly removes the binary and `~/.rpi/`

### `rpi uninstall --global` in plugin mode removes only the binary
Given the binary at `~/.rpi/bin/rpi` exists, no `~/.claude/skills/rpi-*` directories are present, and no `rpi` MCP server is registered outside the plugin
When the user runs `rpi uninstall --global`
Then the binary is removed, `~/.rpi/` is removed if empty, no entries are removed from `~/.claude/settings.json`, no plugin files in the Claude Code plugin cache are touched, and the user sees a summary of what was removed

### `rpi uninstall --global` in standalone mode removes the full standalone install
Given the user has previously run `rpi init --global` (skills under `~/.claude/skills/rpi-*`, `rpi` MCP server registered, hooks and permissions in `~/.claude/settings.json`, binary in PATH or at the install-script location)
When the user runs `rpi uninstall --global`
Then the standalone skills directories are removed, the `rpi` MCP server registration is cleared, the RPI-owned hooks and permissions are removed from `~/.claude/settings.json` while preserving user-added entries, the binary is removed from its install location, and the user sees a summary of what was removed

### Standalone `rpi init` remains available for non-plugin environments
Given the user is in an environment where the plugin marketplace is unavailable or undesired (locked-down system, opencode target, manual install preference)
When the user runs `rpi init --global` or `rpi init --target opencode`
Then the command continues to install skills, hooks, MCP server, permissions, and rules-file content exactly as it does today, with no behavior change from the plugin's existence

## Constraints

- The plugin's `mcpServers` entry registers an MCP server named `rpi`, matching the standalone install. Tool prefixes (`mcp__rpi__*`) are unchanged so existing skill bodies and permission allowlists remain valid.
- Plugin skill folder names keep the `rpi-` prefix so the trigger surface stays unambiguous in Claude Code's slash-command picker (which may display only the skill name without the plugin namespace, leaving bare `/plan` to collide with built-in commands). Commands surface as `/rpi:rpi-<name>` (for example `/rpi:rpi-plan`, `/rpi:rpi-implement`). The MCP server name and CLI command name remain `rpi`.
- `/rpi:rpi-setup` writes the binary only to `~/.rpi/bin/rpi`. It does not place the binary on the user's PATH, modify shell rc files, write to `~/.local/bin`, or use `sudo`.
- `/rpi:rpi-setup` downloads binaries only from official GitHub Releases under `A-NGJ/rpi`. No third-party mirrors. The download is verified against the `checksums.txt` asset of the same release.
- The plugin targets Claude Code only. The manifest declares no Cowork support and no fallback to other harnesses.
- The plugin writes nothing to `~/.claude/CLAUDE.md`, `~/.claude/settings.json`, any project-level `CLAUDE.md`/`AGENTS.md`, or any user-owned configuration file. The only filesystem effect outside the plugin's own directory is the binary at `~/.rpi/bin/rpi`, written by `/rpi:rpi-setup`.
- `rpi uninstall --global` must be safe to run when nothing is installed (exit cleanly with a clear message), idempotent across re-runs, and must not delete files it cannot prove were installed by RPI.
- Detection of standalone vs plugin mode in `rpi uninstall --global` is based on observable filesystem and configuration state, not on a marker file written at install time — so a user who installed via either path is handled correctly without prior coordination.
- Plugin install does not run any setup automatically. The user must explicitly invoke `/rpi:rpi-setup`; no SessionStart, PostCompact, or Stop hook installs the binary on the user's behalf.

## Out of Scope

- Cowork distribution and any Cowork-specific code path.
- Committing per-platform binaries inside the plugin directory.
- Auto-installing the binary on first MCP server launch or first SessionStart fire.
- Writing a markered RPI block into `~/.claude/CLAUDE.md`, project `CLAUDE.md`, or `AGENTS.md`.
- Automatic per-project bootstrap (creating `.rpi/` subdirectories or writing rules-file blocks at the project level).
- A `/rpi:uninstall` skill bundled with the plugin (uninstall is the binary's `rpi uninstall --global` subcommand).
- Migration tooling that converts a standalone install into the plugin install in one step (the documented migration is to run `rpi uninstall --global`, then install the plugin and run `/rpi:rpi-setup`).
- Plugin distribution for opencode or other non-Claude-Code harnesses (those continue to use `rpi init --target <name>`).
- Version coupling or minimum-version checks between plugin and binary (the design ships with no enforcement; added later if drift becomes a real problem).
- Changes to the content of skills' instructional bodies beyond cross-reference renames (`/rpi-X` → `/rpi:rpi-X`).
- Changes to `rpi init`'s rules-file or settings.json writing behavior — those remain governed by `.rpi/specs/init-update.md` and `.rpi/specs/rpi-skill-contract.md`.
