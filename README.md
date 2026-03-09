# AI Agent: Research-Plan-Implement Flow

A structured development workflow for AI coding agents that turns vague feature requests into shipped code through a pipeline of discrete, reviewable stages. Built for [Claude Code](https://docs.anthropic.com/en/docs/claude-code), but the underlying methodology -- Research -> Design -> Plan -> Implement -- works with any AI coding tool.

Instead of asking an AI to "just implement it" and hoping for the best, this workflow forces deliberate progression through **Research -> Design -> Plan -> Implement** -- with optional stages for complex work. Each stage produces a document you can review, edit, and approve before moving on.

```
Research -> Design -> Plan -> Implement
   |          |        |        |
   v          v        v        v
.thoughts/  .thoughts/ .thoughts/ code +
research/   designs/   plans/     tests +
                                  commits
```

## Table of Contents

- [Why This Exists](#why-this-exists)
- [Quick Start](#quick-start)
- [Choosing Your Path](#choosing-your-path)
- [How Each Stage Works](#how-each-stage-works)
- [The `.thoughts/` Directory](#the-thoughts-directory)
- [The `rpi init` Command](#the-rpi-init-command)
- [Tips](#tips)
- [Using with Other AI Coding Tools](#using-with-other-ai-coding-tools)
- [Why a Go Binary](#why-a-go-binary)
- [How It Compares](#how-it-compares)
- [Project Structure](#project-structure)
- [License](#license)

## Why This Exists

AI coding assistants are powerful but unpredictable when given large tasks. They skip steps, make questionable architectural choices, and produce code that doesn't fit the codebase. This workflow solves that by:

- **Separating thinking from doing** -- Research documents facts without opinions. Design makes decisions with trade-offs. Plans specify exact changes. Implementation follows the plan.
- **Creating review checkpoints** -- You approve each stage before the next one starts. Bad decisions get caught early, not after 500 lines of wrong code.
- **Building persistent context** -- All artifacts live in `.thoughts/`, so you and your team (or the AI) can pick up where you left off across sessions.
- **Scaling to complexity** -- Simple bug fix? Skip straight to Plan -> Implement. Complex feature spanning multiple systems? Use the full pipeline with Tickets.
- **Keeping the context window small** -- LLMs produce better output when focused. By breaking work into stages, each conversation stays scoped to one job (research *or* design *or* implementation) rather than cramming everything into a single bloated context. The `.thoughts/` documents carry knowledge between stages, so the AI starts each stage with exactly the context it needs -- no more, no less.

### Key Concepts

- **Staged pipeline** -- Work flows through discrete stages (Research → Design → Plan → Implement), each with a clear input and output. You choose how many stages to use based on task complexity.
- **`.thoughts/` as persistent context** -- All artifacts live in a local directory that survives across sessions. The AI doesn't need to re-discover your codebase every time -- it reads the documents from previous stages.
- **Artifact chains** -- Documents link to each other through frontmatter metadata (a plan links to its design, which links to its research). The `rpi chain` command resolves these links automatically so the AI loads exactly the context it needs.
- **Frontmatter-driven metadata** -- Every document carries YAML frontmatter with status, dates, tags, and cross-references. The CLI uses this for filtering, status transitions, and archive decisions -- keeping mechanical bookkeeping out of the LLM.
- **Deterministic CLI + creative LLM** -- Mechanical operations (template scaffolding, frontmatter parsing, artifact scanning, verification checks) run in a Go binary. Creative operations (research, design decisions, code generation) stay with the LLM. Each does what it's best at.

## Quick Start

### Prerequisites

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and configured
- [Go 1.23+](https://go.dev/dl/) (to build the `rpi` binary from source)
- Git

### Installation

1. Clone this repository:
   ```bash
   git clone <repo-url>
   cd ai-agent-research-plan-implement-flow
   ```

2. Build and install the `rpi` binary:
   ```bash
   make install
   ```

3. Initialize your target project:
   ```bash
   rpi init /path/to/your/project
   ```

   This creates:
   - `.claude/` -- Agents, commands, skills, templates, and hooks (embedded in the binary)
   - `.thoughts/` -- Directory for all pipeline artifacts (gitignored by default)
   - `CLAUDE.md` -- Project-level instructions for Claude Code

   Add `--track-thoughts` to commit `.thoughts/` to git so your team can share research, designs, and plans.

4. Start Claude Code in your project and use the slash commands.

### The Slash Commands

| Command | What It Does | Output |
|---------|-------------|--------|
| `/rpi-research` | Investigates the codebase -- fact-finding with optional assessment | `.thoughts/research/YYYY-MM-DD-topic.md` |
| `/rpi-design` | Makes architectural decisions with trade-off analysis | `.thoughts/designs/YYYY-MM-DD-topic.md` |
| `/rpi-structure` | Defines file layout, module boundaries, interfaces | `.thoughts/structures/YYYY-MM-DD-topic.md` |
| `/rpi-tickets` | Breaks a design into independently plannable work units | `.thoughts/tickets/prefix-NNN-name.md` |
| `/rpi-plan` | Creates phased implementation plan with success criteria | `.thoughts/plans/YYYY-MM-DD-topic.md` |
| `/rpi-implement` | Executes a plan phase-by-phase with verification | Code, tests, and commits |
| `/rpi-commit` | Creates focused git commits with smart grouping | Git commits |
| `/rpi-verify` | Validates implementation matches design artifacts | Verification report |
| `/rpi-archive` | Archives completed artifacts to keep `.thoughts/` clean | Moves files to `.thoughts/archive/` |

## Choosing Your Path

Not every task needs every stage. Match the path to your task's complexity:

- **Small tasks** (bug fixes, config changes) -- skip straight to **Plan -> Implement**. `/rpi-plan` does lightweight research on the fly.
- **Medium tasks** (focused features, single-concern changes) -- use the full **Research -> Design -> Plan -> Implement** pipeline.
- **Large tasks** (multi-concern features, major refactors) -- add **Tickets** to break the design into independently plannable units: Research -> Design -> Tickets -> Plan -> Implement (per ticket).
- **Greenfield or major reorganizations** -- add **Structure** before tickets to define file layout and interfaces upfront.

Not sure where to start? Use `/rpi-research` with any question -- it handles both focused investigation and open-ended exploration.

See the [full workflow guide](docs/workflow-guide.md) for detailed examples of each path.

## How Each Stage Works

Each slash command maps to a pipeline stage with a specific purpose. Research gathers facts, Design makes decisions, Plan specifies changes, and Implement executes them. Optional stages (Structure, Tickets) add precision for complex work.

See [detailed stage descriptions](docs/stages.md) for how each command works, its modes, and what it produces.

## The `.thoughts/` Directory

All pipeline artifacts live in `.thoughts/`, organized by type (research, designs, plans, tickets, specs, etc.). Files follow a `YYYY-MM-DD-descriptive-name.md` naming convention and track progress through a `draft -> active -> complete` status lifecycle.

By default `.thoughts/` is gitignored, but you can share it with your team using `--track-thoughts` during init.

See [full `.thoughts/` documentation](docs/thoughts-directory.md) for directory structure, naming conventions, specs, status lifecycle, and team sharing options.

## The `rpi init` Command

The `rpi init` command bootstraps the workflow into any project. All workflow files (agents, commands, skills, templates) are embedded in the `rpi` binary -- no external dotfiles or source repo needed.

```bash
rpi init ~/projects/my-app                    # Full init
rpi init --track-thoughts                      # Share .thoughts/ via git
```

See [full `rpi init` documentation](docs/rpi-init.md) for all options and flags.

## Tips

- **Start small.** Try `/rpi-plan` on a bug fix to see how the plan -> implement cycle feels before using the full pipeline.
- **Edit the artifacts.** The `.thoughts/` documents are yours. If a design decision is wrong, edit it before planning. If a plan phase is unnecessary, delete it.
- **Use CLAUDE.md.** Add your project's test commands, linting setup, and conventions to `CLAUDE.md`. The pipeline stages pull verification commands from there.
- **Redirect during research.** When `/rpi-research` shows initial findings, tell it to focus on specific areas rather than exploring everything.
- **Skip stages when they don't add value.** The full pipeline exists for complex work. Most daily tasks only need Plan -> Implement.
- **Review the pre-review.** `/rpi-implement` shows you exactly what it plans to change before writing code. This is your last checkpoint -- use it.

## Using with Other AI Coding Tools

This workflow is built for Claude Code, but the methodology applies to any AI coding agent. The `.thoughts/` directory, document templates, and staged pipeline work regardless of tooling. For other tools, follow their documentation on how to register custom commands and load prompt files, then adapt the files in `.claude/` accordingly.

## Why a Go Binary

The `rpi` CLI exists to keep mechanical work out of the LLM's context window. Every token an LLM spends on parsing YAML frontmatter, resolving file links, or generating boilerplate is a token not spent on design thinking or code generation.

The binary handles operations that are **deterministic and error-prone for LLMs**:

- **Template scaffolding** -- `rpi scaffold` generates documents with correct frontmatter, dates, and file paths. An LLM asked to do this will occasionally hallucinate fields or misformat dates.
- **Artifact chain resolution** -- `rpi chain` follows frontmatter links recursively (plan → design → research) and returns a flat list of files to load. This is a mechanical graph traversal, not a creative task.
- **Frontmatter manipulation** -- `rpi frontmatter` reads, writes, and validates status transitions. YAML parsing in natural language is fragile; a CLI does it reliably every time.
- **Directory scanning and filtering** -- `rpi scan` walks `.thoughts/`, parses metadata, and filters by status/type. Fast and deterministic vs. asking the LLM to shell out and parse results.
- **Verification checks** -- `rpi verify` counts checkboxes, checks file coverage against git changes, and scans for TODO markers. Mechanical validation that should never consume context tokens.

Everything is embedded in a single binary via Go's `embed` package -- no external config repos, no dotfile dependencies. `rpi init` bootstraps any project from the binary alone.

## How It Compares

What sets RPI apart from other spec-driven development tools is the combination of two things: **reviewable artifacts that keep a human in the loop at every stage**, and a **compiled CLI that keeps mechanical work out of the LLM's context window**.

Every stage produces a document you can read, edit, reject, or share with your team before the next stage starts. The Go binary handles the bookkeeping (scaffolding, frontmatter, artifact linking, verification) so the LLM spends its tokens on thinking, not parsing.

**vs. [OpenSpec](https://github.com/Fission-AI/OpenSpec)** -- OpenSpec gives the AI more autonomy, implementing an entire plan in one pass. RPI gives you fine-grained control -- you review each implementation phase before it's executed, with git commits between phases for versioning and easy rollback.

**vs. unstructured prompting** -- Without stage boundaries, the LLM researches, designs, and implements in a single pass -- no checkpoints, no review, no way to course-correct before code is written.

## Project Structure

```
.
├── cmd/rpi/                              # rpi CLI binary (Go)
├── internal/
|   ├── workflow/assets/                  # All embedded assets (installed by rpi init)
|   |   ├── agents/                       # Agent definitions
|   |   ├── commands/                     # Slash command definitions (rpi-plan, rpi-research, etc.)
|   |   ├── skills/                       # Skill definitions (find-patterns, locate-codebase, etc.)
|   |   └── templates/                    # Scaffold + document templates (.tmpl, .template)
|   └── templates/                        # Template resolution with user-override support
├── docs/
|   ├── workflow-guide.md                 # Choosing Your Path (detailed examples)
|   ├── stages.md                         # How Each Stage Works (detailed)
|   ├── thoughts-directory.md             # .thoughts/ directory documentation
|   └── rpi-init.md                       # rpi init command documentation
```

## Acknowledgments

Inspired by [HumanLayer](https://github.com/humanlayer/humanlayer) -- their work on human-in-the-loop patterns for AI agents informed the design of this workflow.

## License

MIT
