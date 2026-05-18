---
domain: rpi init / rpi update settings configuration for Claude target
feature: rpi-bash-auto-accept
last_updated: 2026-05-18T12:00:00+02:00
updated_by: .rpi/plans/2026-05-18-remove-rpi-bootstrap-preamble-from-skills.md
---

# rpi-bash-auto-accept

## Purpose

When initializing or updating an RPI-managed project for Claude, seed the project's shared Claude permissions file with a curated allowlist of safe `rpi` Bash invocations so Claude can run them without per-call permission prompts, while continuing to prompt for commands that mutate host config or the rpi binary itself.

## Scenarios

### Initial setup seeds the safe Bash allowlist
Given a directory with no prior Claude permissions configured
When the user runs `rpi init` for the Claude target
Then the resulting permissions file allows the curated set of safe `rpi` subcommand invocations alongside the existing MCP allow entry, and prompts are no longer raised for those commands in subsequent sessions

### Unsafe rpi subcommands are not auto-accepted
Given a project initialized for the Claude target
When Claude attempts to run an unsafe rpi subcommand (one that mutates host config, the installed binary, or runs as a long-lived daemon)
Then the user is prompted to approve it instead of it being auto-accepted

### Update applies the allowlist to projects initialized before this feature
Given a project that was initialized before the safe Bash allowlist existed and currently has only the MCP allow entry
When the user runs `rpi update` for the Claude target
Then the safe Bash allowlist entries are added to the existing permissions, and any pre-existing user-managed allow entries are preserved unchanged

### Re-running setup does not duplicate or reorder entries
Given a project that already has the safe Bash allowlist in its permissions
When the user re-runs `rpi init` (in a fresh dir) or `rpi update`
Then no allowlist entry is duplicated, the existing order of entries is preserved, and the file diff is empty for the permissions block

### Allowlist is independent of MCP wiring
Given the user runs `rpi init --no-mcp` or skips MCP wiring for any reason on the Claude target
When initialization completes
Then the safe Bash allowlist is still written to the project's permissions, even though no `mcp__rpi__*` MCP entry is added

### Targets other than Claude are unaffected
Given the user runs `rpi init --target opencode` or `rpi init --target agents-only`
When initialization completes
Then no Claude permissions file is created or modified by this feature

### Existing user permissions are preserved
Given a project whose permissions file already contains user-managed allow entries unrelated to rpi
When `rpi init` or `rpi update` runs for the Claude target
Then those user entries remain intact, and the safe Bash allowlist entries are appended without removing, reordering, or rewriting them

## Constraints

- The allowlist applies only to the Claude target — OpenCode and `agents-only` targets are not modified by this feature.
- The set of "safe" rpi subcommands is bounded by the rule *"does not mutate host config (`~/.claude/`, `~/.config/`), does not mutate the rpi binary, and does not run as a long-lived daemon"*; commands that fail any of those tests are excluded.
- The allowlist is written to the project-shared permissions file (the one normally checked into version control), not the user-local override.
- Writing the allowlist must be idempotent across repeated runs and must not overwrite or reorder pre-existing entries.
- The feature is independent of MCP availability: it must work when MCP is wired and when MCP is skipped.
- A user-local override that already grants the same patterns must continue to function — the shared file augments, never conflicts with, the local one.

## Out of Scope

- Auto-accepting any rpi subcommand classified as unsafe (today: `init`, `update`, `upgrade`, `serve`).
- Configuring permissions for OpenCode, `agents-only`, or other future targets.
- Removing or narrowing the existing `mcp__rpi__*` MCP allow wildcard.
- Writing `permissions.deny` rules for unsafe commands.
- Migrating or rewriting entries already present in the user-local permissions override.
- Surface UI/UX (no new flag, no prompt, no opt-out) — the seeding is implicit in init/update.
