# How Each Stage Works

## Research (`/rpi-research`)

**Purpose:** Investigate the question -- codebase exploration or external systems/libraries/frameworks, with conversational fact-finding and an optional research artifact.

Works for both focused questions ("how does the auth pipeline work?", "what's the state of vector databases in 2026?") and open-ended exploration ("what could we improve about error handling?", "what agentic frameworks exist for X?"). The command starts with a brief interview to understand motivation, prior attempts, and constraints, then investigates at a depth proportional to the question -- across the codebase, external sources, or both.

Research is conversational by default -- no artifact is created unless you ask or it's clearly valuable for cross-session handoff. For broad queries, Claude shows initial findings and asks if you want to redirect focus before deep-diving.

Findings carry source anchors: `file:line` references for codebase claims, URL or quoted documentation for external claims. Authoritative external sources (project README, official docs, release notes) are preferred; blog posts and forum threads are flagged as such. Facts are presented first; an opinionated assessment is offered only when warranted, clearly separated from the facts. If you want to save findings, ask Claude to write them to `.rpi/research/`.

## Propose (`/rpi-propose`)

**Purpose:** Investigate, analyze, and propose solutions with trade-offs.

Three modes, auto-detected:
- **Quick** -- Focused decision between 2-3 options (single component, one pattern choice)
- **Full** -- Multi-decision feature design with component diagrams, risk tables, and file structure
- **Incremental** -- Update an existing proposal with new information

The propose stage is interactive. Claude investigates the codebase, presents options with concrete trade-offs (grounded in your actual codebase, not generic advice), makes a recommendation, and waits for your direction. After you choose, it validates that your combined choices work together before documenting the design in `.rpi/designs/`. If the proposal changes existing behavior documented in `.rpi/specs/`, it can flag those specs with `pending_changes` for update after implementation.

