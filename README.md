# AI Agent: Research-Propose-Plan-Implement Flow

A structured development workflow for AI coding agents that turns vague feature requests into shipped code through a pipeline of discrete, reviewable stages. Built for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [OpenCode](https://github.com/opencode-ai/opencode), but the underlying methodology works with any AI coding tool.

Instead of asking an AI to "just implement it" and hoping for the best, this workflow forces deliberate progression through **Research -> Propose -> Plan -> Implement**. Each stage produces a document you can review, edit, and approve before moving on.

```
Research -> Propose -> Plan -> Implement
   |          |        |        |
   v          v        v        v
.thoughts/  .thoughts/ .thoughts/ code +
research/   proposals/ plans/     tests +
                                  commits
```

## Why This Exists

AI coding assistants are powerful but unpredictable when given large tasks. They skip steps, make questionable architectural choices, and produce code that doesn't fit the codebase. This workflow solves that by:

- **Separating thinking from doing** -- Research gathers facts. Propose makes decisions with trade-offs. Plan specifies exact changes. Implement executes them.
- **Creating review checkpoints** -- You approve each stage before the next one starts. Bad decisions get caught early, not after 500 lines of wrong code.
- **Building persistent context** -- All artifacts live in `.thoughts/`, so you and your team (or the AI) can pick up where you left off across sessions.
- **Scaling to complexity** -- Simple bug fix? Skip straight to Plan -> Implement. Complex feature? Use Propose -> Plan -> Implement.
- **Keeping the context window small** -- LLMs produce better output when focused. By breaking work into stages, each conversation stays scoped to one job. The `.thoughts/` documents carry knowledge between stages, so the AI starts each stage with exactly the context it needs -- no more, no less.

## Quick Start

### Prerequisites

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) or [OpenCode](https://github.com/opencode-ai/opencode)
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
   rpi init /path/to/your/project                # Claude Code (default)
   rpi init /path/to/your/project --target opencode   # OpenCode
   ```

   This creates:
   - `.claude/` (or `.opencode/`) -- Agents, commands, skills, and templates
   - `.thoughts/` -- Directory for all pipeline artifacts (gitignored by default)
   - `.rpi/index.json` -- Codebase symbol index
   - `CLAUDE.md` (or `AGENTS.md`) -- Project-level instructions for the AI

4. Start your AI coding tool in the project and use the slash commands.

### The Slash Commands

| Command | What It Does | Output |
|---------|-------------|--------|
| `/rpi-research` | Investigates the codebase -- conversational fact-finding | Conversation (optionally `.thoughts/research/YYYY-MM-DD-topic.md`) |
| `/rpi-propose` | Investigates, analyzes, and proposes solutions with trade-offs | `.thoughts/proposals/YYYY-MM-DD-topic.md` |
| `/rpi-plan` | Creates phased implementation plan with success criteria | `.thoughts/plans/YYYY-MM-DD-topic.md` |
| `/rpi-implement` | Executes a plan phase-by-phase with verification | Code, tests, and commits |
| `/rpi-commit` | Creates focused git commits with smart grouping | Git commits |
| `/rpi-verify` | Validates implementation matches proposal artifacts | Verification report |
| `/rpi-archive` | Archives completed artifacts to keep `.thoughts/` clean | Moves files to `.thoughts/archive/` |

## Choosing Your Path

Not every task needs every stage. Match the path to your task's complexity:

- **Small tasks** (bug fixes, config changes) -- skip straight to **Plan -> Implement**. `/rpi-plan` does lightweight research on the fly.
- **Medium tasks** (focused features, single-concern changes) -- use **Propose -> Plan -> Implement**. Optionally run `/rpi-research` first if the codebase is unfamiliar.
- **Large tasks** (multi-concern features, major refactors) -- use **Propose -> Plan -> Implement**, where `/rpi-plan` decomposes the proposal into independently plannable units.

Not sure where to start? Use `/rpi-research` with any question -- it handles both focused investigation and open-ended research.

See the [full workflow guide](docs/workflow-guide.md) for detailed examples of each path.

## Documentation

- [Workflow Guide](docs/workflow-guide.md) -- Detailed examples of each path with tips
- [Stage Descriptions](docs/stages.md) -- How each command works, its modes, and output
- [`.thoughts/` Directory](docs/thoughts-directory.md) -- Artifact structure, naming, status lifecycle, and team sharing
- [`rpi init`](docs/rpi-init.md) -- CLI bootstrapping, flags, shell completion, and OpenCode support
- [Architecture](docs/architecture.md) -- Why a Go binary, CLI commands, and project structure

## How It Compares

What sets RPI apart from other spec-driven development tools is the combination of two things: **reviewable artifacts that keep a human in the loop at every stage**, and a **compiled CLI that keeps mechanical work out of the LLM's context window**.

Every stage produces a document you can read, edit, reject, or share with your team before the next stage starts. The Go binary handles the bookkeeping (scaffolding, frontmatter, artifact linking, verification) so the LLM spends its tokens on thinking, not parsing.

**vs. [OpenSpec](https://github.com/Fission-AI/OpenSpec)** -- OpenSpec gives the AI more autonomy, implementing an entire plan in one pass. RPI gives you fine-grained control -- you review each implementation phase before it's executed, with git commits between phases for versioning and easy rollback. RPI also gives you full ownership of every command and skill -- they're plain markdown files you can read, edit, and customize after `rpi init`. OpenSpec's prompts are compiled into its npm package and regenerated on `openspec update`, so the workflow logic stays inside the tool rather than in your project.

**vs. unstructured prompting** -- Without stage boundaries, the LLM researches, designs, and implements in a single pass -- no checkpoints, no review, no way to course-correct before code is written.

## Acknowledgments

Inspired by [HumanLayer](https://github.com/humanlayer/humanlayer) -- their work on human-in-the-loop patterns for AI agents informed the design of this workflow.

## License

MIT
