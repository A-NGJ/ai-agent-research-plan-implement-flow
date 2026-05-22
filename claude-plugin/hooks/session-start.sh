#!/bin/sh
# SessionStart hook for the rpi Claude Code plugin.
#
# Injects the RPI workflow framing into the conversation context so the model
# understands the pipeline, which skill to invoke when, and the cross-skill
# flag contract — without writing anything to user-owned files.
#
# stdout is captured by Claude Code and appended to the session context.

cat <<'EOF'
# RPI workflow

This project has the rpi plugin loaded. RPI is a Spec-Driven Development
workflow with persistent artifacts under `.rpi/`:

  .rpi/research/   research notes (codebase or external)
  .rpi/designs/    solution designs (with tradeoffs)
  .rpi/specs/      living behavioral specs (source of truth)
  .rpi/plans/      phased implementation plans
  .rpi/diagnoses/  bug post-mortems
  .rpi/reviews/    verification reports
  .rpi/archive/    completed/superseded artifacts

## Pipeline

  Research → Propose → Plan → Implement → Verify

Each step has a skill. Each skill suggests the next.

  /rpi:rpi-research   Investigate a question — codebase ("how does X work?") or
                  external ("what frameworks exist for X?"). Optional.
  /rpi:rpi-propose    Design a new feature with tradeoffs → produces design + spec.
                  Approval gate.
  /rpi:rpi-plan       Plan a scoped, concrete change → phased implementation plan.
                  Use for narrow tweaks or bug fixes; use /rpi:rpi-propose for
                  changes that require weighing tradeoffs.
  /rpi:rpi-implement  Execute an approved plan phase by phase with per-phase
                  verification.
  /rpi:rpi-verify     Validate an implementation against its design and spec.
                  Severity-classified review.
  /rpi:rpi-diagnose   Iterative root-cause analysis for complex bugs.
  /rpi:rpi-commit     Stage and commit current work with safety scans.
  /rpi:rpi-explain    Walk through a recent diff and explain what changed and why.
  /rpi:rpi-archive    Move complete/superseded artifacts to .rpi/archive/.
  /rpi:rpi-handoff    Capture in-flight context for the next session.
  /rpi:rpi-spec-sync  Detect spec drift and resync .rpi/specs/ to the codebase.

## Flag contract

  --ff     Fast-forward: skip pre-review and manual-verification pauses; mark
           unverified items as [unverified — --ff] and run the chain's
           terminal step automatically.
  --grill  After producing a plan or design, enter the grill-me skill to
           stress-test it before approval.

## When to start

  Features (new behavior, tradeoffs):     /rpi:rpi-propose
  Concrete narrow change or bug fix:      /rpi:rpi-plan
  Complex bug with unclear cause:         /rpi:rpi-diagnose
  Open-ended exploration:                 /rpi:rpi-research

If unsure: ask the user before invoking a pipeline skill.
EOF
