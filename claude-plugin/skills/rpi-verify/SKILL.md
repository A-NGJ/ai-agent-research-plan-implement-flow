---
name: rpi-verify
description: Verify an implementation against its design and spec across completeness, correctness, and coherence — produce a severity-classified review. Use when user says 'verify the implementation', 'check the implementation matches the design', or 'review what I just shipped'.
---

# Verify Implementation

## Goal

Validate that an implementation matches its design artifacts across three dimensions: completeness, correctness, and coherence. Produce a severity-classified verification report in `.rpi/reviews/`. This command is purely advisory — it never blocks anything.

If no path provided, auto-detect from recent git changes. If artifacts found, announce what you're verifying.

## Invariants

- Resolve the artifact chain from the provided or detected artifact — read all linked files (plan → design → research)
- Search for prior verify-reports covering the same area; use them as historical context, not as a new spec source
- Check `.rpi/specs/` for relevant specs and get the list of changed files
- Read actual implementation files — never trust summaries or checkboxes
- **Completeness**: check all plan phases/tasks done, tests exist, all planned files created/modified, scan for TODO/FIXME/HACK markers
- **Correctness**: extract scenarios from linked specs using the verify spec tool, then verify each scenario against actual code and tests with pass/fail per scenario and file:line references; check API contracts match design, flag silent deviations
- **Coherence**: verify naming conventions, error handling, code organization follow existing patterns; check for unnecessary dependencies
- Classify each finding as: blocker (must fix), warning (should fix), or note (consider fixing)
- Scaffold a verification report, fill in findings grouped by dimension and severity
- Present summary: overall status, counts by severity, report path; list blockers directly

## Principles

- Be specific — every finding includes a file:line reference
- Severity matters — distinguish genuine blockers from style nits
- Scale effort — small implementations get lighter verification; large ones get thorough checks
- Re-runnable — each run produces a new report file
