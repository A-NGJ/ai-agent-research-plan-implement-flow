---
domain: init-update-cleanup
feature: init-update
last_updated: 2026-05-18T00:00:00+02:00
updated_by: .rpi/plans/2026-05-18-rpi-skill-contract-block-in-rules-file.md
---

# Init/Update Cleanup

## Purpose

Ensure `rpi init` and `rpi update` share a single code path for project synchronization. Update always installs the latest managed files, creating `.bak` backups of any files that differ before overwriting.

## Scenarios

### Init and update share sync logic
Given the codebase
When inspecting the init and update command implementations
Then both call a shared sync function for installing skills, templates, rules file, and settings

### Update backs up modified files before overwriting
Given an initialized project with customized skill files
When `rpi update` runs
Then modified files are backed up to `.bak` and replaced with the latest embedded versions. The project rules file (CLAUDE.md / AGENTS.md) is the carve-out: its content is preserved in place, and only the fenced RPI Skill Contract block is refreshed (see `.rpi/specs/rpi-skill-contract.md`).

### Update skips backup when file content is identical
Given an initialized project with no local modifications to managed files
When `rpi update` runs
Then no `.bak` files are created and no files are rewritten

### Init creates fresh project with rules file
Given an empty directory
When `rpi init` runs
Then a `.rpi/` directory is created with subdirectories, CLAUDE.md is written, and the MCP server is registered

### Init rejects existing tool directory
Given a directory with `.claude/` already present
When `rpi init` runs
Then it fails with an error indicating the tool directory already exists

### Update auto-detects target from existing directories
Given a project initialized with `--target opencode`
When `rpi update` runs without specifying a target
Then it detects the OpenCode target and operates on `.opencode/`

### Init gitignores .rpi/ artifacts but keeps specs tracked by default
Given an empty directory
When `rpi init` runs without `--no-track`
Then `.gitignore` contains `.rpi/*` and `!.rpi/specs/` so behavioral specs are checked in while research, designs, plans, reviews, and diagnoses stay local

### Init with --no-track gitignores the entire .rpi/ tree
Given an empty directory
When `rpi init --no-track` runs
Then `.gitignore` contains `.rpi/` (no negation), excluding specs from version control as well

### Update applies gitignore policy on existing projects
Given an initialized project whose `.gitignore` is missing the current policy entries
When `rpi update` runs
Then `.gitignore` is appended with `.rpi/*` and `!.rpi/specs/` (and `.claude/` for the Claude target), without removing or rewriting existing lines

## Constraints
- Init and update remain separate Cobra commands
- Do not introduce new dependencies or packages
- Rules-file updates beyond the contract block are governed by `.rpi/specs/rpi-skill-contract.md`; update preserves user content outside the contract fence byte-for-byte.

## Out of Scope
- Merging init/update into one command
- Changes to skill content or template content
