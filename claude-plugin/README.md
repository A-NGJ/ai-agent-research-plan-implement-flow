# rpi — Claude Code plugin

Research → Propose → Plan → Implement → Verify. Spec-Driven Development for Claude Code, with persistent artifacts under `.rpi/`.

## Install

In Claude Code:

```
/plugin install rpi
```

Then run the one-step setup to fetch the matching `rpi` binary:

```
/rpi:setup
```

`/rpi:setup` downloads the release archive from `A-NGJ/rpi`, verifies it against the release's `checksums.txt`, and installs the binary to `~/.rpi/bin/rpi`. It writes nothing outside that directory. Re-running `/rpi:setup` upgrades the binary; no other state is modified.

## Conflict with a prior standalone install

If you previously installed RPI via `rpi init --global` (skills under `~/.claude/skills/rpi-*`, `rpi` MCP server registered, hooks/permissions in `~/.claude/settings.json`), the plugin's `/rpi:setup` will refuse to proceed. Remove the standalone install first:

```
rpi uninstall --global
```

Then re-run `/rpi:setup`.

## Skills

All skills live under the `rpi:` namespace. Users migrating from the standalone install should note the command rename:

| Standalone   | Plugin         |
| ------------ | -------------- |
| `/rpi-plan`     | `/rpi:plan`       |
| `/rpi-implement`| `/rpi:implement`  |
| `/rpi-verify`   | `/rpi:verify`     |
| `/rpi-propose`  | `/rpi:propose`    |
| `/rpi-research` | `/rpi:research`   |
| `/rpi-diagnose` | `/rpi:diagnose`   |
| `/rpi-commit`   | `/rpi:commit`     |
| `/rpi-archive`  | `/rpi:archive`    |
| `/rpi-explain`  | `/rpi:explain`    |
| `/rpi-handoff`  | `/rpi:handoff`    |
| `/rpi-spec-sync`| `/rpi:spec-sync`  |
| _(new)_      | `/rpi:setup`      |

The MCP server name (`rpi`) and tool prefix (`mcp__rpi__*`) are unchanged.

## Workflow context

The plugin's `SessionStart` hook injects the pipeline framing (skill list, `--ff`/`--grill` flag contract, `.rpi/` layout) into each new session's context. Nothing is written to `~/.claude/CLAUDE.md`, `~/.claude/settings.json`, or any project file.

## Project

- Source: <https://github.com/A-NGJ/rpi>
- Specs: `.rpi/specs/rpi-claude-plugin.md`, `.rpi/specs/rpi-skill-contract.md`
