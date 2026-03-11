# How Each Stage Works

## Explore (`/rpi-explore`)

**Purpose:** Investigate the codebase -- conversational fact-finding with optional research artifact.

Works for both focused questions ("how does the auth pipeline work?") and open-ended exploration ("what could we improve about error handling?"). The command spawns parallel sub-agents that use specialized skills:
- **locate-codebase** -- Finds where files and components live
- **codebase-analyzer** -- Understands how specific code works (traces data flow, documents patterns)
- **find-patterns** -- Finds examples of existing patterns to model after
- **locate-thoughts** -- Discovers relevant historical documents in `.thoughts/`
- **analyze-thoughts** -- Extracts key insights from existing documents

Exploration is conversational by default -- no artifact is created unless you ask. For broad queries, Claude shows you what it found in an initial scan and asks if you want to redirect focus before deep-diving.

Findings are always factual with concrete file:line references. When findings reveal clear pain points, opportunities, or trade-offs, an Assessment section provides an opinionated take -- clearly separated from the facts. If you want to save findings, ask Claude to write them to `.thoughts/research/`. If the exploration comprehensively documents a module's behavior, it can optionally create or update a spec in `.thoughts/specs/`.

## Propose (`/rpi-propose`)

**Purpose:** Investigate, analyze, and propose solutions with trade-offs.

Three modes, auto-detected:
- **Quick** -- Focused decision between 2-3 options (single component, one pattern choice)
- **Full** -- Multi-decision feature design with component diagrams, risk tables, and file structure
- **Incremental** -- Update an existing proposal with new information

The propose stage is interactive. Claude investigates the codebase, presents options with concrete trade-offs (grounded in your actual codebase, not generic advice), makes a recommendation, and waits for your direction. After you choose, it validates that your combined choices work together before documenting the proposal in `.thoughts/proposals/`. If the proposal changes existing behavior documented in `.thoughts/specs/`, it can flag those specs with `pending_changes` for update after implementation.

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

## Implement (`/rpi-implement`)

**Purpose:** Execute a plan phase-by-phase with verification at each step.

The implementation stage:
1. Reads the plan completely
2. Checks for sensitive files (credentials, secrets) before proceeding
3. For each phase, shows a **pre-review** of all intended changes before writing code
4. Implements the phase after approval
5. Runs automated verification (tests, linting, type checks)
6. Updates checkboxes in the plan file
7. If the plan specifies spec updates, updates the relevant `.thoughts/specs/` files
8. Proposes a commit
9. Pauses for manual verification before the next phase

If the plan doesn't match reality (codebase drifted since the plan was written), it stops and clearly explains the mismatch rather than silently improvising.

Resumable: if you invoke `/rpi-implement` on a partially-completed plan, it picks up from the first unchecked item.

## Commit (`/rpi-commit`)

**Purpose:** Create clean, focused git commits without thinking about `git add` and message formatting.

`/rpi-commit` inspects your working tree (staged, unstaged, and untracked files), groups related changes into logical commits, drafts messages matching your repo's existing commit style, and presents the plan for your approval before executing. It never adds Claude attribution or co-author lines -- commits look like you wrote them.

You can use `/rpi-commit` standalone anytime, or let `/rpi-implement` handle commits at the end of each phase.

## Verify (`/rpi-verify`)

**Purpose:** Validate that an implementation matches its proposal artifacts.

Checks three dimensions:
- **Completeness** -- Are all planned changes implemented?
- **Correctness** -- Does the code match the proposal decisions?
- **Coherence** -- Do the pieces work together as intended?

Can auto-detect what to verify from recent git changes and active plans, or you can point it at a specific proposal, plan, or research doc. Produces a severity-classified report. Purely advisory -- it doesn't block anything, and can be re-run after fixes.

## Archive (`/rpi-archive`)

**Purpose:** Move completed artifacts to `.thoughts/archive/` to keep the active directory clean.

Two modes:
- **Specific paths** -- `/rpi-archive .thoughts/research/2026-01-15-auth-flow.md`
- **Scan mode** -- `/rpi-archive` with no arguments scans for completed artifacts

Warns before archiving anything still in `draft` or `active` status. Preserves the full directory structure inside `archive/` (e.g., `archive/research/`, `archive/proposals/`).
