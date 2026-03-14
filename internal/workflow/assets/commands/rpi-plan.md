---
description: Create implementation plans — works standalone for simple tasks or with prior proposals for complex ones
model: opus
---

# Implementation Plan

Create implementation plans with phased tasks, success criteria, and verification steps.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.
Run `rpi --help` to discover available commands and `rpi <command> --help` for detailed usage with examples.

**Two modes — auto-detected from input:**

- **Standalone mode**: Plain task description → lightweight research, then plan directly
- **Pipeline mode**: Path to a proposal document → plan built from prior pipeline work
- **Nothing provided** → Ask for input with brief examples of each mode

---

## Standalone Mode

For tasks that don't need the full propose → plan pipeline: bug fixes, small features, refactors, config changes, adding tests.

### Step 1: Understand the task

1. Check the project's conventions and configuration for the actual commands for running tests, linting, and type checking — these will be used in success criteria instead of generic placeholders
2. Read any provided files fully
3. Research proportional to complexity:
   - **Obvious** (specific file/function named): read those files directly
   - **Moderate** (area known, pattern unclear): use `rpi index query "[task]"` to locate files, then read them
   - **Cross-cutting** (multiple systems): investigate in parallel — find relevant files with `rpi index query`, understand how similar things are done in the codebase, read the key implementation files
4. If the task is ambiguous or you have questions, present findings and open questions before writing the plan. If everything is clear, write the plan directly.

### Step 2: Write the plan

1. Break the work into phases (often just 1-2 for simple tasks)
2. Use `rpi` to scaffold and save a plan artifact for this topic
3. Fill in phases with: tasks and file paths, key code snippets, success criteria (automated using the project's actual test/lint commands + manual), and commit steps. Include tests in the same phase as the code they test.

### Step 3: Review & iterate

Present the plan summary and ask if anything needs adjusting. Keep iterating until the user confirms.

---

## Pipeline Mode

For complex tasks that already went through the pipeline. Triggered by proposal documents from `/rpi-propose`.

### Step 1: Read inputs & validate

1. Check the project's conventions for test/lint/build commands
2. Use `rpi` to check the proposal's status — warn if it's still in draft or already marked complete
3. Use `rpi` to resolve the full artifact chain from the proposal. Read all linked files fully.
4. Check `.thoughts/specs/` for specs covering modules affected by this proposal
5. Spot-check 3-5 key files from the proposal against the current codebase — flag any significant drift
6. Present validation results and any scoping questions before proceeding

### Step 2: Scope assessment

- **Single concern, ≤4 phases** → proceed to phase definition
- **Multiple concerns or >4 phases** → propose decomposition into separate plans with scope, files, and dependencies. After approval, scaffold individual plan files for each unit.

### Step 3: Phase definition

Break the proposal's changes into ordered phases:
- Group related changes that must ship together
- Respect dependency order (data model → business logic → API → UI)
- Each phase should leave the codebase in a working, testable state
- Include tests in the same phase as the code they test

Present proposed phases for buy-in before writing the full plan.

### Step 4: Write the plan

Use `rpi` to scaffold and save a plan artifact linked to the proposal. Fill in all phases with:
- Overview of what the phase accomplishes and its dependencies
- Tasks with file paths and change descriptions (include key code snippets)
- Tests in the same phase as the code they test
- Success criteria split into automated (use the project's actual test/lint commands) and manual verification
- Commit step (stage list + message)
- "Pause for manual confirmation" between phases

### Step 5: Transition upstream artifacts

After the plan is written, verify it covers all the proposal's design decisions — nothing silently dropped. Use `rpi` to transition the proposal to complete. If the proposal links to research still marked active, check it too and transition if covered. Note any gaps and ask.

### Step 6: Review & iterate

Present the plan summary and ask if anything needs adjusting. Keep iterating until the user confirms.

---

## Guidelines

1. **Be interactive** — get buy-in on phases before writing the full plan
2. **Be practical** — incremental, testable changes that keep the codebase working
3. **Separate verification** — always split success criteria into automated and manual
4. **Right-size the plan** — simple tasks get simple plans (1 phase, minimal ceremony); complex tasks get detailed phasing
5. **Commit after each phase** — stage only that phase's files
6. **Tests belong to their phase** — write tests alongside the code they cover, not in a separate section
7. **Trust prior stages** (pipeline mode) — don't redo research or proposal work; reference those docs
8. **Spot-check reality** (pipeline mode) — verify the codebase matches the proposal before planning

If the user confirms the plan looks good, proceed to `/rpi-implement` with the plan path.
