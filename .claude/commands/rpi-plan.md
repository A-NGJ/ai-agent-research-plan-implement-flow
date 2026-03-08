---
description: Create implementation plans — works standalone for simple tasks or with prior design docs for complex ones
model: opus
---

# Implementation Plan

Create implementation plans with phased tasks, success criteria, and verification steps.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or use `claude-init` to set it up.

**Two modes — auto-detected from input:**

- **Standalone mode**: For simple/short tasks. You describe what needs to be done, the plan does its own lightweight research and produces a plan directly. No prior `/rpi-research` or `/rpi-design` needed.
- **Pipeline mode**: For complex tasks with existing docs. You provide a design document, structure document, or a **ticket from `/rpi-tickets`** that links back to the pipeline (research → design → [structure] → tickets → plan → implement).

## Initial Response

When this command is invoked:

1. **Check what was provided:**
   - If a path to a design document (or structure document) was provided → **Pipeline mode**
   - If a path to a ticket file was provided → read its frontmatter:
     - If the ticket has a `design:` field (it came from `/rpi-tickets`) → **Pipeline mode** (ticket-originated)
     - If the ticket has no `design:` field (standalone ticket) → **Standalone mode**
   - If a plain task description was provided → **Standalone mode**
   - If nothing was provided, respond:
   ```
   I'll help you create an implementation plan.

   You can use this in two ways:

   **Simple task** (standalone):
   `/rpi-plan Add a retry mechanism to the webhook handler`

   **From the pipeline** (with prior docs):
   `/rpi-plan .thoughts/designs/2025-01-08-feature-name.md`
   `/rpi-plan .thoughts/tickets/auth-001-user-signup.md`
   ```

---

## Standalone Mode

For tasks that don't need a full research -> design -> structure pipeline. Typically: bug fixes, small features, refactors, config changes, adding tests, etc.

### Step 1: Understand the Task

1. **Read project conventions** — check for `CLAUDE.md` in the project root and note the actual commands for running tests, type checking, and linting. These will be used in success criteria instead of generic placeholders.
2. **Read any provided files fully** (tickets, referenced files, etc.)
3. **Research proportional to complexity** — scale effort to what the task actually needs:
   - **Obvious** (specific file/function named, single localized change): read those files directly, no sub-tasks needed
   - **Moderate** (area known, pattern unclear): spawn 1-2 targeted sub-tasks as needed
   - **Cross-cutting** (multiple systems, unclear file landscape): spawn all three in parallel:
     - Sub-task: "Load the `locate-codebase` skill, then find files related to [task]"
     - Sub-task: "Load the `find-patterns` skill, then find how similar things are done in the codebase for [task]"
     - Sub-task (@codebase-analyzer): Understand the specific code that needs to change
4. **Read the key files** identified by research
5. **If the task is ambiguous or you have questions**, present findings before proceeding:
   ```
   Here's what I found:

   Relevant files:
   - `path/to/file.ext:line` — [what it does, what needs to change]
   - `path/to/other.ext:line` — [what it does, what needs to change]

   Existing patterns:
   - [How similar things are handled in the codebase]

   My approach:
   - [1-3 sentences on what the plan will do]

   Questions before I write the plan:
   - [specific ambiguity to resolve]
   ```
   If everything is clear with no open questions, skip this step and write the plan directly.

### Step 2: Write the Plan

After understanding is confirmed (or immediately if the task is unambiguous):

1. **Break the work into phases** (often just 1-2 for simple tasks)
2. **Create the plan file**: Run `rpi scaffold plan --topic "..." --write`
   This creates `.thoughts/plans/YYYY-MM-DD-description.md` with frontmatter pre-populated.
3. **Fill in the plan content**: phases, tasks, code snippets, success criteria, commit steps. Each phase should include:
   - Overview of what the phase accomplishes
   - Tasks with file paths and change descriptions (include key code snippets)
   - Tests in the same phase as the code they test
   - Success criteria split into automated verification (use actual commands from `CLAUDE.md`) and manual verification
   - Commit step (stage list + message)
   - "Pause for manual confirmation" note between phases

### Step 3: Review & Iterate

Present the plan with a brief summary:
```
Plan saved: `.thoughts/plans/YYYY-MM-DD-description.md`

Phases:
- Phase 1: [name] — [what it does] ([scope])

Anything you'd like to adjust?
```

When revising based on feedback:
- **Scope change** (add/remove tasks): update the Scope line and affected phase tasks
- **Approach change**: re-research if needed — say so rather than guessing
- Keep iterating until the user confirms or stops giving feedback

---

## Pipeline Mode

For complex tasks that already went through the pipeline. Triggered by design docs, structure docs, or **ticket files from `/rpi-tickets`** that link back to design docs.

### Step 1: Read Inputs & Validate

1. **Read project conventions** — check for `CLAUDE.md` in the project root and note the actual commands for running tests, type checking, and linting. These will be used in success criteria instead of generic placeholders.

