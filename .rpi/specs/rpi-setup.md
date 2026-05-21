---
feature: rpi-setup
description: Install or upgrade the rpi binary via the Claude Code plugin skill
tags: [spec]
---

# Spec: rpi-setup

## Scenarios

### Scenario 1: Fresh install on a clean system

Given the user has no `~/.rpi/bin/rpi` binary  
And no standalone rpi skills under `~/.claude/skills/`  
And no `rpi` key under `mcpServers` in `~/.claude/settings.json`  
When the user runs `/rpi:rpi-setup`  
Then the skill downloads and installs the rpi binary to `~/.rpi/bin/rpi`  
And reports the installed version  
And instructs the user to restart Claude Code

---

### Scenario 2: Re-run with existing plugin binary

Given `~/.rpi/bin/rpi` already exists and runs successfully  
When the user runs `/rpi:rpi-setup`  
Then the skill delegates to `rpi upgrade` without re-downloading  
And exits with the upgrade command's exit code

---

### Scenario 3: Standalone install detected via skills directory

Given one or more `rpi-*` directories exist under `~/.claude/skills/`  
When the user runs `/rpi:rpi-setup`  
Then the skill exits without installing  
And the error message names the conflicting directory path  
And offers `rpi uninstall --global` as the primary remediation  
And offers `~/.rpi/bin/rpi uninstall --global` as an alternative if rpi is not in PATH  
And offers a manual `rm -rf ~/.claude/skills/rpi-*` fallback  
And instructs the user to re-run `/rpi:rpi-setup` after cleanup

---

### Scenario 4: Standalone install detected via mcpServers entry

Given `~/.claude/settings.json` contains an `rpi` key under `mcpServers`  
And no `rpi-*` directories exist under `~/.claude/skills/`  
When the user runs `/rpi:rpi-setup`  
Then the skill exits without installing  
And the error message identifies the `mcpServers.rpi` entry as the conflict  
And offers `rpi uninstall --global` as the primary remediation  
And offers a manual JSON-edit instruction as an alternative  
And instructs the user to re-run `/rpi:rpi-setup` after cleanup

---

### Scenario 5: Plugin user with no standalone install (no false positive)

Given `~/.claude/settings.json` contains `rpi` only in `enabledPlugins` or `extraKnownMarketplaces`  
And no `rpi-*` directories exist under `~/.claude/skills/`  
And no `rpi` key exists under `mcpServers`  
When the user runs `/rpi:rpi-setup`  
Then the skill does not report a conflict  
And proceeds to install normally

---

### Scenario 6: Missing required dependency

Given `curl` or `jq` is not installed  
When the user runs `/rpi:rpi-setup`  
Then the skill exits without installing  
And the error message names the missing command  
And provides the package manager command to install it (`brew install jq curl` / `apt install jq curl`)  
And instructs the user to re-run `/rpi:rpi-setup` after installing

---

### Scenario 7: Checksum mismatch during download

Given the skill proceeds to the install step  
When the downloaded archive's SHA256 does not match `checksums.txt`  
Then the skill exits without writing to `~/.rpi/bin/`  
And reports the expected and observed checksums  
And leaves no partial files behind

---

### Scenario 8: Successful install ends with restart prompt

Given the install step completes without error  
When the binary version check succeeds  
Then the skill reports the installed version number  
And explicitly instructs the user to restart Claude Code for the MCP server to become active

---

## Constraints

- The skill never writes outside `~/.rpi/bin/`. No edits to shell rc files, `~/.claude/settings.json`, or PATH.
- Conflict detection checks both the skills directory and `mcpServers` independently; both conflicts are reported before exiting if both are present.
- The install step delegates entirely to `install.sh` piped via curl; the skill does not duplicate platform detection or checksum logic.
- The skill is safely re-runnable: repeated invocations on an already-installed system must never produce an error or partial state.

## Out of Scope

- Automatic conflict resolution (auto-removing standalone installs)
- Pinned-version installs (`--version` flag)
- Windows / WSL environments
- The `rpi upgrade` flow (owned by the binary, not the skill)
