# rpi — Claude Code plugin

Research → Propose → Plan → Implement → Verify. Spec-Driven Development for Claude Code, with persistent artifacts under `.rpi/`.

## Install

In Claude Code:

```
/plugin install rpi
```

Then run the one-step setup to fetch the matching `rpi` binary:

```
/rpi:rpi-setup
```

`/rpi:rpi-setup` downloads the release archive from `A-NGJ/rpi`, verifies it against the release's `checksums.txt`, and installs the binary to `~/.rpi/bin/rpi`. It writes nothing outside that directory. Re-running `/rpi:rpi-setup` upgrades the binary; no other state is modified.

## Conflict with a prior standalone install

If you previously installed RPI via `rpi init --global` (skills under `~/.claude/skills/rpi-*`, `rpi` MCP server registered, hooks/permissions in `~/.claude/settings.json`), the plugin's `/rpi:rpi-setup` will refuse to proceed. Remove the standalone install first:

```
rpi uninstall --global
```

Then re-run `/rpi:rpi-setup`.

## Skills

Skill folders keep the `rpi-` prefix so the trigger surface is unambiguous in Claude Code's slash-command picker (which may strip the plugin namespace when displaying suggestions, otherwise letting `/plan` collide with built-in commands). Triggers surface as `/rpi:rpi-<name>`:

| Standalone        | Plugin              |
| ----------------- | ------------------- |
| `/rpi-plan`       | `/rpi:rpi-plan`     |
| `/rpi-implement`  | `/rpi:rpi-implement`|
| `/rpi-verify`     | `/rpi:rpi-verify`   |
| `/rpi-propose`    | `/rpi:rpi-propose`  |
| `/rpi-research`   | `/rpi:rpi-research` |
| `/rpi-diagnose`   | `/rpi:rpi-diagnose` |
| `/rpi-commit`     | `/rpi:rpi-commit`   |
| `/rpi-archive`    | `/rpi:rpi-archive`  |
| `/rpi-explain`    | `/rpi:rpi-explain`  |
| `/rpi-handoff`    | `/rpi:rpi-handoff`  |
| `/rpi-spec-sync`  | `/rpi:rpi-spec-sync`|
| _(new)_           | `/rpi:rpi-setup`    |

The MCP server name (`rpi`) and tool prefix (`mcp__rpi__*`) are unchanged.

## Workflow context

The plugin's `SessionStart` hook injects the pipeline framing (skill list, `--ff`/`--grill` flag contract, `.rpi/` layout) into each new session's context. Nothing is written to `~/.claude/CLAUDE.md`, `~/.claude/settings.json`, or any project file.

## Project

- Source: <https://github.com/A-NGJ/rpi>
- Specs: `.rpi/specs/rpi-claude-plugin.md`, `.rpi/specs/rpi-skill-contract.md`
