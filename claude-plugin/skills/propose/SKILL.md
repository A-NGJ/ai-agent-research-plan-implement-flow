---
name: propose
description: Design a new feature or non-trivial change with tradeoffs — produce a design and behavioral spec. Use when user says 'add feature X', 'introduce mode Y', 'support Z', or proposes new functionality, even if they don't say 'propose'. Do NOT invoke for narrow tweaks (use rpi-plan) or open-ended exploration (use rpi-research).
---

# Solution Design

## Goal

Create a design document (`.rpi/designs/`) and behavioral spec (`.rpi/specs/`) for a given topic or research.

Auto-detect the mode from input:
- **Focused decision** (e.g., "should we use X or Y?") → quick investigation, trade-off analysis, write design+spec
- **Complex feature** (description or path to research doc) → thorough investigation, explore options, synthesize, write design+spec
- **Updating existing design** (path to existing design) → read it, understand what changed, propose updates in place
- **Nothing provided** → ask for input with brief examples of each mode

When the user approves the spec, suggest → `/rpi:plan <design-path>`.

## Invariants

- Before drafting, search for prior designs and specs on this topic; decide whether to supersede or extend
- See the project's RPI Skill Contract for `--ff` / `--grill` semantics; both flags apply here and are mutually exclusive
- If a research doc is provided, read it and resolve its full artifact chain — warn if still draft or already complete
- Investigate the codebase before proposing — ground decisions in evidence with file:line refs
- Get user buy-in on trade-offs before writing the design
- Link the design to upstream research via frontmatter
- Create a behavioral spec with 5-8 Given/When/Then scenarios describing user-observable behavior — scenarios must not reference internal structure (structs, file paths, function names); include a Constraints section for boundaries and an Out of Scope section. Name the spec file after its `feature` field (e.g., `feature: rpi-status` → `rpi-status.md`)
- Present the spec for approval — iterate until accepted
- Under `--ff`, skip the trade-off buy-in, the mid-flight decision checkpoints, and the spec approval gate — auto-accept the drafted design and spec and immediately invoke `/rpi:plan --ff <design-path>` via the Skill tool
- Under `--grill` (or matching natural-language phrasing) and when `grill-me` is available, invoke `grill-me` on the drafted design+spec before the approval gate; apply revisions inline. If `grill-me` is unavailable, tell the user and ask whether to proceed with the standard approval gate.
- Transition artifacts: design → active, research → complete (if fully addressed)
- For incremental mode: update in place, add an Update Log entry, update affected specs

## Principles

- Be opinionated — recommend with clear reasoning grounded in codebase evidence
- Be interactive — checkpoint before major decisions; a design that surprises during review means the process failed
- Scale effort to complexity — a focused decision needs less investigation than a new subsystem
- Specs are the contract — every design culminates in a spec
- Grilling is opt-in and single-pass — re-invoke if a second round is needed
