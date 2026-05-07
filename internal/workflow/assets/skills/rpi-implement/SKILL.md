---
name: rpi-implement
description: Execute an approved plan from .rpi/plans phase by phase with TDD and per-phase verification. Use when user says 'implement the plan', 'start implementing', 'run the next phase', or just approved a plan.
---

> Before doing anything else, run `rpi bootstrap` (silent and idempotent — initializes `.rpi/` if a global install is present and the project hasn't been bootstrapped yet).

# Implement Plan

## Goal

Execute an active plan from `.rpi/plans/` phase by phase. Plans come in two forms: pipeline plans (reference designs and research) or standalone plans (self-contained).

After all phases are complete and verified, announce completion and update the plan status. Then suggest → `/rpi-verify <plan-path>` for an independent verification report.

**`--ff` (fast-forward) mode**: if invoked with `--ff`, skip the per-phase pre-review and any manual verification pauses, and after the plan transitions to complete, invoke `/rpi-verify <plan-path>` via the Skill tool as the chain's terminal step. The "On mismatch" gate, the sensitive-content scan, end-of-plan spec verification, and phase failure handling all run unchanged. Error if `--grill` is also passed.

## Invariants

- Validate the plan's status before starting — draft/active: proceed; complete: warn about duplication
- Resolve the plan's artifact chain and read all upstream context (designs, research, specs)
- Read all files mentioned in the plan fully — never use limit/offset
- Transition the plan to active status when starting
- Check current progress (completed vs remaining items) — resume from first unchecked item
- **Pre-review**: before writing code for a phase, present all intended changes for approval — flag any deviations from plan. Under `--ff`, skip the preview and write the changes directly.
- **Red/green TDD**: for new code, write tests first (confirm they fail), then implement until they pass
- Run success criteria checks after each phase — fix issues before proceeding
- Update checkboxes in the plan file as items complete
- **Before committing**: scan staged files for sensitive content — warn and exclude if flagged
- **Before committing**: scan staged files against `.gitignore` rules — warn and silently drop any matches from the stage list (commonly: plan, design, research, diagnosis, or review artifacts under the default tracked-specs policy)
- **Auto-commit**: after each phase passes its checks, commit automatically without manual confirmation — use descriptive messages matching repo style
- **After hook failure**: read error, fix the issue, re-stage, create a new commit (never amend)
- If a phase's success criteria are fully covered by automated checks (tests, linting, etc.), run them and proceed automatically when they pass — only pause for manual verification when the plan includes manual verification items not covered by automated tests. Under `--ff`, skip manual verification pauses too — mark unverified items as `[unverified — --ff]` in the plan checkbox text.
- **On mismatch**: stop, present what the plan says vs what you found, ask how to proceed
- **Context recovery**: if context seems lost or you're unsure which phase you're on, call the context essentials tool to restore your implementation context
- **On completion**: verify spec conformance for all linked specs — extract scenarios using the verify spec tool, then check each scenario against actual code and tests; plan → complete
- Under `--ff`, after the plan transitions to complete, invoke `/rpi-verify <plan-path>` via the Skill tool as the chain's terminal step.

## Principles

- Follow the plan's intent while adapting to what you find — your judgment matters
- Implement each phase fully before moving to the next
- Trust completed checkmarks when resuming — verify only if something seems off
