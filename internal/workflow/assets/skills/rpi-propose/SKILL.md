---
name: rpi-propose
description: Investigate, analyze, and propose solutions — from quick decisions to complex features
---

# Solution Design

## Goal

Create a design document (`.rpi/designs/`) and behavioral spec (`.rpi/specs/`) for a given topic or research. This is part of: research → **propose** → plan → implement.

Auto-detect the mode from input:
- **Focused decision** (e.g., "should we use X or Y?") → quick investigation, trade-off analysis, write design+spec
- **Complex feature** (description or path to research doc) → thorough investigation, explore options, synthesize, write design+spec
- **Updating existing design** (path to existing design) → read it, understand what changed, propose updates in place
- **Nothing provided** → ask for input with brief examples of each mode

Any mode can be combined with **`--grill`** — pass `--grill` or use phrasing like "grill me on this" / "stress-test this" to invoke the externally-installed `grill-me` skill at the approval gate (see invariants).

When the user approves the spec, suggest → `/rpi-plan <design-path>`.

## Invariants

- Check for existing designs on this topic before creating a new one
- If a research doc is provided, read it and resolve its full artifact chain — warn if still draft or already complete
- Investigate the codebase before proposing — ground decisions in evidence with file:line refs
- Get user buy-in on trade-offs before writing the design
- Link the design to upstream research via frontmatter
- Create a behavioral spec with 5-8 Given/When/Then scenarios describing user-observable behavior — scenarios must not reference internal structure (structs, file paths, function names); include a Constraints section for boundaries and an Out of Scope section. Name the spec file after its `feature` field (e.g., `feature: rpi-status` → `rpi-status.md`)
- Present the spec for approval — iterate until accepted
- If the user requested grilling (via `--grill` or natural-language phrasing) and `grill-me` is in your available skills, invoke `grill-me` on the drafted design+spec before the approval gate. Apply revisions inline to the design and spec, then present for approval. If `grill-me` is unavailable, tell the user it must be installed externally and ask whether to proceed with the standard approval gate.
- Transition artifacts: design → active, research → complete (if fully addressed)
- For incremental mode: update in place, add an Update Log entry, update affected specs

## Principles

- Be opinionated — recommend with clear reasoning grounded in codebase evidence
- Be interactive — checkpoint before major decisions; a design that surprises during review means the process failed
- Scale effort to complexity — a focused decision needs less investigation than a new subsystem
- Specs are the contract — every design culminates in a spec
- Grilling is opt-in and single-pass — re-invoke if a second round is needed
