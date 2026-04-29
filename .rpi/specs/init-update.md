---
domain: init-update-cleanup
feature: init-update
last_updated: 2026-04-29T00:00:00+02:00
updated_by: .rpi/diagnoses/2026-04-29-rpi-init-gitignore-default.md
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
Given an initialized project with customized skill files or CLAUDE.md
When `rpi update` runs
Then modified files are backed up to `.bak` and replaced with the latest embedded versions

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

## Constraints
- Init and update remain separate Cobra commands
- Do not introduce new dependencies or packages

## Out of Scope
- Merging init/update into one command
- Changes to skill content or template content
