---
date: 2026-04-29T00:00:00+02:00
research: .rpi/research/2026-04-29-rpi-implement-worktree-mode-fragility.md
spec: .rpi/specs/custom-agents.md
status: complete
tags:
    - plan
    - rpi-implement
    - worktree
topic: remove worktree mode from rpi-implement skill
---

# Remove Worktree Mode from rpi-implement — Implementation Plan

## Overview

The `rpi-implement` skill currently mandates worktree isolation by default,
but the dedicated `rpi-implement-worktree` agent the original design
depended on was never built. The result is that worktree mode runs through a
generic agent and frequently fails on merge-back / cleanup with "worktree
locked" or "uncommitted changes" errors. Remove worktree mode from the skill
prompt and trim the corresponding scenarios from the custom-agents spec.
Implementation reverts to the in-place flow already described by the rest of
the skill's invariants.

**Scope**: 2 prose files modified. No Go code changes. No agent definition
changes (the `rpi-verify` agent does not use worktree mode and stays as-is).

## Source Documents

- **Research**: .rpi/research/2026-04-29-rpi-implement-worktree-mode-fragility.md

## Spec Scenario Map

| Action  | Scenario                                                  | Reason                                       |
| ------- | --------------------------------------------------------- | -------------------------------------------- |
| Remove  | Implement skill uses worktree isolation                   | Behavior being removed                       |
| Remove  | Implement skill falls back to in-place mode on failure    | No worktree mode → no fallback needed        |
| Remove  | Auto-merge on verification pass                           | No worktree branch to merge                  |
| Remove  | Manual verification pauses merge                          | No worktree branch to merge                  |
| Remove  | Base branch stays clean during implementation             | Was a worktree-mode invariant                |
| Keep    | Init installs agent definitions for Claude target         | Still applies to `rpi-verify`                |
| Keep    | Init skips agent definitions for non-Claude targets       | Still applies to `rpi-verify`                |
| Keep    | Update syncs agent definitions                            | Still applies to `rpi-verify`                |
| Keep    | Verification agent restricted to read-only operations     | Unchanged                                    |
| Keep    | Verification agent returns structured results             | Unchanged                                    |

---

## Phase 1: Remove worktree mode from skill and spec

### Overview

Strip the `## Worktree Mode` section from the source `rpi-implement` skill,
remove worktree-only scenarios and constraints from the `custom-agents`
spec, and verify all prompt-structure tests still pass.

### Tasks

#### 1. Remove Worktree Mode section from rpi-implement skill

**File**: `internal/workflow/assets/skills/rpi-implement/SKILL.md`

**Changes**: Delete the entire `## Worktree Mode` block (currently
SKILL.md:33–49, including the blank line that precedes it). The skill ends
at `## Principles`, matching the structure of other pipeline skills.
Frontmatter, `## Goal`, `## Invariants`, and `## Principles` are unchanged.

#### 2. Trim worktree scenarios from custom-agents spec

**File**: `.rpi/specs/custom-agents.md`

**Changes**:

- Remove the five worktree-only scenarios listed in the Spec Scenario Map
  above (lines 41–64 of the current file: "Implement skill uses worktree
  isolation" through "Base branch stays clean during implementation").
- Trim the worktree-specific constraint lines:
  - `Merge uses regular merge (not squash) to preserve per-phase commit history`
  - `After successful merge, the worktree and its branch are cleaned up`
- Update frontmatter:
  - `last_updated` → `2026-04-29T00:00:00+02:00`
  - `updated_by` → `.rpi/plans/2026-04-29-remove-worktree-mode-from-rpi-implement.md`

The "Purpose", "Constraints", and "Out of Scope" sections stay; only the
worktree-specific lines are removed.

### Success Criteria

#### Automated Verification

- [x] `grep -rn "worktree\|isolation" internal/workflow/assets/skills/ internal/workflow/assets/agents/` returns no matches
- [x] `go test ./internal/workflow/...` passes (covers `TestPromptStructure_HasRequiredSections`, `TestPromptStructure_LineCount`, `TestPromptStructure_NoToolReferences`)
- [x] `make test` passes (full Go suite)
- [x] `make build` succeeds

### Commit

- [x] Stage: `internal/workflow/assets/skills/rpi-implement/SKILL.md`, `.rpi/specs/custom-agents.md`
- [x] Message: `refactor(implement): remove worktree mode from rpi-implement skill`

---

## Out of Scope

- Renaming `custom-agents.md` to `verify-agent.md`. The spec still describes
  a custom agent (the read-only verify agent), so the name remains accurate.
  Can be revisited if more agents land later.
- Editing the archived design `.rpi/designs/2026-04-07-custom-agents-for-verification-and-worktree-implementation.md`.
  Archived artifacts are historical record; the research note already
  documents the reversal.
- Changes to the `rpi-verify` skill or agent (do not use worktree mode).
- Building the originally-planned `rpi-implement-worktree` agent.

## References

- Research: .rpi/research/2026-04-29-rpi-implement-worktree-mode-fragility.md
- Spec being trimmed: .rpi/specs/custom-agents.md
- Current skill (worktree mode lives at lines 33–49): `internal/workflow/assets/skills/rpi-implement/SKILL.md`
- Originating commit (introduced mandatory worktree mode): `2bfc597 feat(agents): add worktree mode instructions to implement skill`
