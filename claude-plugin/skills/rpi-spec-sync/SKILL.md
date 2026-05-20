---
name: rpi-spec-sync
description: Detect spec drift and sync .rpi/specs/ to the current codebase — rewrite, rename, merge, or archive obsolete specs. Use when user says 'sync the specs', 'the spec is stale', 'specs need updating', or 'clean up old specs'.
---

# Spec Sync

## Goal

Treat the codebase as ground truth and sync specs to match it — the reverse of verification. Scan all specs for drift, then propose and execute actions with user approval.

Two phases:
- **Scan**: discover drift using artifact scanning, scenario parsing, reference checking, and git history
- **Act**: for each flagged spec, read implementation code and determine action

## Invariants

- Scan all specs: parse scenarios, check staleness (last_updated vs git activity on related code), detect naming mismatches (filename vs feature field), find orphaned specs (zero incoming references). When semantic search is available, also surface content-level near-duplicate specs that differ only in filename (use a high relevance threshold, ~0.7). When unavailable, the structural pass alone is sufficient.
- Read actual implementation code for every flagged spec — never judge drift from metadata alone
- For each flagged spec, propose exactly one action: **keep** (still accurate), **rewrite** (update scenarios to match code), **rename** (filename ≠ feature field), **merge** (overlapping specs → combine into one), or **archive** (feature removed)
- Present all proposals as a summary table before executing anything — get explicit user approval
- **Rewrite**: preserve domain, feature field, and constraints; only rewrite scenarios to match current code behavior; show old vs new scenarios for review
- **Rename**: rename file to match feature field, update all references across .rpi/ artifacts
- **Merge**: combine scenarios into a single spec with unified feature name, archive source specs, update all references
- **Archive**: move to archive using the standard archive flow, update references
- After all approved actions execute, summarize what changed

## Principles

- Code is truth — when spec and code disagree, the spec is wrong
- Flag generously, act conservatively — surface all drift signals but require human judgment for every action
- Batch over piecemeal — present all findings at once rather than spec-by-spec to let the user see the full picture
