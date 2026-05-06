---
name: rpi-plan
description: Plan a concrete, scoped change to existing behavior — produce a phased implementation plan. Use when user says 'remove the X flag', 'update Y to do Z', 'make sure X happens', or 'improve A to do B', even if they don't say 'plan'. Do NOT invoke when the change requires weighing design tradeoffs (use rpi-propose).
---

# Implementation Plan

## Goal

Create phased implementation plans with tasks, success criteria, and verification steps. This is part of: research → propose → **plan** → implement.

Auto-detect the mode from input:
- **Standalone** (plain task description) → lightweight research, then plan directly (bug fixes, small features, refactors)
- **Pipeline** (path to a design document) → plan built from prior pipeline work
- **Nothing provided** → ask for input with brief examples of each mode

Any mode can be combined with **`--grill`** — pass `--grill` or use phrasing like "grill me on this" / "stress-test this" to invoke the bundled `grill-me` skill at the approval gate (see invariants).

Any mode can be combined with **`--ff`** — pass `--ff` to suppress approval gates and auto-chain through `/rpi-implement` and `/rpi-verify`. Mutually exclusive with `--grill`.

When the user confirms the plan, suggest → `/rpi-implement <plan-path>`.

## Invariants

- Check the project's conventions for test/lint/build commands — use these in success criteria, not generic placeholders
- Read all provided files fully; research proportional to complexity
- Before drafting, search for prior plans and designs on this topic — prefer semantic search when available (default relevance threshold ~0.4), and fall back to keyword-based artifact discovery when not. Read snippets first; for high-relevance hits (score ≥ 0.7), expand the artifact chain to see lineage. If a prior plan covers the same scope, ask the user whether to extend it instead of opening a new one. Most valuable in standalone mode, where no chain is pre-resolved. If semantic search reports an installed-but-failing state, surface its hint before falling back.
- Check `.rpi/specs/` for specs covering the affected area — the plan must satisfy these behavioral contracts
- **Pipeline mode**: validate the design's status, resolve its full artifact chain, read all linked files, spot-check key files against current codebase for drift
- Break changes into ordered phases — each leaves the codebase working and testable
- Include tests in the same phase as the code they test
- Each phase has: tasks with file paths, success criteria (automated + manual), and a commit step
- When drafting each phase's Stage list, exclude paths matching `.gitignore` rules — gitignored artifacts (commonly the plan file itself, plus other `.rpi/` subdirectories under the default tracked-specs policy) must not appear in commit instructions
- Map phases to spec scenarios where applicable
- Get buy-in on proposed phases before writing the full plan
- If the user passed `--ff`, skip the phase outline buy-in — write the full plan immediately, run the existing automated coverage check, transition the design to complete, and invoke `/rpi-implement --ff <plan-path>` via the Skill tool. Error if `--grill` was also passed.
- If the user requested grilling (via `--grill` or natural-language phrasing) and `grill-me` is in your available skills, invoke `grill-me` on the phase outline before writing the full plan. Apply revisions inline, then continue with normal phase approval. If `grill-me` is unavailable, tell the user `grill-me` is not currently available and ask whether to proceed with the standard approval gate.
- **Pipeline mode**: after writing, verify the plan covers all design decisions — nothing silently dropped; transition design → complete
- Scaffold and save the plan artifact, linking to upstream design and spec

## Principles

- Right-size the plan — simple tasks get 1 phase with minimal ceremony; complex tasks get detailed phasing
- Be practical — incremental, testable changes that keep the codebase working
- Trust prior stages (pipeline) — don't redo research or design work
- Grilling is opt-in and single-pass — re-invoke if a second round is needed