**Grill mode (opt-in):** Pass `--grill` (or use phrasing like "grill me on this", "stress-test this") to invoke the bundled `grill-me` skill (sourced from [mattpocock/skills](https://github.com/mattpocock/skills) under MIT) at the approval gate. Once the design and spec are drafted, `grill-me` interrogates them one question at a time, walking every branch of the decision tree before you accept the artifact. Findings are applied inline to the design and spec; no separate audit log is written. Grilling is single-pass -- re-invoke if you want a second round.

If a user has removed the bundled `grill-me` skill, you'll be told and asked whether to proceed with the standard approval gate.

**Fast-forward mode (opt-in):** Pass `--ff` to suppress all approval gates (trade-off buy-in, mid-flight checkpoints, spec approval) and auto-chain into `/rpi-plan --ff <design-path>` immediately. The chain continues through `/rpi-implement` and ends at `/rpi-verify`, producing a verification report in `.rpi/reviews/` as the terminal artifact. Mutually exclusive with `--grill`; the explicit flag is required (no natural-language trigger). Use when you trust the defaults and want full autopilot. Safety gates (codebase drift detection, sensitive-content scan) still stop the chain.

## Plan (`/rpi-plan`)

**Purpose:** Create a phase-by-phase implementation plan with specific code changes and verification steps.

Works in two modes:
- **Standalone** -- For simple tasks. Does its own lightweight research and produces a plan directly.
- **Pipeline** -- For complex tasks with existing proposal documents. Reads the full document chain, spot-checks the codebase against the docs, and breaks work into verified phases. For large proposals, decomposes the work into independently plannable units within the plan.

Every plan phase includes:
- Specific file changes with code snippets
- Tests (in the same phase as the code they test, not a separate "testing phase")
- Automated verification commands (pulled from your project's `CLAUDE.md`)
- Manual verification steps
- A commit with specific files and message

All open questions must be resolved before the plan is finalized.

**Grill mode (opt-in):** Pass `--grill` (or use phrasing like "grill me on this") to invoke the bundled `grill-me` skill on the *phase outline* before the full plan is written. Same fall-back behavior as Propose -- if a user has removed `grill-me` locally, you'll be asked whether to proceed without it.

**Fast-forward mode (opt-in):** Pass `--ff` to skip the phase outline buy-in and auto-chain into `/rpi-implement --ff <plan-path>`, terminating at `/rpi-verify`. Pre-flight checks (design status, artifact chain, drift spot-check) and the design-coverage check still run and can stop the chain. Mutually exclusive with `--grill`.

## Implement (`/rpi-implement`)

**Purpose:** Execute a plan phase-by-phase with verification at each step.

The implementation stage works phase-by-phase:

1. Reads the plan completely
2. Checks for sensitive files (credentials, secrets) before proceeding
3. For each phase, presents intended changes before writing code
4. Implements the phase
5. Runs automated verification (tests, linting, type checks)
6. Auto-commits when checks pass -- no manual confirmation needed
7. Updates checkboxes in the plan file
8. Advances to the next phase automatically if all success criteria are covered by automated checks; pauses only when the plan includes manual verification items
9. On completion, verifies spec conformance for all linked specs -- extracts Given/When/Then scenarios and checks each against the implementation

If the plan doesn't match reality (codebase drifted since the plan was written), it stops and clearly explains the mismatch rather than silently improvising.

Resumable: if you invoke `/rpi-implement` on a partially-completed plan, it picks up from the first unchecked item.

**Fast-forward mode (opt-in):** Pass `--ff` to skip the per-phase pre-review and any manual verification pauses. After the plan completes, `/rpi-verify <plan-path>` is invoked automatically, producing a verification report in `.rpi/reviews/` as the chain's terminal artifact. The "On mismatch" gate, the sensitive-content scan, end-of-plan spec verification, and phase failure handling all run unchanged -- `--ff` is a review override, not a safety override. Mutually exclusive with `--grill`.

## Commit (`/rpi-commit`)

**Purpose:** Create clean, focused git commits without thinking about `git add` and message formatting.

`/rpi-commit` inspects your working tree (staged, unstaged, and untracked files), groups related changes into logical commits, drafts messages matching your repo's existing commit style, and presents the plan for your approval before executing. It never adds Claude attribution or co-author lines -- commits look like you wrote them.

You can use `/rpi-commit` standalone anytime, or let `/rpi-implement` handle commits at the end of each phase.

## Verify (`/rpi-verify`)

**Purpose:** Validate that an implementation matches its proposal artifacts.

Checks three dimensions:
- **Completeness** -- Are all planned changes implemented?
- **Correctness** -- Does the code match the proposal decisions? Extracts Given/When/Then scenarios from linked specs and verifies each against actual code and tests, reporting pass/fail per scenario with file:line references.
- **Coherence** -- Do the pieces work together as intended?

Can auto-detect what to verify from recent git changes and active plans, or you can point it at a specific proposal, plan, or research doc. Produces a severity-classified report. Purely advisory -- it doesn't block anything, and can be re-run after fixes.

## Diagnose (`/rpi-diagnose`)

**Purpose:** Iteratively diagnose and fix complex bugs through root-cause analysis.

Takes a bug description, error message, or path to a failing test. The workflow:
1. Interviews you to establish expected vs actual behavior
2. Checks for existing diagnoses on the same topic
3. Reproduces the bug before investigating
4. Traces the code path with file:line references
5. Attempts a fix (up to 3 attempts, reverting failed ones)
6. Auto-commits on success, escalates to `/rpi-plan` or `/rpi-propose` if the bug is too complex

Always produces a diagnosis artifact in `.rpi/diagnoses/` regardless of outcome -- useful as a post-mortem even if the fix requires a larger effort.

## Explain (`/rpi-explain`)

**Purpose:** Walk through an implemented solution with a diff-scoped explanation.

Two modes, auto-detected:
- **With artifact path** -- Resolves the full artifact chain and walks through the diff with design context. Explains *why* each change was made, citing the design or plan.
- **No arguments** -- Auto-detects changed files from git and explains without artifact context.

File-by-file walkthrough with file:line references. Straightforward changes are summarized briefly; non-obvious changes get detailed reasoning. No artifact is saved by default -- ask if you want one.

## Spec Sync (`/rpi-spec-sync`)

**Purpose:** Sync behavioral specs in `.rpi/specs/` to match the current codebase.

Works in two phases:
1. **Scan** -- Reads all specs and compares against the actual implementation. Detects drift, naming mismatches, orphaned specs, and stale scenarios.
2. **Act** -- Presents a summary table of proposed actions (keep, rewrite, rename, merge, archive) and waits for approval before executing.

When code and spec disagree, code is truth -- the spec gets updated, not the code. On rename or merge, all cross-references across `.rpi/` are updated automatically.

Use after a batch of changes to keep specs current, or periodically as maintenance.

## Archive (`/rpi-archive`)

**Purpose:** Move completed artifacts to `.rpi/archive/` to keep the active directory clean.

Two modes:
- **Specific paths** -- `/rpi-archive .rpi/research/2026-01-15-auth-flow.md`
- **Scan mode** -- `/rpi-archive` with no arguments scans for completed artifacts

Warns before archiving anything still in `draft` or `active` status. Preserves the full directory structure inside `archive/` (e.g., `archive/2026-04/research/`, `archive/2026-04/designs/`).

## Status (`rpi status`)

**Purpose:** Single-screen dashboard of all RPI artifacts.

Shows:
- Artifact counts grouped by type and status
- Active plan progress with checkbox completion percentages
- Stale artifacts (non-terminal status, past a configurable threshold -- default 14 days)
- Specs with scenario counts
- Archive-ready artifacts with reference counts

```bash
rpi status                  # text dashboard
rpi status --format json    # machine-readable output
rpi status --stale-days 7   # flag artifacts stale after 7 days
```

## Resume (`rpi resume`)

**Purpose:** Session-level overview to answer "where did I leave my work?" after returning to a project.

Shows:
- Active and draft artifacts across all types (path, type, status, topic)
- Current phase, checkbox progress, and the next unchecked items of the most recent active plan
- A suggested next pipeline action (e.g., "design has no implementation plan -- run /rpi-plan")

```bash
rpi resume                  # human-readable text summary (default)
rpi resume --format json    # structured JSON (same shape as the MCP tool)
```

Also exposed as the `rpi_session_resume` MCP tool. Claude Code's `rpi init` wires a `SessionStart` hook that prompts the assistant to call it automatically, so the AI orients itself on the in-flight work before you type. You can invoke it manually any time by running `rpi resume` or by asking the AI something like "where did I leave my work?".
