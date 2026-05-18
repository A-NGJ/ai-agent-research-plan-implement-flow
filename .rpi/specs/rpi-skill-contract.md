---
domain: rpi init / rpi update rules-file management
feature: rpi-skill-contract
last_updated: 2026-05-18T12:01:03+02:00
updated_by: .rpi/plans/2026-05-18-append-missing-rules-file-sections-on-bootstrap-and-update.md
---

# rpi-skill-contract

## Purpose

Factor cross-skill invariants — search before drafting, `--ff`/`--grill` flag semantics, pipeline order — out of every rpi-* skill body and into a single delimited block written by `rpi init` and `rpi update` into the project's rules file. The host agent auto-loads the rules file once per session, so the rules apply across every skill without each skill restating them.

## Scenarios

### Initial setup writes the contract block for Claude
Given an empty directory
When the user runs `rpi init` with the Claude target
Then the project's `CLAUDE.md` contains a delimited block of cross-skill rules covering at minimum: searching for prior artifacts before drafting, the `--ff` and `--grill` flag contract, and the pipeline order

### Initial setup writes the contract block for OpenCode
Given an empty directory
When the user runs `rpi init` with the OpenCode target
Then the project's `AGENTS.md` contains the same delimited block of cross-skill rules

### agents-only target writes no rules file and no contract block
Given an empty directory
When the user runs `rpi init` with the `agents-only` target
Then no rules file is created, no contract block appears anywhere on disk, and the rpi-* skills still install successfully

### Update inserts the contract block into a project that predates the feature
Given an already-initialized project whose rules file has no contract block
When the user runs `rpi update`
Then the rules file gains the contract block at its end (or in its conventional location), and no other content in the file is removed, reordered, or rewritten

### Update refreshes stale contract content without touching surrounding edits
Given an initialized project whose rules file contains a contract block with outdated content and additional user-added sections before and after the block
When the user runs `rpi update`
Then the contents inside the contract delimiters are replaced with the current version, and every line outside the delimiters is preserved — with the sole exception that missing top-level template sections may be appended at EOF (see "Update appends missing template sections at EOF")

### Update appends missing template sections at EOF
Given an initialized project whose rules file is missing one or more `## Heading` sections present in the current rendered template
When the user runs `rpi update`
Then each missing section is appended to the end of the rules file in template order, separated by blank lines, and every existing line (including the contract block and user-added sections) is preserved

### Section reconciliation is idempotent
Given an initialized project whose rules file already contains every `## Heading` from the current rendered template
When the user runs `rpi update` a second time
Then no section is appended, the rules file mtime is unchanged, and the file is not rewritten

### Drifted section bodies are left alone
Given an initialized project whose rules file contains all template `## Heading`s but the body of one section has been edited by the user
When the user runs `rpi update`
Then the edited body is preserved byte-for-byte and no template body overrides the user edit

### Re-running update is idempotent
Given an initialized project whose rules file already contains the current contract block
When the user runs `rpi update` a second time
Then the rules file is not rewritten, no backup is created, and no diff is observable in the file

### Skipping the rules file leaves the contract block untouched
Given an initialized project
When the user runs `rpi update --no-claude-md`
Then the rules file is not opened, the contract block is not refreshed, and any previous content (including a stale block) is left intact

### Malformed contract delimiters are left untouched
Given a rules file containing a begin delimiter but no matching end delimiter (or vice versa)
When the user runs `rpi update`
Then the writer leaves the file unchanged, surfaces a warning naming the affected file, and exits without error so the rest of the update completes

### Trimmed skills still trigger and operate without the contract block
Given an rpi-* skill whose body no longer contains the cross-skill prose
When the skill is invoked in an environment where the contract block has not been loaded (for example, the agents-only target, or a project whose rules file was deleted)
Then the skill still triggers on its description, runs its skill-specific invariants, and produces a usable result — at most relying on tool-surface hints (such as the semantic-search backend's recovery hint) rather than in-body reminders

## Constraints

- Skill bodies remain the source of truth for skill-specific behavior. Only invariants that apply across two or more rpi-* skills migrate into the contract block.
- The contract block is delimited so that idempotent in-place rewrites are possible across `rpi update` runs; user content outside the delimiters is never modified or reordered. The writer may only *append* missing top-level template sections at EOF — it never edits or replaces user content elsewhere in the file.
- The contract block carries a visible version marker inside its opening delimiter so future writers can detect and migrate older formats.
- The bootstrap preamble at the top of every rpi-* skill body is retained — the contract block cannot solve the cold-start case where the rules file does not yet exist.
- The contract block is written only when the target writes a rules file. Targets that do not produce a rules file (today: `agents-only`) receive no contract block, and rpi-* skills must still function in that environment.
- The contract block content is shared between Claude and OpenCode targets — no per-target wording divergence.
- The writer fails closed: malformed or partially-edited contract blocks cause it to warn and skip, never to attempt repair that could destroy user content.
- The behavior is independent of MCP configuration — the contract block is written regardless of `--no-mcp`.

## Out of Scope

- Moving the bootstrap preamble out of skill bodies (handled by a separate, possibly future, hook-based proposal).
- Trimming skill-specific invariants (the per-skill prose that does not apply across skills stays in skill bodies).
- A new sidecar file (e.g., `.rpi/CONTRACT.md`) referenced via an include directive — the contract lives inside the existing rules file.
- Migrating user-local rules files outside the project (`~/.claude/CLAUDE.md`, `~/.config/opencode/AGENTS.md`) — those are touched only by `rpi init --global` / `rpi update --global` under their existing semantics.
- Extending the skill-description eval to measure invariant adherence (a separate proposal).
- A user-visible flag or opt-out to suppress the contract block — the block is part of the rules file the same way the project-overview sections are.
- Automatic recovery from a manually corrupted contract block — the user is expected to delete the broken block and re-run `rpi update`, which will re-insert a fresh copy.
