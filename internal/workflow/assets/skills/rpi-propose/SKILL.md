---
name: rpi-propose
description: Design a new feature or non-trivial change with tradeoffs — produce a design and behavioral spec. Use when user says 'add feature X', 'introduce mode Y', 'support Z', or proposes new functionality, even if they don't say 'propose'. Do NOT invoke for narrow tweaks (use rpi-plan) or open-ended exploration (use rpi-research).
---

# Solution Design

## Goal

Create a design document (`.rpi/designs/`) and behavioral spec (`.rpi/specs/`) for a given topic or research. This is part of: research → **propose** → plan → implement.

Auto-detect the mode from input:
- **Focused decision** (e.g., "should we use X or Y?") → quick investigation, trade-off analysis, write design+spec
- **Complex feature** (description or path to research doc) → thorough investigation, explore options, synthesize, write design+spec
- **Updating existing design** (path to existing design) → read it, understand what changed, propose updates in place
- **Nothing provided** → ask for input with brief examples of each mode

Any mode can be combined with **`--grill`** — pass `--grill` or use phrasing like "grill me on this" / "stress-test this" to invoke the bundled `grill-me` skill at the approval gate (see invariants).

Any mode can be combined with **`--ff`** — pass `--ff` to suppress approval gates and auto-chain through `/rpi-plan`, `/rpi-implement`, and `/rpi-verify`. Mutually exclusive with `--grill`.

When the user approves the spec, suggest → `/rpi-plan <design-path>`.

## Invariants

- Before drafting, search for prior designs (and optionally specs) on this topic — prefer semantic search when available (default relevance threshold ~0.4), and fall back to keyword-based artifact discovery when not. Read snippets first. For high-relevance hits (score ≥ 0.7), expand the artifact chain to see lineage. Decide whether the new design supersedes or extends prior work; if extending, link via frontmatter. If semantic search reports an installed-but-failing state, run the recovery command from its hint and retry once; only fall back to keyword discovery (and surface the hint) if the retry still fails.
- If a research doc is provided, read it and resolve its full artifact chain — warn if still draft or already complete
- Investigate the codebase before proposing — ground decisions in evidence with file:line refs
- Get user buy-in on trade-offs before writing the design
- Link the design to upstream research via frontmatter
- Create a behavioral spec with 5-8 Given/When/Then scenarios describing user-observable behavior — scenarios must not reference internal structure (structs, file paths, function names); include a Constraints section for boundaries and an Out of Scope section. Name the spec file after its `feature` field (e.g., `feature: rpi-status` → `rpi-status.md`)
- Present the spec for approval — iterate until accepted
- If the user passed `--ff`, skip the trade-off buy-in, the mid-flight decision checkpoints, and the spec approval gate — auto-accept the drafted design and spec and immediately invoke `/rpi-plan --ff <design-path>` via the Skill tool. Error if `--grill` was also passed.
- If the user requested grilling (via `--grill` or natural-language phrasing) and `grill-me` is in your available skills, invoke `grill-me` on the drafted design+spec before the approval gate. Apply revisions inline to the design and spec, then present for approval. If `grill-me` is unavailable, tell the user `grill-me` is not currently available and ask whether to proceed with the standard approval gate.
- Transition artifacts: design → active, research → complete (if fully addressed)
- For incremental mode: update in place, add an Update Log entry, update affected specs

## Principles

- Be opinionated — recommend with clear reasoning grounded in codebase evidence
- Be interactive — checkpoint before major decisions; a design that surprises during review means the process failed
- Scale effort to complexity — a focused decision needs less investigation than a new subsystem
- Specs are the contract — every design culminates in a spec
- Grilling is opt-in and single-pass — re-invoke if a second round is needed
