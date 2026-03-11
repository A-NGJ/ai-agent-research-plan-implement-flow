# Choosing Your Path

Not every task needs every stage. Pick the path that matches your task's complexity.

## Small Tasks (bug fixes, config changes, simple refactors)

**Path: Plan -> Implement**

Skip exploration and proposals entirely. `/rpi-plan` has a standalone mode that does lightweight research on the fly and produces a focused plan.

**Example -- Fix a broken date formatter:**
```
You:  /rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings without timezone
```

Claude researches the file, writes a 1-phase plan with the fix and test, saves it to `.thoughts/plans/`.

```
You:  /rpi-implement .thoughts/plans/2026-03-04-fix-date-formatter.md
```

Claude shows you the intended changes, you approve, it implements, runs tests, and proposes a commit.

**Example -- Add a missing environment variable:**
```
You:  /rpi-plan Add REDIS_URL to the config module with a default of localhost:6379
```

One phase, two files modified, done in minutes.

## Medium Tasks (focused features, single-concern changes)

**Path: Propose -> Plan -> Implement**

Use when the feature touches multiple files, involves a choice between approaches, or you're working in unfamiliar code. Optionally run `/rpi-explore` first if you need to understand the codebase before proposing.

**Example -- Add rate limiting to the API:**

**Step 1 (optional): Explore**
```
You:  /rpi-explore How does the API middleware chain work? Where are requests authenticated and validated?
```
Claude explores your codebase conversationally. You discuss findings interactively -- no artifact is created by default. If the exploration is thorough enough, you can ask it to save findings to `.thoughts/research/`.

**Step 2: Propose**
```
You:  /rpi-propose I want to add per-endpoint rate limiting. Should handle both authenticated and anonymous users.
```
Claude investigates the codebase, presents 2-3 options (e.g., in-memory vs Redis, middleware vs decorator pattern), with pros/cons tied to your actual codebase. You pick an approach. It writes the proposal with the decision rationale and risk assessment to `.thoughts/proposals/`.

**Step 3: Plan**
```
You:  /rpi-plan .thoughts/proposals/2026-03-04-api-rate-limiting.md
```
Claude reads the proposal, spot-checks the codebase against the docs, and breaks the work into phases:
- Phase 1: Rate limiter core module + unit tests
- Phase 2: Middleware integration + integration tests
- Phase 3: Per-endpoint configuration + documentation

Each phase has specific file changes, code snippets, automated verification commands (pulled from your `CLAUDE.md`), and manual verification steps.

**Step 4: Implement**
```
You:  /rpi-implement .thoughts/plans/2026-03-04-api-rate-limiting.md
```
Claude implements Phase 1, shows you a preview of all changes before writing code, runs the test suite, updates checkboxes in the plan, and proposes a commit. Then it pauses for your manual verification before starting Phase 2.

If something doesn't match the plan -- say a file was refactored since the proposal was written -- Claude stops and tells you exactly what diverged, rather than silently improvising.

## Large Tasks (multi-concern features, major refactors, greenfield projects)

**Path: Propose -> Plan (decomposes) -> Implement (per unit)**

Use when the work spans multiple systems, would produce more than ~4 implementation phases, or you want to ship incrementally over multiple sessions.

**Example -- Build a notification system (email, push, in-app):**

**Step 1 (optional): Explore**
```
You:  /rpi-explore How do we currently send emails? Is there any notification infrastructure? How do user preferences work?
```
Conversational exploration to build understanding before proposing.

**Step 2: Propose**
```
You:  /rpi-propose Build a notification system supporting email, push, and in-app channels. Users should be able to set per-channel preferences.
```
Claude goes into Full mode -- investigates the codebase deeply, identifies key design dimensions (channel abstraction, delivery strategy, preference storage, template system), presents options with trade-offs for each, validates that the chosen options compose well together, and writes the full proposal to `.thoughts/proposals/`.

**Step 3: Plan (with decomposition)**
```
You:  /rpi-plan .thoughts/proposals/2026-03-04-notification-system.md
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
You:  /rpi-implement .thoughts/plans/2026-03-04-notification-system.md
```

Claude implements unit by unit. You can stop between units, come back the next day, and pick up where you left off -- the `.thoughts/` directory preserves all context and checkboxes track progress.

## Not Sure Where to Start?

Use `/rpi-explore` even when you have a vague idea. It handles both focused questions ("how does auth work?") and open-ended exploration ("what could we improve about error handling?"). It's conversational -- you discuss findings interactively and decide whether to save research or move straight to `/rpi-propose`.

```
You:  /rpi-explore What could we improve about error handling?
```

## After Implementation

Two optional commands help close the loop:

- **`/rpi-verify`** -- Validates that your implementation matches the proposal artifacts. Checks completeness, correctness, and coherence. Run it after `/rpi-implement` or anytime you want a second opinion on whether the code matches the plan.
- **`/rpi-archive`** -- Moves completed artifacts to `.thoughts/archive/` to keep the active directory clean. Run it when a feature is fully shipped and you no longer need the research/proposals/plan documents in the active directories.
