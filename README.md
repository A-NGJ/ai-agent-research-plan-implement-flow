![RPI Banner](docs/assets/rpi-banner.svg)

# AI Agent: Research-Propose-Plan-Implement Flow

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/A-NGJ/rpi)](https://github.com/A-NGJ/rpi/releases/latest)
[![CI](https://github.com/A-NGJ/rpi/actions/workflows/release.yml/badge.svg)](https://github.com/A-NGJ/rpi/actions/workflows/release.yml)

AI coding agents are capable -- the challenge is steering them. Without structure, you end up re-running prompts hoping the output lands closer to what you actually need. RPI gives you a framework to direct that capability: staged decisions, reviewable artifacts, and behavioral specs that keep work on track.

Each stage produces a document you can read, edit, and approve before the next one starts. A compiled Go CLI handles the bookkeeping so the LLM spends its tokens on thinking, not parsing. Built for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [OpenCode](https://github.com/opencode-ai/opencode), but the methodology works with any AI coding tool.

Best fit: solo devs and teams who want control over their AI dev flow — review where it matters, keep a durable trace of decisions, and split work across sessions without losing context.

## Quick Start

In Claude Code:

```
/plugin marketplace add A-NGJ/rpi
/plugin install rpi@rpi
/rpi:rpi-setup
```

These register the marketplace, install the plugin (skills, hooks, MCP server), and fetch the `rpi` binary into `~/.rpi/bin/rpi`. **First-time only: restart Claude Code after the third command** so the `rpi` MCP server can launch. For OpenCode, standalone, global, or from-source installs, see [Installation](#installation).

## What RPI helps with

| What's hard | RPI's answer |
|---|---|
| Pointing the agent at what's already in your codebase before it acts | `/rpi-research` (primary), `/rpi-propose` for tradeoff decisions |
| Keeping a durable source of truth for what the system is supposed to do | `.rpi/specs/` (living behavioral contracts) |
| Confirming the code matches intent, not just that it compiles | `/rpi-verify` (severity-classified spec conformance) |
| Reviewing big changes before they're written, not after | per-phase verification + checkboxes in `/rpi-implement` |
| Splitting work across sessions without re-explaining context | persistent `.rpi/` artifacts + `rpi resume` |

## See It in Action

You have a product with real data — and users keep asking, "can I just ask it questions in plain English?" Add an agentic workflow on top of what you already have, without rebuilding it:

```
/rpi-research What do we have today, and where could an agent plug in?
```
Brainstorm with Claude: it explores the data model, the query path, and the auth boundary, then surfaces the natural seams where an agent can attach. Optionally writes findings to `.rpi/research/`.

```
/rpi-propose Add an AI chat over our data
```
Crystallizes the brainstorm into 2-3 concrete approaches — each with tradeoffs grounded in what research found. You pick. Writes `.rpi/designs/` and a behavioral spec.

```
/rpi-plan .rpi/designs/2026-03-04-data-agent.md
```
Decomposes into phases — each with file changes, verification commands, and success criteria. Writes `.rpi/plans/`.

```
/rpi-implement .rpi/plans/2026-03-04-data-agent.md
```
Executes phase-by-phase: tests between each, commits as it goes, pauses only on manual verification or plan divergence.

## Try it

```
/rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings
```
Claude investigates, writes a phased plan to `.rpi/plans/`.

```
/rpi-implement .rpi/plans/2026-04-08-fix-date-formatter.md
```
Review the changes, approve, done. See the [full workflow guide](docs/workflow-guide.md) for more.

## The Slash Commands

| Command | What It Does | Output |
|---------|-------------|--------|
| `/rpi-research` | Investigates a question -- codebase or external, with conversational fact-finding | Conversation (optionally `.rpi/research/YYYY-MM-DD-topic.md`) |
| `/rpi-propose` | Investigates, analyzes, and designs solutions with trade-offs | `.rpi/designs/YYYY-MM-DD-topic.md` + `.rpi/specs/feature.md` |
| `/rpi-plan` | Creates phased implementation plan with success criteria | `.rpi/plans/YYYY-MM-DD-topic.md` |
| `/rpi-implement` | Executes a plan phase-by-phase with verification | Code, tests, and commits |
| `/rpi-commit` | Creates focused git commits with smart grouping | Git commits |
| `/rpi-verify` | Validates implementation matches design artifacts | `.rpi/reviews/YYYY-MM-DD-topic.md` |
| `/rpi-diagnose` | Iterative root-cause analysis for complex bugs | `.rpi/diagnoses/YYYY-MM-DD-topic.md` + fix |
| `/rpi-explain` | Diff-scoped walkthrough of an implemented solution | Conversation |
| `/rpi-spec-sync` | Syncs specs to match current codebase (detect drift, rewrite, rename, merge) | Updated `.rpi/specs/` |
| `/rpi-archive` | Archives completed artifacts to keep `.rpi/` clean | Moves files to `.rpi/archive/` |
| `/rpi-handoff` | Captures in-flight session context to a per-project temp file so the next session can resume | `/tmp/claude-handoff-<sha>.md` |

> **Modes.** `/rpi-propose` and `/rpi-plan` accept `--grill` (or "grill me on this") — hands off to the bundled [`grill-me`](https://github.com/mattpocock/skills) skill for adversarial, one-question-at-a-time interrogation of the draft. `/rpi-propose`, `/rpi-plan`, and `/rpi-implement` also accept `--ff` (fast-forward) — suppresses approval gates and auto-chains to `/rpi-verify`, stopping only on safety gates (codebase drift, sensitive content). Mutually exclusive with `--grill`.

## How RPI Is Different

RPI combines two things other tools don't: **reviewable artifacts that keep a human in the loop at every stage**, and a **compiled Go CLI that keeps mechanical work out of the LLM's context window**. The CLI handles scaffolding, frontmatter, artifact linking, and verification so the LLM focuses on the actual problem. Separating thinking from doing -- research gathers facts, propose makes decisions, plan specifies changes, implement executes them -- means review checkpoints catch bad decisions early, not after 500 lines of wrong code. All artifacts live in `.rpi/`, so context persists across sessions and teams. And by breaking work into stages, each conversation stays scoped to one job, keeping the context window small and output quality high.

**vs. [OpenSpec](https://github.com/Fission-AI/OpenSpec)** -- OpenSpec gives the AI more autonomy, implementing an entire plan in one pass. RPI gives you fine-grained control -- you review each implementation phase before it's executed, with git commits between phases for versioning and easy rollback. RPI also gives you full ownership of every command and skill -- they're plain markdown files you can read, edit, and customize after `rpi init`. OpenSpec's prompts are compiled into its npm package and regenerated on `openspec update`, so the workflow logic stays inside the tool rather than in your project.

**vs. unstructured prompting** -- Without stage boundaries, the LLM researches, designs, and implements in a single pass -- no checkpoints, no review, no way to course-correct before code is written.

## Choosing Your Path

Not every task needs every stage. Match the path to your task's complexity:

- **Small tasks** (bug fixes, config changes) -- skip straight to **Plan -> Implement**. `/rpi-plan` does lightweight research on the fly.
- **Medium tasks** (focused features, single-concern changes) -- use **Propose -> Plan -> Implement**. Optionally run `/rpi-research` first if the codebase is unfamiliar.
- **Large tasks** (multi-concern features, major refactors) -- use **Propose -> Plan -> Implement**, where `/rpi-plan` decomposes the proposal into independently plannable units.

Not sure where to start? Use `/rpi-research` with any question -- it handles both focused investigation and open-ended research. For complex bugs, use `/rpi-diagnose` to iterate on root-cause analysis.

Not sure what's in flight, or coming back after a break? Run `rpi status` for a dashboard of artifacts and progress, or `rpi resume` for a session-level overview — active/draft artifacts, the current phase of the most recent plan, and a suggested next action. Claude Code calls `rpi resume` automatically on session start so the assistant orients itself before you type.

<details>
<summary><code>rpi status</code> example output</summary>

![rpi status output](docs/assets/rpi-status.png)

</details>

See the [full workflow guide](docs/workflow-guide.md) for detailed examples of each path.

## Documentation

- [Workflow Guide](docs/workflow-guide.md) -- Detailed examples of each path with tips
- [Stage Descriptions](docs/stages.md) -- How each command works, its modes, and output
- [`.rpi/` Directory](docs/thoughts-directory.md) -- Artifact structure, naming, status lifecycle, and team sharing
- [`rpi init`](docs/rpi-init.md) -- CLI bootstrapping, flags, shell completion, and OpenCode support
- [Architecture](docs/architecture.md) -- Why a Go binary, CLI commands, and project structure
- [Semantic Search](docs/semantic-search.md) -- Optional qmd backend setup, warmup, status contract, and troubleshooting

## Installation

### Claude Code plugin (recommended)

In Claude Code:

```
/plugin marketplace add A-NGJ/rpi
/plugin install rpi@rpi
/rpi:rpi-setup
```

The first two register the marketplace and install the plugin (skills, hooks, MCP server). The third downloads the matching release, verifies its SHA256 against `checksums.txt`, and installs the binary at `~/.rpi/bin/rpi`. Re-running upgrades in place. The plugin writes nothing outside `~/.rpi/bin/`.

**Plugin command names.** Triggers are `/rpi:rpi-plan`, `/rpi:rpi-implement`, `/rpi:rpi-verify`, etc. — skill folders keep the `rpi-` prefix so Claude Code's namespace-stripping picker stays unambiguous. The MCP server name (`rpi`) and tool prefix (`mcp__rpi__*`) are unchanged.

If you previously installed via `rpi init --global`, run `rpi uninstall --global` before `/rpi:rpi-setup` — the plugin refuses to overwrite a standalone install. See the [plugin README](claude-plugin/README.md) for the full marketplace listing.

### OpenCode and standalone CLI

For OpenCode users or environments without the Claude Code plugin marketplace:

```bash
curl -sSfL https://raw.githubusercontent.com/A-NGJ/rpi/main/install.sh | bash
rpi init                                         # current directory, Claude Code
rpi init /path/to/project --target opencode      # OpenCode in a different directory
```

`rpi init` creates the Agent Skills directory (`.claude/` or `.opencode/`), `.rpi/` for pipeline artifacts, a project instructions file (`CLAUDE.md` or `AGENTS.md`), and (Claude Code only) MCP server registration.

```bash
VERSION=v0.3.0 curl -sSfL https://raw.githubusercontent.com/A-NGJ/rpi/main/install.sh | bash   # pin a version
rpi upgrade               # update the binary
rpi update                # sync project workflow files
rpi uninstall --global    # remove everything RPI owns (detects plugin- vs standalone-mode; user entries in settings.json preserved)
```

### One-time global setup

Optional, for users who work across many repos. Installs skills, agents, and the MCP server at the user level once:

```bash
rpi init --global                          # ~/.claude/
rpi init --global --target opencode        # ~/.config/opencode/
rpi update --global                        # refresh the global install
```

`--global` writes only to your user config directory — no `.rpi/`, `CLAUDE.md`, or `.gitignore` is touched. See [docs/rpi-init.md](docs/rpi-init.md) for the full layout.

### From source

Requires Go 1.25+. Either install directly or clone and build:

```bash
go install github.com/A-NGJ/rpi/cmd/rpi@latest                              # direct
git clone https://github.com/A-NGJ/rpi && cd rpi && make install            # clone + build (copies to ~/.local/bin)
```

Make sure `~/.local/bin` (or your chosen install dir) is in your PATH.

### Upgrading

```bash
rpi upgrade    # download and install the latest release
```
Plugin users can re-run `/rpi:rpi-setup` (which delegates to `rpi upgrade`).

## MCP Server

The `rpi` binary doubles as an [MCP](https://modelcontextprotocol.io/) server. Running `rpi serve` starts a stdio-based server that exposes all CLI operations as typed tools (`rpi_scaffold`, `rpi_scan`, `rpi_chain`, `rpi_frontmatter_get`, etc.). AI assistants call these tools with validated JSON schemas instead of constructing shell commands.

`rpi init` auto-registers the MCP server with Claude Code when both `rpi` and `claude` are in your PATH. Use `--no-mcp` to skip this. See [Architecture](docs/architecture.md) for details.

## Optional: Semantic Search

`rpi serve` exposes an `rpi_search` MCP tool that returns ranked, semantically relevant `.rpi/` artifacts for a natural-language query. When the optional [qmd](https://github.com/tobi/qmd) backend is installed, seven skills (`rpi-research`, `rpi-propose`, `rpi-plan`, `rpi-verify`, `rpi-explain`, `rpi-diagnose`, `rpi-spec-sync`) call it automatically before producing artifacts. Without qmd, the same skills fall back to keyword scan — RPI ships fully functional without it.

See [docs/semantic-search.md](docs/semantic-search.md) for installation, warmup, the status contract, and troubleshooting.

## Acknowledgments

Inspired by [HumanLayer](https://github.com/humanlayer/humanlayer) — their work on human-in-the-loop patterns for AI agents informed this workflow. RPI bundles the [`grill-me`](https://github.com/mattpocock/skills/blob/main/skills/productivity/grill-me/SKILL.md) skill by Matt Pocock under MIT; see [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for the full attribution.

## License

MIT
