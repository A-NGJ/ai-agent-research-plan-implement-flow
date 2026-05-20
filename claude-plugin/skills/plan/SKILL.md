---
name: plan
description: Plan a concrete, scoped change to existing behavior — produce a phased implementation plan. Use when user says 'remove the X flag', 'update Y to do Z', 'make sure X happens', or 'improve A to do B', even if they don't say 'plan'. Do NOT invoke when the change requires weighing design tradeoffs (use rpi-propose).
---

# Implementation Plan

## Goal

Create phased implementation plans with tasks, success criteria, and verification steps.

Auto-detect the mode from input:
- **Standalone** (plain task description) → lightweight research, then plan directly (bug fixes, small features, refactors)
- **Pipeline** (path to a design document) → plan built from prior pipeline work
- **Nothing provided** → ask for input with brief examples of each mode

When the user confirms the plan, suggest → `/rpi:implement <plan-path>` (or the first sibling plan with no dependencies, if the design was split).

## Invariants

- Check the project's conventions for test/lint/build commands — use these in success criteria, not generic placeholders
- Read all provided files fully; research proportional to complexity
- Before drafting, search for prior plans and designs on this topic; if a prior plan covers the same scope, ask whether to extend it
- See the project's RPI Skill Contract for `--ff` / `--grill` semantics; both flags apply here and are mutually exclusive
- Check `.rpi/specs/` for specs covering the affected area — the plan must satisfy these behavioral contracts
- **Pipeline mode**: validate the design's status, resolve its full artifact chain, read all linked files, spot-check key files against current codebase for drift
- **Pipeline-mode split**: call the split-score tool on the design path to get the complexity score and per-signal breakdown. If `should_split` is true (score ≥ threshold), propose a labeled breakdown of sibling plans before writing any plan: each plan covers ≥1 design component (use the returned `components.headings` to cluster), the union covers every component, slugs follow `<design-topic>-<scope-slug>`, and dependencies form a DAG. Run accept / edit / single-plan with the user; under `--ff`, accept the auto-generated proposal. On cycle detection, surface the participating plans and re-enter the edit loop without writing files. Splits are pipeline-only.
- **Resume detection**: in pipeline mode, scan for existing plans linked to the same design before proposing a split or writing a single plan. If any exist, surface them with their statuses and ask whether to resume the split, treat it as complete, or start a new split. Under `--ff`, default to resume if a `<!-- rpi-plan-split: pending [...] -->` marker exists in the design body, else treat as complete.
- Break changes into ordered phases — each leaves the codebase working and testable
- Include tests in the same phase as the code they test
- Each phase has: tasks with file paths, success criteria (automated + manual), and a commit step
- When drafting each phase's Stage list, exclude paths matching `.gitignore` rules — gitignored artifacts (commonly the plan file itself, plus other `.rpi/` subdirectories under the default tracked-specs policy) must not appear in commit instructions
- Map phases to spec scenarios where applicable
- Get buy-in on proposed phases before writing the full plan
- **Per sibling plan in a split**: scaffold each plan with its `depends_on:` list and `Sibling plans` body block, write them in topological order, run the per-plan approval gate (or skip under `--ff`), and update the resume marker on the design until the last sibling is written
- Under `--ff`, skip approval gates — write the full plan(s) immediately, run the existing automated coverage check, transition the design to complete only after every proposed sibling exists on disk, then invoke `/rpi:implement --ff <plan-path>` via the Skill tool for each plan in topological order (waiting for each to finish before starting the next), then `/rpi:verify <last-plan-path>` once at the end
- Under `--grill` (or matching natural-language phrasing) and when `grill-me` is available, invoke `grill-me` on the split proposal (when offered) or on the phase outline (single-plan case) before writing the full plan; apply revisions inline. If `grill-me` is unavailable, tell the user and ask whether to proceed with the standard approval gate.
- **Pipeline mode**: after writing, verify the plan(s) cover all design decisions — nothing silently dropped. Transition design → complete only when every proposed sibling plan has been written; otherwise leave it active and record remaining plans in the resume marker.
- Scaffold and save the plan artifact, linking to upstream design and spec

## Principles

- Right-size the plan — simple tasks get 1 phase; complex multi-component designs get split into sibling plans, each independently shippable
- Be practical — incremental, testable changes that keep the codebase working
- Trust prior stages (pipeline) — don't redo research or design work
- Grilling is opt-in and single-pass — re-invoke if a second round is needed
- Splits are user-confirmed (or `--ff`-accepted) — never silently break a design into siblings without surfacing the breakdown first