2. **Resolve the input document chain.**

   Run: `rpi chain <input-path>`

   This recursively follows frontmatter links (ticket → design → research, or design → research) and returns the full artifact chain with metadata. Read all the files it identifies.

   **If given a ticket file** — the ticket's **Scope**, **Design Context**, and **Acceptance Criteria** sections define the boundaries for this plan. The design doc provides broader context, but the ticket scopes what this specific plan covers.

   **If given a design or structure doc** — also check for related tickets:
   Run: `rpi scan --type ticket --design <path>`

   **IMPORTANT**: Read entire files — no limit/offset
   **CRITICAL**: Read these yourself before spawning sub-tasks

3. **Extract the ticket ID for naming** — if the input is a ticket file, or if any source document references a ticket, note the `ticket:` field value (e.g., `auth-001`). This will be used in the plan filename.

4. **Spot-check critical files from the design doc** (or structure doc if available):
   - Read 3-5 of the most important files mentioned
   - Verify the codebase still matches what the docs describe
   - If anything has drifted significantly, flag it immediately

5. **Present validation results**:
   ```
   I've read the pipeline docs:
   - Ticket: [path] — [ticket ID]: [title] (if ticket-originated)
   - Research: [path] — [topic summary]
   - Design: [path] — [key decisions: A, B, C]
   - Structure: [path] — [scope summary] (if available)

   Validation against current codebase:
   - [file:line] — confirmed, matches docs
   - [file:line] — DRIFT DETECTED: [explanation]

   Questions before I define phases:
   - [Phasing/ordering questions only — design decisions are already made]
   ```

### Step 2: Phase Definition

1. **Create a planning todo list** using TodoWrite
2. **Break the design's changes into ordered phases:**
   - Group related changes that must ship together
   - Respect dependency order (data model -> business logic -> API -> UI)
   - Each phase should leave the codebase in a working, testable state
   - Include tests in the same phase as the code they test — do not put tests in a separate phase or bottom section
   - If a structure doc is available, use its file listings; otherwise, identify files from the design doc's architecture/integration sections
3. **Present proposed phases for buy-in:**
   ```
   Proposed phases:

   ## Phase 1: [Name] — [what it accomplishes]
   Files: [list from design/structure doc]
   Depends on: nothing (foundation)

   ## Phase 2: [Name] — [what it accomplishes]
   Files: [list from design/structure doc]
   Depends on: Phase 1

   Does this phasing make sense?
   ```

### Step 3: Success Criteria & Verification

After phase buy-in:

1. **Define success criteria for each phase:**
   - **Automated Verification**: Use actual commands from `CLAUDE.md` (tests, linting, type checking, build)
   - **Manual Verification**: Human testing steps (UI, edge cases, integration)
2. **Define commit strategy per phase**
3. **Present for review**

### Step 4: Write the Plan

**Create the plan file**:
- Without ticket ID: `rpi scaffold plan --topic "..." --write`
- With ticket ID: `rpi scaffold plan --ticket <id> --design <path> --topic "..." --write`

This creates the file at `.thoughts/plans/YYYY-MM-DD-[ticket-id-]description.md` with frontmatter pre-populated.

**Fill in the plan content**: source documents section, phases (tasks, code snippets, success criteria, commit steps), migration notes if applicable, and references. Each phase should include:
- Overview (what it accomplishes and why it comes first / dependencies on prior phases)
- Tasks with file paths and change descriptions (include key code snippets, reference design doc for interface details)
- Tests in the same phase as the code they test
- Success criteria split into automated and manual verification
- Commit step (stage list + message)
- "Pause for manual confirmation" note between phases

### Step 5: Review & Iterate

Present the plan with a brief summary:
```
Plan saved: `.thoughts/plans/YYYY-MM-DD-<ticket-id>-description.md`

Phases:
- Phase 1: [name] — [what it does] ([N files])
- Phase 2: [name] — [what it does] ([N files])

Anything you'd like to adjust?
```

When revising based on feedback:
- **Phase reordering**: update dependencies and the "why it comes first" rationale
- **Scope change** (add/remove tasks): update the Scope line and affected phase tasks
- **Approach change**: re-research if needed — say so rather than guessing
- Keep iterating until the user confirms or stops giving feedback

---

## Guidelines (Both Modes)

1. **Be Interactive**: Get buy-in on approach/phases before writing the full plan
2. **Be Practical**: Focus on incremental, testable changes that keep the codebase working
3. **Separate Verification**: Always split success criteria into automated and manual
4. **No Open Questions**: Resolve ambiguity before finalizing — ask the user if needed
5. **Right-size the Plan**: Simple tasks get simple plans (1 phase, minimal ceremony). Complex tasks get detailed phasing with full verification
6. **Commit After Each Phase**: Every phase ends with a commit step — stage only that phase's files, not everything at once
7. **Tests Belong to Their Phase**: Write tests alongside the code they cover, not in a separate section at the bottom

### Pipeline Mode Only
8. **Trust Prior Stages**: Don't redo research or design work — reference those docs
9. **Spot-check Reality**: Verify the codebase still matches the design doc before planning
