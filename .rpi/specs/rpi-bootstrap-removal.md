---
domain: rpi CLI surface
feature: rpi-bootstrap-removal
last_updated: 2026-05-21T10:40:43+02:00
updated_by: .rpi/designs/2026-05-21-remove-rpi-bootstrap.md
---

# rpi-bootstrap-removal

## Purpose

Document the intentional absence of the `rpi bootstrap` subcommand after
its responsibilities were absorbed by `rpi update` and the Claude Code
plugin install flow. This spec describes the user-observable surface that
results from the removal so future contributors know the command is gone
on purpose, not by oversight.

## Scenarios

### `rpi bootstrap` is not a recognized subcommand
Given a user with the `rpi` binary installed
When the user runs `rpi bootstrap`
Then the CLI prints an unknown-command error to stderr and exits with a
non-zero status, and no project files are read or written

### `rpi --help` does not list bootstrap
Given a user with the `rpi` binary installed
When the user runs `rpi --help` or `rpi help`
Then the listed subcommands include `init`, `update`, and the other
artifact and verification commands, and do not include `bootstrap`

### `rpi init` continues to initialize a fresh project
Given an empty directory inside a git repository
When the user runs `rpi init`
Then the project is fully initialized — `.rpi/` subdirectories, the rules
file, the gitignore entries, the embedded skills, and the MCP/hook
configuration are written as documented for `rpi init`, with no
dependency on a separate bootstrap step

### `rpi update` continues to reconcile the rules file on an existing project
Given an initialized project whose rules file is missing one or more
top-level template sections, or whose RPI Skill Contract block has drifted
from the current template
When the user runs `rpi update`
Then the missing top-level sections are appended at end of file in
template order, the contract block is refreshed in place, and user
content outside the contract fences is preserved

### Plugin-mode setup does not include a per-project bootstrap step
Given a user installing the Claude Code plugin via the documented flow
When the user follows the install instructions (`/plugin marketplace add`,
`/plugin install`, `/rpi:rpi-setup`, restart)
Then no step in the documented flow invokes `rpi bootstrap` and the user
reaches a working RPI workflow without ever running it

### Documentation does not advertise bootstrap as a setup or maintenance step
Given the project's README and the `docs/` directory
When a user reads the install, setup, or maintenance instructions
Then no section instructs the user to run `rpi bootstrap` and no fenced
code block contains the literal `rpi bootstrap` command

### Existing workflows that previously called `rpi bootstrap` fail loudly, not silently
Given a user with a shell alias, CI step, or script that invokes
`rpi bootstrap` after upgrading to a release that removes the command
When that workflow runs
Then the CLI exits non-zero with a visible unknown-command error so the
user can switch to `rpi update` for rules-file reconciliation or
`rpi init` for fresh setup

## Constraints

- Rules-file reconciliation behavior — appending missing top-level
  template sections and refreshing the RPI Skill Contract block — is
  preserved end-to-end. The contract for that behavior continues to live
  in `.rpi/specs/init-update.md` and `.rpi/specs/rpi-skill-contract.md`,
  scoped to `rpi update` only.
- `rpi init` and `rpi init --global` behavior is unchanged.
- The safe-bash allowlist written by `rpi init` does not include any
  `rpi bootstrap` permission entry. (It already did not at the time of
  this spec; the spec records the steady state.)
- The plugin install flow remains the recommended path; the standalone
  install flow (`rpi init --global` plus per-project `rpi init`) is
  retained for users who prefer it.
- Removal is hard, not soft — there is no deprecation alias, no warning
  shim, and no transitional command that prints a notice and exits zero.

## Out of Scope

- The behavior of `rpi update` itself — already specified in
  `.rpi/specs/init-update.md`.
- The behavior of the RPI Skill Contract block writer — already specified
  in `.rpi/specs/rpi-skill-contract.md`.
- The `/rpi:rpi-setup` skill behavior — already specified in
  `.rpi/specs/rpi-setup.md`.
- The plugin's MCP server, session-start hook, post-compact hook, or stop
  hook behavior.
- Historical references to `rpi bootstrap` in archived research, designs,
  plans, or reviews under `.rpi/`. Those are time-bound snapshots and
  remain readable as historical record.
- A formal deprecation policy or release-notes template for future
  command removals.
