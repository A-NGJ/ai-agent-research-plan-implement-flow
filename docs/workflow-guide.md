# Choosing Your Path

Not every task needs every stage. Pick the path that matches your task's complexity.

## Small Tasks (bug fixes, config changes, simple refactors)

**Path: Plan -> Implement**

Skip exploration and proposals entirely. `/rpi-plan` has a standalone mode that does lightweight research on the fly and produces a focused plan.

**Example -- Fix a broken date formatter:**
```
You:  /rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings without timezone
```

Claude researches the file, writes a 1-phase plan with the fix and test, saves it to `.rpi/plans/`.

```
You:  /rpi-implement .rpi/plans/2026-03-04-fix-date-formatter.md
```

Claude implements the phase, runs tests, and auto-commits when they pass.

**Example -- Add a missing environment variable:**
```
You:  /rpi-plan Add REDIS_URL to the config module with a default of localhost:6379
```

One phase, two files modified, done in minutes.

## Medium Tasks (focused features, single-concern changes)

**Path: Propose -> Plan -> Implement**

Use when the feature touches multiple files, involves a choice between approaches, or you're working in unfamiliar code. Optionally run `/rpi-research` first if you need to understand the codebase -- or survey external systems/frameworks -- before proposing.

**Example -- Add rate limiting to the API:**

**Step 1 (optional): Research**
```
You:  /rpi-research How does the API middleware chain work? Where are requests authenticated and validated?
```
Claude explores your codebase conversationally. You discuss findings interactively -- no artifact is created by default. If the exploration is thorough enough, you can ask it to save findings to `.rpi/research/`.

For external-system questions (e.g., *"what agentic frameworks exist for a data analytics tool?"*), `/rpi-research` surveys docs, READMEs, and release notes instead of source files -- citing URLs or quoted documentation rather than `file:line` references.

**Step 2: Propose**
```
You:  /rpi-propose I want to add per-endpoint rate limiting. Should handle both authenticated and anonymous users.
```
Claude investigates the codebase, presents 2-3 options (e.g., in-memory vs Redis, middleware vs decorator pattern), with pros/cons tied to your actual codebase. You pick an approach. It writes the design with the decision rationale and risk assessment to `.rpi/designs/`.

**Step 3: Plan**
```
You:  /rpi-plan .rpi/designs/2026-03-04-api-rate-limiting.md
```
Claude reads the proposal, spot-checks the codebase against the docs, and breaks the work into phases:
- Phase 1: Rate limiter core module + unit tests
- Phase 2: Middleware integration + integration tests
- Phase 3: Per-endpoint configuration + documentation

Each phase has specific file changes, code snippets, automated verification commands (pulled from your `CLAUDE.md`), and manual verification steps.

**Step 4: Implement**
```
You:  /rpi-implement .rpi/plans/2026-03-04-api-rate-limiting.md
```
For each phase, Claude runs the test suite, auto-commits when checks pass, updates checkboxes in the plan, and advances to the next phase. It only pauses when the plan includes manual verification items not covered by automated tests.

If something doesn't match the plan -- say a file was refactored since the proposal was written -- Claude stops and tells you exactly what diverged, rather than silently improvising.

## Large Tasks (multi-concern features, major refactors, greenfield projects)

**Path: Propose -> Plan (decomposes) -> Implement (per unit)**

Use when the work spans multiple systems, would produce more than ~4 implementation phases, or you want to ship incrementally over multiple sessions.

**Example -- Build a notification system (email, push, in-app):**

**Step 1 (optional): Research**
```
You:  /rpi-research How do we currently send emails? Is there any notification infrastructure? How do user preferences work?
```
Conversational research to build understanding before proposing.

**Step 2: Propose**
```
You:  /rpi-propose Build a notification system supporting email, push, and in-app channels. Users should be able to set per-channel preferences.
```
Claude goes into Full mode -- investigates the codebase deeply, identifies key design dimensions (channel abstraction, delivery strategy, preference storage, template system), presents options with trade-offs for each, validates that the chosen options compose well together, and writes the full design to `.rpi/designs/`.

**Step 3: Plan (with decomposition)**
```
You:  /rpi-plan .rpi/designs/2026-03-04-notification-system.md
```
Claude reads the proposal and decomposes it into independently plannable units:
```
1. Core notification model + channel interface
2. Email channel implementation
3. Push notification channel
4. In-app notification channel + UI
5. User preference management
6. Notification dispatch service (orchestrator)
```
Each unit has its own phases, file changes, and verification steps within the plan.

**Step 4: Implement (per unit)**
```
You:  /rpi-implement .rpi/plans/2026-03-04-notification-system.md
```

Claude implements unit by unit. You can stop between units, come back the next day, and pick up where you left off -- the `.rpi/` directory preserves all context and checkboxes track progress.

## Not Sure Where to Start?

Use `/rpi-research` even when you have a vague idea. It handles focused codebase questions ("how does auth work?"), open-ended exploration ("what could we improve about error handling?"), and external surveys ("what agentic frameworks exist for X?", "what's the state of vector databases in 2026?"). It's conversational -- you discuss findings interactively and decide whether to save research or move straight to `/rpi-propose`.

```
You:  /rpi-research What could we improve about error handling?
```

## After Implementation

Several optional commands help close the loop:

- **`/rpi-verify`** -- Validates that your implementation matches the proposal artifacts. Checks completeness, correctness, and coherence. Run it after `/rpi-implement` or anytime you want a second opinion on whether the code matches the plan.
- **`/rpi-explain`** -- Walks through the diff with a file-by-file explanation. Useful for self-review or explaining changes to a teammate.
- **`/rpi-spec-sync`** -- Syncs specs in `.rpi/specs/` to match the current codebase. Run it after a batch of changes to detect drift, rewrite stale scenarios, rename or merge specs, and archive obsolete ones.
- **`/rpi-archive`** -- Moves completed artifacts to `.rpi/archive/` to keep the active directory clean. Run it when a feature is fully shipped and you no longer need the research/designs/plan documents in the active directories.
- **`rpi status`** -- Dashboard showing artifact counts, active plan progress, stale artifacts, and archive readiness. Run it anytime to see the overall state of `.rpi/`.

## Tips

- **Use `/rpi-diagnose` for complex bugs.** When a bug spans multiple files or the root cause isn't obvious, `/rpi-diagnose` will iteratively trace, fix, and verify -- producing a post-mortem artifact even if the fix requires escalation to `/rpi-plan`.
- **Start small.** Try `/rpi-plan` on a bug fix to see how the plan -> implement cycle feels before using the full pipeline.
- **Edit the artifacts.** The `.rpi/` documents are yours. If a proposal decision is wrong, edit it before planning. If a plan phase is unnecessary, delete it.
- **Use CLAUDE.md.** Add your project's test commands, linting setup, and conventions to `CLAUDE.md`. The pipeline stages pull verification commands from there.
- **Redirect during research.** When `/rpi-research` shows initial findings, tell it to focus on specific areas rather than researching everything.
- **Skip stages when they don't add value.** The full pipeline exists for complex work. Most daily tasks only need Plan -> Implement.
- **Review the pre-review.** `/rpi-implement` shows you exactly what it plans to change before writing code. This is your last checkpoint -- use it.
