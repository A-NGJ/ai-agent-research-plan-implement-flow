# AI Agent: Research-Plan-Implement Flow

A structured development workflow for AI coding agents that turns vague feature requests into shipped code through a pipeline of discrete, reviewable stages. Built for [Claude Code](https://docs.anthropic.com/en/docs/claude-code), but the underlying methodology — Research → Design → Plan → Implement — works with any AI coding tool.

Instead of asking an AI to "just implement it" and hoping for the best, this workflow forces deliberate progression through **Research → Design → Plan → Implement** — with optional stages for complex work. Each stage produces a document you can review, edit, and approve before moving on.

```
Research → Design → Plan → Implement
   │          │        │        │
   ▼          ▼        ▼        ▼
.thoughts/  .thoughts/ .thoughts/ code +
research/   designs/   plans/     tests +
                                  commits
```

## Table of Contents

- [Why This Exists](#why-this-exists)
- [Quick Start](#quick-start)
- [Choosing Your Path](#choosing-your-path)
  - [Small Tasks](#small-tasks-bug-fixes-config-changes-simple-refactors)
  - [Medium Tasks](#medium-tasks-focused-features-single-concern-changes)
  - [Large Tasks](#large-tasks-multi-concern-features-major-refactors-greenfield-projects)
  - [Greenfield or Major Reorganization](#greenfield-or-major-reorganization)
  - [Not Sure Where to Start?](#not-sure-where-to-start)
  - [After Implementation](#after-implementation)
- [How Each Stage Works](#how-each-stage-works)
- [The `.thoughts/` Directory](#the-thoughts-directory)
  - [Sharing with Your Team](#sharing-thoughts-with-your-team)
- [The `claude-init` Script](#the-claude-init-script)
- [Tips](#tips)
- [Using with Other AI Coding Tools](#using-with-other-ai-coding-tools)
- [Project Structure](#project-structure)
- [License](#license)

## Why This Exists

AI coding assistants are powerful but unpredictable when given large tasks. They skip steps, make questionable architectural choices, and produce code that doesn't fit the codebase. This workflow solves that by:

- **Separating thinking from doing** — Research documents facts without opinions. Design makes decisions with trade-offs. Plans specify exact changes. Implementation follows the plan.
- **Creating review checkpoints** — You approve each stage before the next one starts. Bad decisions get caught early, not after 500 lines of wrong code.
- **Building persistent context** — All artifacts live in `.thoughts/`, so you and your team (or the AI) can pick up where you left off across sessions.
- **Scaling to complexity** — Simple bug fix? Skip straight to Plan → Implement. Complex feature spanning multiple systems? Use the full pipeline with Tickets.
- **Keeping the context window small** — LLMs produce better output when focused. By breaking work into stages, each conversation stays scoped to one job (research *or* design *or* implementation) rather than cramming everything into a single bloated context. The `.thoughts/` documents carry knowledge between stages, so the AI starts each stage with exactly the context it needs — no more, no less.

## Quick Start

### Prerequisites

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and configured

### Installation

1. Clone this repository:
   ```bash
   git clone <repo-url>
   cd ai-agent-research-plan-implement-flow
   ```

2. Run the initialization script in your target project:
   ```bash
   # From your project directory
   /path/to/ai-agent-research-plan-implement-flow/bin/claude-init --all /path/to/your/project
   ```

   This creates:
   - `.claude/` — Agents, commands, skills, and hooks for Claude Code
   - `.thoughts/` — Directory for all pipeline artifacts (gitignored by default)
   - `CLAUDE.md` — Project-level instructions for Claude Code

   Add `--track-thoughts` to commit `.thoughts/` to git so your team can share research, designs, and plans.

3. Alternatively, copy the `.claude/` directory manually into your project:
   ```bash
   cp -r .claude/ /path/to/your/project/.claude/
   mkdir -p /path/to/your/project/.thoughts/{research,designs,plans,structures,tickets,prs,reviews}
   ```

4. Start Claude Code in your project and use the slash commands.

### The Slash Commands

| Command | What It Does | Output |
|---------|-------------|--------|
| `/rpi-explore` | Investigates opportunities and ideas — opinionated scouting | Conversational (optional save) |
| `/rpi-research` | Investigates the codebase — pure fact-finding, no opinions | `.thoughts/research/YYYY-MM-DD-topic.md` |
| `/rpi-design` | Makes architectural decisions with trade-off analysis | `.thoughts/designs/YYYY-MM-DD-topic.md` |
| `/rpi-structure` | Defines file layout, module boundaries, interfaces | `.thoughts/structures/YYYY-MM-DD-topic.md` |
| `/rpi-tickets` | Breaks a design into independently plannable work units | `.thoughts/tickets/prefix-NNN-name.md` |
| `/rpi-plan` | Creates phased implementation plan with success criteria | `.thoughts/plans/YYYY-MM-DD-topic.md` |
| `/rpi-implement` | Executes a plan phase-by-phase with verification | Code, tests, and commits |
| `/rpi-commit` | Creates focused git commits with smart grouping | Git commits |
| `/rpi-verify` | Validates implementation matches design artifacts | Verification report |
| `/rpi-archive` | Archives completed artifacts to keep `.thoughts/` clean | Moves files to `.thoughts/archive/` |

## Choosing Your Path

Not every task needs every stage. Pick the path that matches your task's complexity.

### Small Tasks (bug fixes, config changes, simple refactors)

**Path: Plan → Implement**

Skip research and design entirely. `/rpi-plan` has a standalone mode that does lightweight research on the fly and produces a focused plan.

**Example — Fix a broken date formatter:**
```
You:  /rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings without timezone
```

Claude researches the file, writes a 1-phase plan with the fix and test, saves it to `.thoughts/plans/`.

```
You:  /rpi-implement .thoughts/plans/2026-03-04-fix-date-formatter.md
```

Claude shows you the intended changes, you approve, it implements, runs tests, and proposes a commit.

**Example — Add a missing environment variable:**
```
You:  /rpi-plan Add REDIS_URL to the config module with a default of localhost:6379
```

One phase, two files modified, done in minutes.

### Medium Tasks (focused features, single-concern changes)

**Path: Research → Design → Plan → Implement**

Use when the feature touches multiple files, involves a choice between approaches, or you're working in unfamiliar code.

**Example — Add rate limiting to the API:**

**Step 1: Research**
```
You:  /rpi-research How does the API middleware chain work? Where are requests authenticated and validated?
```
Claude spawns sub-agents to explore your codebase in parallel. Returns a document with every relevant file path, the middleware execution order, existing patterns, and how requests flow through the system.

You review `.thoughts/research/2026-03-04-api-middleware.md` — maybe you redirect: "Focus more on the error handling middleware, I want rate limit errors to follow that pattern."

**Step 2: Design**
```
You:  /rpi-design .thoughts/research/2026-03-04-api-middleware.md
      I want to add per-endpoint rate limiting. Should handle both authenticated and anonymous users.
```
Claude presents 2-3 options (e.g., in-memory vs Redis, middleware vs decorator pattern), with pros/cons tied to your actual codebase. You pick an approach. It writes the design document with the decision rationale, component diagram, and risk assessment.

**Step 3: Plan**
```
You:  /rpi-plan .thoughts/designs/2026-03-04-api-rate-limiting.md
```
Claude reads the design, spot-checks that the codebase still matches, and breaks the work into phases:
- Phase 1: Rate limiter core module + unit tests
- Phase 2: Middleware integration + integration tests
- Phase 3: Per-endpoint configuration + documentation

Each phase has specific file changes, code snippets, automated verification commands (pulled from your `CLAUDE.md`), and manual verification steps.

**Step 4: Implement**
```
You:  /rpi-implement .thoughts/plans/2026-03-04-api-rate-limiting.md
```
Claude implements Phase 1, shows you a preview of all changes before writing code, runs the test suite, updates checkboxes in the plan, and proposes a commit. Then it pauses for your manual verification before starting Phase 2.

If something doesn't match the plan — say a file was refactored since the design was written — Claude stops and tells you exactly what diverged, rather than silently improvising.

### Large Tasks (multi-concern features, major refactors, greenfield projects)

**Path: Research → Design → Tickets → Plan (per ticket) → Implement (per ticket)**

Use when the work spans multiple systems, would produce more than ~4 implementation phases, or you want to ship incrementally over multiple sessions.

**Example — Build a notification system (email, push, in-app):**

**Step 1: Research**
```
You:  /rpi-research How do we currently send emails? Is there any notification infrastructure? How do user preferences work?
```
Produces a thorough map of everything notification-adjacent in your codebase.

**Step 2: Design**
```
You:  /rpi-design .thoughts/research/2026-03-04-notification-infrastructure.md
      Build a notification system supporting email, push, and in-app channels. Users should be able to set per-channel preferences.
```
Claude goes into comprehensive mode — spawns multiple research sub-agents, identifies key design dimensions (channel abstraction, delivery strategy, preference storage, template system), presents options with trade-offs for each, validates that the chosen options compose well together, and writes the full design document.

**Step 3: Tickets**
```
You:  /rpi-tickets .thoughts/designs/2026-03-04-notification-system.md
```
Claude decomposes the design into discrete work units:
```
1. notif-001: Core notification model + channel interface
2. notif-002: Email channel implementation
3. notif-003: Push notification channel
4. notif-004: In-app notification channel + UI
5. notif-005: User preference management
6. notif-006: Notification dispatch service (orchestrator)
```
Each ticket is self-contained with its own scope, acceptance criteria, file list, and extracted design context. An index file shows the dependency graph and recommended implementation order.

**Step 4-5: Plan and Implement (per ticket)**
```
You:  /rpi-plan .thoughts/tickets/notif-001-core-model.md
You:  /rpi-implement .thoughts/plans/2026-03-04-notif-001-core-model.md

You:  /rpi-plan .thoughts/tickets/notif-002-email-channel.md
You:  /rpi-implement .thoughts/plans/2026-03-04-notif-002-email-channel.md

...and so on for each ticket
```

Each ticket becomes its own plan → implement cycle. You can stop between tickets, come back the next day, and pick up where you left off — the `.thoughts/` directory preserves all context.

### Greenfield or Major Reorganization

**Path: Research → Design → Structure → Tickets → Plan → Implement**

Add the `/rpi-structure` stage when you need to make deliberate decisions about file layout, module boundaries, and interfaces before planning implementation — typically for new projects with many new files or large-scale reorganizations.

```
You:  /rpi-structure .thoughts/designs/2026-03-04-new-service.md
```

This produces a structure document with every new and modified file, their responsibilities, module boundaries, public interfaces with concrete signatures, and a dependency graph. The structure document then feeds into `/rpi-tickets` or `/rpi-plan`.

### Not Sure Where to Start?

Use `/rpi-explore` when you have a vague idea but aren't ready to commit to a direction. It's opinionated scouting — unlike `/rpi-research` (which maps the terrain neutrally), explore forms assessments and recommends next steps.

```
You:  /rpi-explore What could we improve about error handling?
```

Claude investigates broadly, surfaces opportunities, and suggests which pipeline path to take next.

### After Implementation

Two optional commands help close the loop:

- **`/rpi-verify`** — Validates that your implementation matches the design artifacts. Checks completeness, correctness, and coherence. Run it after `/rpi-implement` or anytime you want a second opinion on whether the code matches the plan.
- **`/rpi-archive`** — Moves completed artifacts to `.thoughts/archive/` to keep the active directory clean. Run it when a feature is fully shipped and you no longer need the research/design/plan documents in the active directories.

## How Each Stage Works

### Explore (`/rpi-explore`)

**Purpose:** Investigate a topic or vague idea to surface opportunities and recommended next steps — with opinions.

Unlike `/rpi-research` (pure fact-finding), explore is opinionated scouting. It forms assessments, identifies what's worth doing, and suggests which pipeline path to take. Use it when you have a question like "should we migrate to X?" or "what could we improve about Y?" but aren't ready to commit to a direction.

Output is conversational by default. You can optionally save findings to `.thoughts/research/` if they're worth preserving.

### Research (`/rpi-research`)

**Purpose:** Document the codebase as-is. Pure fact-finding — no opinions, no recommendations.

The research command spawns parallel sub-agents that use specialized skills:
- **locate-codebase** — Finds where files and components live
- **codebase-analyzer** — Understands how specific code works (traces data flow, documents patterns)
- **find-patterns** — Finds examples of existing patterns to model after
- **locate-thoughts** — Discovers relevant historical documents in `.thoughts/`
- **analyze-thoughts** — Extracts key insights from existing documents

For broad queries, Claude shows you what it found in an initial scan and asks if you want to redirect focus before deep-diving.

Output is a structured markdown document with YAML frontmatter, file:line references throughout, and a clear separation between findings and open questions.

### Design (`/rpi-design`)

**Purpose:** Make architectural decisions and document trade-offs.

Three modes, auto-detected:
- **Lightweight** — Focused decision between 2-3 options (single component, one pattern choice)
- **Comprehensive** — Multi-decision feature design with component diagrams and risk tables
- **Incremental** — Update an existing design with new information

The design stage is interactive. Claude presents options with concrete trade-offs (grounded in your actual codebase, not generic advice), makes a recommendation, and waits for your direction. After you choose, it validates that your combined choices work together before documenting.

### Structure (`/rpi-structure`) — Optional

**Purpose:** Map a design to concrete file layout when the structure itself is complex.

Only needed for greenfield projects or major reorganizations. Defines file changes, module boundaries, public APIs with concrete signatures, and a dependency graph.

### Tickets (`/rpi-tickets`) — Optional

**Purpose:** Break a large design into independently plannable work units.

Each ticket is self-contained — it extracts the relevant design decisions, interfaces, and constraints so that `/rpi-plan` can produce a focused plan without reading the full design document. Tickets include dependency graphs and recommended implementation order.

### Plan (`/rpi-plan`)

**Purpose:** Create a phase-by-phase implementation plan with specific code changes and verification steps.

Works in two modes:
- **Standalone** — For simple tasks. Does its own lightweight research and produces a plan directly.
- **Pipeline** — For complex tasks with existing design/ticket documents. Reads the full document chain, spot-checks the codebase against the docs, and breaks work into verified phases.

Every plan phase includes:
- Specific file changes with code snippets
- Tests (in the same phase as the code they test, not a separate "testing phase")
- Automated verification commands (pulled from your project's `CLAUDE.md`)
- Manual verification steps
- A commit with specific files and message

All open questions must be resolved before the plan is finalized.

### Implement (`/rpi-implement`)

**Purpose:** Execute a plan phase-by-phase with verification at each step.

The implementation stage:
1. Reads the plan completely
2. For each phase, shows a **pre-review** of all intended changes before writing code
3. Implements the phase after approval
4. Runs automated verification (tests, linting, type checks)
5. Updates checkboxes in the plan file
6. Proposes a commit (no Claude attribution)
7. Pauses for manual verification before the next phase

If the plan doesn't match reality (codebase drifted since the plan was written), it stops and clearly explains the mismatch rather than silently improvising.

Resumable: if you invoke `/rpi-implement` on a partially-completed plan, it picks up from the first unchecked item.

### Bonus: Commit (`/rpi-commit`)

**Purpose:** Create clean, focused git commits without thinking about `git add` and message formatting.

`/rpi-commit` inspects your working tree (staged, unstaged, and untracked files), groups related changes into logical commits, drafts messages matching your repo's existing commit style, and presents the plan for your approval before executing. It never adds Claude attribution or co-author lines — commits look like you wrote them.

You can use `/rpi-commit` standalone anytime, or let `/rpi-implement` handle commits at the end of each phase.

### Verify (`/rpi-verify`)

**Purpose:** Validate that an implementation matches its design artifacts.

Checks three dimensions:
- **Completeness** — Are all planned changes implemented?
- **Correctness** — Does the code match the design decisions?
- **Coherence** — Do the pieces work together as intended?

Can auto-detect what to verify from recent git changes and active plans, or you can point it at a specific design doc, plan, or ticket. Produces a severity-classified report. Purely advisory — it doesn't block anything, and can be re-run after fixes.

### Archive (`/rpi-archive`)

**Purpose:** Move completed artifacts to `.thoughts/archive/` to keep the active directory clean.

Two modes:
- **Specific paths** — `/rpi-archive .thoughts/research/2026-01-15-auth-flow.md`
- **Scan mode** — `/rpi-archive` with no arguments scans for completed artifacts

Warns before archiving anything still in `draft` or `active` status. Preserves the full directory structure inside `archive/` (e.g., `archive/research/`, `archive/designs/`).

## The `.thoughts/` Directory

All pipeline artifacts live in `.thoughts/`, which is **gitignored by default**:

```
.thoughts/
├── PIPELINE.md          # This workflow reference guide
├── research/            # Codebase research documents
├── designs/             # Architectural design documents
├── structures/          # File layout and interface documents
├── tickets/             # Scoped work unit tickets + index
├── plans/               # Implementation plans with checkboxes
├── prs/                 # PR descriptions
├── reviews/             # Code review and verification reports
└── archive/             # Completed artifacts (mirrors above structure)
```

Files follow the naming convention: `YYYY-MM-DD-descriptive-name.md`

This directory serves as persistent context across sessions. You can read, edit, or delete any document. Claude will check for existing documents before creating new ones to avoid duplication.

### Sharing `.thoughts/` with Your Team

By default, `.thoughts/` is added to `.gitignore` — useful when you want pipeline artifacts to stay local. But these documents can be valuable to the whole team: research captures institutional knowledge about the codebase, designs document *why* decisions were made, and plans provide a record of what was implemented and how.

To track `.thoughts/` in git, use the `--track-thoughts` flag during initialization:

```bash
claude-init --all --track-thoughts
```

This skips adding `.thoughts/` to `.gitignore`, so all pipeline artifacts get committed alongside your code.

**Why share with the team:**
- **Research documents** become searchable codebase documentation that stays current — new team members can read them to understand how systems work instead of spelunking through code.
- **Design documents** preserve decision rationale. When someone asks "why did we use Redis instead of Memcached?", the answer is in `.thoughts/designs/`, not lost in a Slack thread.
- **Tickets and plans** give visibility into how features were decomposed and implemented, making code review easier and providing a template for similar future work.
- **Any team member can pick up where another left off** — if one person does the research and design, another can run `/rpi-plan` and `/rpi-implement` using those same documents.

**When to keep it local:**
- Exploratory or throwaway research you don't want cluttering the repo
- Plans for personal refactors or experiments
- When your team prefers to keep this context in other tools (Notion, Linear, etc.)

You can also take a hybrid approach: gitignore `.thoughts/` by default but selectively commit specific documents that have lasting value.

## The `claude-init` Script

The `bin/claude-init` script bootstraps the workflow into any project:

```bash
# Basic init (creates .claude/ directory structure + CLAUDE.md + .thoughts/)
claude-init

# Init with all agents, commands, and skills from your dotfiles
claude-init --all

# Init a specific project directory
claude-init --all ~/projects/my-app

# Init with .thoughts/ tracked in git (for team sharing)
claude-init --all --track-thoughts

# Update existing configs from dotfiles (preserves local changes)
claude-init --update

# Options
claude-init --force            # Overwrite existing .claude/
claude-init --no-claude-md     # Skip CLAUDE.md creation
claude-init --agents-only      # Only copy agents
claude-init --commands-only    # Only copy commands
claude-init --skills-only     # Only copy skills
claude-init --track-thoughts   # Don't gitignore .thoughts/ (track in git)
```

The script copies agents, commands, skills, and hooks from your global `~/.claude/` directory. Set the `DOTFILES_CLAUDE` environment variable to use a different source directory (e.g., `DOTFILES_CLAUDE=~/dotfiles/.claude claude-init --all`).

## Tips

- **Start small.** Try `/rpi-plan` on a bug fix to see how the plan → implement cycle feels before using the full pipeline.
- **Edit the artifacts.** The `.thoughts/` documents are yours. If a design decision is wrong, edit it before planning. If a plan phase is unnecessary, delete it.
- **Use CLAUDE.md.** Add your project's test commands, linting setup, and conventions to `CLAUDE.md`. The pipeline stages pull verification commands from there.
- **Redirect during research.** When `/rpi-research` shows initial findings, tell it to focus on specific areas rather than exploring everything.
- **Skip stages when they don't add value.** The full pipeline exists for complex work. Most daily tasks only need Plan → Implement.
- **Review the pre-review.** `/rpi-implement` shows you exactly what it plans to change before writing code. This is your last checkpoint — use it.

## Using with Other AI Coding Tools

This workflow is built for Claude Code, but the methodology applies to any AI coding agent. The `.thoughts/` directory, document templates, and staged pipeline work regardless of tooling. For other tools, follow their documentation on how to register custom commands and load prompt files, then adapt the files in `.claude/` accordingly.

## Project Structure

```
.
├── bin/
│   ├── claude-init                    # Project initialization script
│   └── templates/
│       ├── CLAUDE.md.template         # Template for project CLAUDE.md
│       └── PIPELINE.md.template       # Pipeline reference document
├── .claude/
│   ├── agents/
│   │   └── codebase-analyzer.md       # Agent: traces code flow and documents implementations
│   ├── commands/
│   │   ├── rpi-explore.md             # Command: /rpi-explore
│   │   ├── rpi-research.md            # Command: /rpi-research
│   │   ├── rpi-design.md              # Command: /rpi-design
│   │   ├── rpi-structure.md           # Command: /rpi-structure
│   │   ├── rpi-tickets.md             # Command: /rpi-tickets
│   │   ├── rpi-plan.md                # Command: /rpi-plan
│   │   ├── rpi-implement.md           # Command: /rpi-implement
│   │   ├── rpi-verify.md              # Command: /rpi-verify
│   │   ├── rpi-archive.md             # Command: /rpi-archive
│   │   └── rpi-commit.md              # Command: /rpi-commit
│   └── skills/
│       ├── find-patterns/SKILL.md     # Skill: finds existing code patterns to model after
│       ├── analyze-thoughts/SKILL.md  # Skill: extracts insights from .thoughts/ documents
│       ├── locate-thoughts/SKILL.md   # Skill: discovers documents in .thoughts/
│       └── locate-codebase/SKILL.md   # Skill: finds where code lives in the codebase
```

## Acknowledgments

Inspired by [HumanLayer](https://github.com/humanlayer/humanlayer) — their work on human-in-the-loop patterns for AI agents informed the design of this workflow.

## License

MIT
