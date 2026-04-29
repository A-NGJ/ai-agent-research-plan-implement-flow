---
name: rpi-plan
description: Create implementation plans — works standalone for simple tasks or with prior designs for complex ones
---

# Implementation Plan

## Goal

Create phased implementation plans with tasks, success criteria, and verification steps. This is part of: research → propose → **plan** → implement.

Auto-detect the mode from input:
- **Standalone** (plain task description) → lightweight research, then plan directly (bug fixes, small features, refactors)
- **Pipeline** (path to a design document) → plan built from prior pipeline work
- **Nothing provided** → ask for input with brief examples of each mode

Any mode can be combined with **`--grill`** — pass `--grill` or use phrasing like "grill me on this" / "stress-test this" to invoke the externally-installed `grill-me` skill at the approval gate (see invariants).

When the user confirms the plan, suggest → `/rpi-implement <plan-path>`.

## Invariants

- Check the project's conventions for test/lint/build commands — use these in success criteria, not generic placeholders
- Read all provided files fully; research proportional to complexity
- Check `.rpi/specs/` for specs covering the affected area — the plan must satisfy these behavioral contracts
- **Pipeline mode**: validate the design's status, resolve its full artifact chain, read all linked files, spot-check key files against current codebase for drift
- Break changes into ordered phases — each leaves the codebase working and testable
- Include tests in the same phase as the code they test
- Each phase has: tasks with file paths, success criteria (automated + manual), and a commit step
- Map phases to spec scenarios where applicable
- Get buy-in on proposed phases before writing the full plan
- If the user requested grilling (via `--grill` or natural-language phrasing) and `grill-me` is in your available skills, invoke `grill-me` on the phase outline before writing the full plan. Apply revisions inline, then continue with normal phase approval. If `grill-me` is unavailable, tell the user it must be installed externally and ask whether to proceed with the standard approval gate.
- **Pipeline mode**: after writing, verify the plan covers all design decisions — nothing silently dropped; transition design → complete
- Scaffold and save the plan artifact, linking to upstream design and spec

## Principles

- Right-size the plan — simple tasks get 1 phase with minimal ceremony; complex tasks get detailed phasing
- Be practical — incremental, testable changes that keep the codebase working
- Trust prior stages (pipeline) — don't redo research or design work
- Grilling is opt-in and single-pass — re-invoke if a second round is needed
