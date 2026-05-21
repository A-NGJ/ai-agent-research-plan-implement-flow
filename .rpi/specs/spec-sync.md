---
domain: rpi-spec-sync
feature: spec-sync
last_updated: 2026-05-21T11:11:08+02:00
updated_by: .rpi/designs/2026-05-21-spec-drift-scan-mcp-tool-for-spec-sync.md
---

# rpi-spec-sync Skill

## Purpose

A Claude skill that syncs specs to match the current codebase — the reverse of verification. Uses a deterministic drift-scan tool for structural signals, then applies LLM judgment to rewrite, rename, archive, or keep each spec.

## Scenarios

### Scan identifies stale specs
Given specs exist in `.rpi/specs/` and some have not been updated in over 30 days while related code has changed
When the user runs `/rpi-spec-sync`
Then a drift report is presented showing which specs are flagged with the reason for each flag

### Scan identifies obsolete specs
Given a spec describes a feature that has been removed from the codebase
When the user runs `/rpi-spec-sync`
Then the spec is flagged for archival with an explanation of what's missing

### Scan identifies naming mismatches
Given a spec's filename does not match its `feature` frontmatter field
When the user runs `/rpi-spec-sync`
Then the spec is flagged for rename with the suggested new filename

### Scan identifies orphaned specs
Given a spec is not referenced by any other `.rpi/` artifact
When the user runs `/rpi-spec-sync`
Then the spec is surfaced in the drift report so the user can decide whether to keep, merge, or archive it

### Drift scan is available outside the skill
Given the rpi binary is installed
When the user runs `rpi spec-drift scan` from the project root
Then per-spec drift signals are printed as JSON with one record per spec containing the spec path, the list of fired signals, and supporting details

### Drift scan results are deterministic
Given the same project state and the same `.rpi/specs/` contents
When the user runs the drift scan twice in succession
Then both runs produce identical signals for every spec

### Drift scan does not require optional services
Given the qmd semantic-search daemon is not running
When the user runs the drift scan
Then it completes successfully with all structural signals; semantic near-duplicate detection is the skill's responsibility and is reported as unavailable if attempted

### User approves actions before execution
Given the drift report contains flagged specs with proposed actions (rewrite, rename, archive, keep)
When the user reviews the proposals
Then no changes are made until the user explicitly approves each action or approves all at once

### Rewrite updates scenarios to match code
Given a spec is approved for rewrite
When the skill executes the rewrite
Then the scenarios are updated to reflect current code behavior while preserving the spec's domain, feature name, and constraints

### Archive removes obsolete specs cleanly
Given a spec is approved for archival
When the skill executes the archive
Then the spec is moved to `.rpi/archive/` and all references in other artifacts are updated

### Rename updates filename and references
Given a spec is approved for rename
When the skill executes the rename
Then the file is renamed to match the feature field and all references in other artifacts are updated

### Merge combines overlapping specs
Given two or more specs cover closely related or overlapping behavior
When the skill flags them for merge and the user approves
Then the scenarios are combined into a single spec with a unified feature name, the source specs are archived, and all references are updated

## Constraints
- Structural drift signals must be produced by a deterministic, reproducible mechanism — same inputs produce same outputs across runs
- The drift scan must run on a stock checkout without external services (no qmd, no network)
- Semantic near-duplicate detection (qmd-backed) stays in the skill and is optional
- Must get user confirmation before any destructive action (rewrite, archive, rename)
- Rewrites preserve the spec's domain, feature field, and constraints — only scenarios change
- Archives use the existing archive flow (move to `.rpi/archive/`, update frontmatter)
- Reference updates cover all `.rpi/` artifacts (plans, designs, reviews, archives)
- Specs with frontmatter `orphaned: false` are not flagged as orphaned (opt-out for intentional meta-specs)

## Out of Scope
- Creating new specs from scratch (that's `/rpi-propose`)
- Syncing non-spec artifacts (plans, designs, research)
- Unattended/automated execution without user confirmation
- Cross-spec semantic similarity inside the drift-scan tool — that path lives in the skill

## Update Log

- **2026-05-21** — Split structural drift detection out of the skill into a dedicated deterministic tool (`rpi spec-drift scan` / `rpi_spec_drift_scan`). Removed the constraint forbidding new CLI commands and Go code. Added scenarios for tool availability, determinism, and qmd-independence. Added the `orphaned` signal as a first-class scan output and the `orphaned: false` opt-out. Design: `.rpi/designs/2026-05-21-spec-drift-scan-mcp-tool-for-spec-sync.md`.
