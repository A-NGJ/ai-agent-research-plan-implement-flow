---
description: Verify implementation against proposal artifacts for completeness, correctness, and coherence
model: opus
---

# Verify Implementation

Validate that an implementation matches its proposal artifacts across three dimensions: completeness, correctness, and coherence. Produces a severity-classified verification report.

This command is purely advisory — it does not block anything. It can be re-run after fixes to confirm resolution.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.
Run `rpi --help` to discover available commands and `rpi <command> --help` for detailed usage with examples.

## Step 1: Receive the input

If the user provided a path to a proposal, plan, or other artifact as command arguments, proceed to Step 3.

If no arguments were provided, proceed to Step 2 (auto-detection).

## Step 2: Auto-detect what to verify

When no path is provided, detect from recent git changes:

1. Use `rpi` to get the list of changed files and find active plans/proposals
2. If artifacts found, announce what you're verifying and proceed to Step 3
3. If nothing found, ask for a path to a plan or proposal

## Step 3: Read referenced artifacts

Read the provided or detected artifact(s) fully. Use `rpi` to resolve the artifact chain — this returns the full chain (plan → proposal → research) with metadata. Read all linked files.

Also check `.thoughts/specs/` for relevant specs, and use `rpi` to get the list of changed files.

Present a brief summary of what you're verifying before proceeding.

## Step 4: Verify across three dimensions

Verify all three dimensions — parallelize when possible. Each dimension requires reading the actual implementation files, not trusting summaries or checkboxes.

### Completeness

Check whether everything planned has been done:
- Are all plan phases and tasks complete? Use `rpi` for mechanical checks (checkbox counts, file coverage, marker scans for TODO/FIXME/HACK).
- If a ticket exists, are all acceptance criteria met?
- Do tests exist for new functionality?
- Were all planned files created or modified?

### Correctness

Check whether the implementation matches the design intent:
- Does the implementation follow the approach chosen in the proposal?
- Do API contracts, function signatures, and data shapes match what was specified?
- Were edge cases identified in the proposal handled in code?
- Are there silent deviations — files changed or approaches used that weren't in the plan?
- If specs exist in `.thoughts/specs/`, is the implementation consistent with them?

### Coherence

Check whether the implementation fits the existing codebase:
- Do naming conventions, error handling, and code organization follow existing patterns?
- Does the new code reuse existing utilities rather than reinventing?
- Were unnecessary dependencies introduced?
- Do tests follow the project's existing test style?

For each finding, classify as: blocker (must fix), warning (should fix), or note (consider fixing).

## Step 5: Synthesize verification report

After all three dimensions are verified:

1. **Determine overall status**: Pass / Pass with warnings / Issues found
2. **Write the report**: Use `rpi` to scaffold a verification report in `.thoughts/reviews/`. Fill in findings grouped by dimension and severity.
3. **Present the summary** — overall status, counts by severity, report path. If blockers exist, list them directly so the user sees them immediately.

## Guidelines

- **Purely advisory** — this command never blocks anything
- **Re-runnable** — each run produces a new report file
- **Read actual code** — don't trust checkboxes or summaries
- **Be specific** — every finding should include a file:line reference
- **Severity matters** — distinguish genuine blockers from style nits
- **Check specs** — if `.thoughts/specs/` contains relevant specs, verify behavioral consistency
- **Scale effort** — small implementations get lighter verification; large ones get thorough checks
