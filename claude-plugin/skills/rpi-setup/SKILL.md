---
name: rpi-setup
description: Install or upgrade the rpi binary into ~/.rpi/bin/rpi for the Claude Code plugin. Use when user says '/rpi:rpi-setup', 'install rpi', 'set up the rpi plugin binary', or after first installing the rpi plugin.
---

# RPI Setup

## Goal

Install the `rpi` binary so the plugin's MCP server (`~/.rpi/bin/rpi serve`) and CLI subcommands work. Safely re-runnable — when the binary is already present, delegate to `rpi upgrade`. Refuse to proceed when a prior standalone install is detected and surface the remediation command.

The plugin writes the binary only to `~/.rpi/bin/rpi`. It does not touch `~/.local/bin`, shell rc files, PATH, or any other location.

## Invariants

- Download binaries only from official GitHub Releases under `A-NGJ/rpi`. No mirrors.
- Verify every download's SHA256 against the release's `checksums.txt` asset. On mismatch: delete the temp dir and abort with a clear error naming the expected and observed checksums. Leave no partial files at `~/.rpi/bin/`.
- The only filesystem mutation outside `~/.rpi/` is reading from network and writing to `mktemp -d` (which is cleaned up before return).
- Print what each step is about to do before executing.
- Never `sudo`. Never write outside `~/.rpi/`. Never modify the user's PATH.
- On any failure mid-flight: clean up the temp dir and exit non-zero with a clear message.

## Steps

There are four invocations. Run them in order; stop the flow if any prints a remediation message and exits non-zero.

**Execution preference**, in order of decreasing preference:
1. **Bash tool** — the natural choice. Step 3 is a single short pipe, so output is small and unlikely to trip any PreToolUse hook that restricts large-output Bash calls.
2. **A single sandboxed-shell call** (if your environment provides one and it intercepts the Bash tool) — run step 3's whole pipe in **one** call so `install.sh`'s internal `trap` and `$tmp` survive. Never split the pipe across multiple calls.
3. **User runs it with `!` prefix** — only when neither of the above is available. Surface the exact one-liner from step 3 to the user, prefixed with `!`, and ask them to paste it into the Claude Code prompt.

### 1. Detect a prior standalone install

A standalone install — created by `rpi init --global` — leaves skills under `~/.claude/skills/rpi-*` and registers an `rpi` MCP server in `~/.claude/settings.json` outside the plugin. If detected, refuse and instruct the user to clean up first.

This step requires `jq` so the MCP-server check looks at `.mcpServers.rpi` specifically rather than grepping for the bare string `"rpi"` (which false-positives on `extraKnownMarketplaces.rpi` whenever the user has run `/plugin marketplace add A-NGJ/rpi`).

```sh
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for /rpi:rpi-setup. Install via your package manager (Homebrew: 'brew install jq'; Debian/Ubuntu: 'apt install jq'), then re-run /rpi:rpi-setup."
  exit 1
fi

conflict=0
if ls -d "$HOME/.claude/skills/rpi-"* >/dev/null 2>&1; then conflict=1; fi
if [ -f "$HOME/.claude/settings.json" ] && jq -e '.mcpServers.rpi // empty' "$HOME/.claude/settings.json" >/dev/null 2>&1; then
  # Look at the mcpServers.rpi key specifically — a plain grep for "rpi"
  # would also match marketplace entries (extraKnownMarketplaces.rpi) and
  # any other place the string happens to appear.
  conflict=1
fi
if [ "$conflict" = 1 ]; then
  echo "Standalone rpi install detected (skills under ~/.claude/skills/rpi-* or mcpServers.rpi entry in ~/.claude/settings.json)."
  echo "Run 'rpi uninstall --global' to remove the standalone install, then re-run /rpi:rpi-setup."
  exit 1
fi
```

### 2. Detect an existing plugin-mode install → delegate to `rpi upgrade`

If `~/.rpi/bin/rpi` already exists and works, this is a re-run. Hand off to the binary's own upgrade flow.

```sh
if [ -x "$HOME/.rpi/bin/rpi" ] && "$HOME/.rpi/bin/rpi" version >/dev/null 2>&1; then
  echo "Existing rpi binary found at ~/.rpi/bin/rpi; delegating to 'rpi upgrade'..."
  "$HOME/.rpi/bin/rpi" upgrade
  exit $?
fi
```

### 3. Install (delegate to the project's `install.sh`)

The repo already ships an `install.sh` that handles platform detection, version resolution, download, SHA256 verification, temp-dir cleanup via `trap`, and installation. Honour its `INSTALL_DIR` env var to target the plugin's well-known location (`~/.rpi/bin`) instead of the standalone default. One pipe; the script does the rest.

```sh
INSTALL_DIR="$HOME/.rpi/bin" bash -c 'curl -sSfL https://raw.githubusercontent.com/A-NGJ/rpi/main/install.sh | bash'
```

Confirm it landed:

```sh
"$HOME/.rpi/bin/rpi" version
```

If either command exits non-zero, surface the message to the user verbatim — `install.sh` prints actionable errors (`Checksum mismatch`, `No sha256sum or shasum found`, `Unsupported OS`, etc.). Don't retry blindly.

### 4. Tell the user to restart Claude Code

After the install succeeds, surface this message in your reply (do not skip it):

> **Restart Claude Code** (close and reopen this session) so the `rpi` MCP server can launch. It tried to start at the previous session start, but the binary wasn't installed yet — Claude Code has no live MCP reload, so a session restart is required. After restart, the RPI tools are available.

## Behavioral guarantees

- After step 1 fires with `conflict=1`, no further steps run — the binary is not downloaded.
- After step 2 succeeds (delegated upgrade), the script returns the binary's own upgrade exit code; no files under the plugin directory are modified.
- Step 3 delegates to `install.sh`, which downloads to its own `mktemp -d` and registers an `EXIT` trap inside its own shell — every exit path (including SHA256 mismatch) cleans up the temp dir. The only file ever written outside `$tmp` is `$HOME/.rpi/bin/rpi`.
- `install.sh` only downloads binaries from official `A-NGJ/rpi` GitHub Releases. The plugin never points elsewhere.
