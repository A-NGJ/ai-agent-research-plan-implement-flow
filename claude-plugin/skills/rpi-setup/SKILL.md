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

There are three Bash invocations. Run them in order; stop the flow if any prints a remediation message and exits non-zero.

Steps 3 through 9 must run **inside a single Bash invocation**: they share `$tmp` and the cleanup `trap`, neither of which survives across separate Bash tool calls. Splitting the install block into multiple Bash calls will leak partial downloads on failure.

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

### 3–9. Detect, fetch, verify, install (single Bash invocation)

The remaining work is one bash session so the temp dir, `$tmp`, and the cleanup `trap` survive across it. Run the entire block in one Bash tool call.

```sh
# 3a. Prerequisites — fail fast with install hints if anything is missing.
missing=
for cmd in curl jq tar awk install mktemp; do
  command -v "$cmd" >/dev/null 2>&1 || missing="$missing $cmd"
done
if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
  missing="$missing shasum-or-sha256sum"
fi
if [ -n "$missing" ]; then
  echo "Missing required commands:$missing"
  echo "Install them via your package manager (Homebrew: 'brew install jq curl'; Debian/Ubuntu: 'apt install jq curl coreutils'), then re-run /rpi:rpi-setup."
  exit 1
fi

# 3b. Platform detection.
uname_s=$(uname -s)
uname_m=$(uname -m)
case "$uname_s/$uname_m" in
  Darwin/arm64)  os=darwin; arch=arm64 ;;
  Darwin/x86_64) os=darwin; arch=amd64 ;;
  Linux/x86_64)  os=linux;  arch=amd64 ;;
  Linux/aarch64) os=linux;  arch=arm64 ;;
  *)
    echo "Unsupported platform: $uname_s/$uname_m. Supported: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64."
    exit 1
    ;;
esac
echo "Platform detected: $os/$arch."

# 4. Fetch release metadata.
echo "Fetching latest release metadata from GitHub..."
meta=$(curl -fsSL https://api.github.com/repos/A-NGJ/rpi/releases/latest) || {
  echo "Failed to fetch release metadata."; exit 1;
}
# goreleaser strips the leading "v" from the tag in the archive name.
tag=$(printf '%s' "$meta" | jq -r .tag_name)
version=${tag#v}
archive_name="rpi_${version}_${os}_${arch}.tar.gz"
archive_url=$(printf '%s' "$meta" | jq -r ".assets[] | select(.name == \"$archive_name\") | .browser_download_url")
checksums_url=$(printf '%s' "$meta" | jq -r '.assets[] | select(.name == "checksums.txt") | .browser_download_url')
if [ -z "$archive_url" ] || [ "$archive_url" = "null" ]; then
  echo "Release $tag has no asset named $archive_name. Aborting."
  exit 1
fi
echo "Latest release: $tag → $archive_name"

# 5. Download archive and checksums into a temp dir. The trap covers every
#    remaining exit path, including the SHA256 mismatch case.
tmp=$(mktemp -d) || { echo "mktemp failed."; exit 1; }
trap 'rm -rf "$tmp"' EXIT INT TERM
echo "Downloading $archive_name and checksums.txt into $tmp..."
curl -fsSL -o "$tmp/$archive_name"  "$archive_url"  || { echo "Download failed: $archive_url";  exit 1; }
curl -fsSL -o "$tmp/checksums.txt"  "$checksums_url" || { echo "Download failed: checksums.txt"; exit 1; }

# 6. Verify SHA256 — prefer shasum (macOS), fall back to sha256sum (Linux).
if command -v shasum >/dev/null 2>&1; then
  observed=$(shasum -a 256 "$tmp/$archive_name" | awk '{print $1}')
else
  observed=$(sha256sum "$tmp/$archive_name" | awk '{print $1}')
fi
expected=$(awk -v n="$archive_name" '$2 == n {print $1}' "$tmp/checksums.txt")
if [ -z "$expected" ]; then
  echo "Could not find $archive_name in checksums.txt. Aborting."
  exit 1
fi
if [ "$observed" != "$expected" ]; then
  echo "SHA256 mismatch for $archive_name:"
  echo "  expected: $expected"
  echo "  observed: $observed"
  echo "Aborting. No files installed."
  exit 1
fi
echo "SHA256 OK: $observed"

# 7. Extract and install.
echo "Extracting and installing to ~/.rpi/bin/rpi..."
mkdir -p "$HOME/.rpi/bin"
tar -xzf "$tmp/$archive_name" -C "$tmp"
install -m 0755 "$tmp/rpi" "$HOME/.rpi/bin/rpi"

# 8. PATH hint.
echo
echo "rpi binary installed at: $HOME/.rpi/bin/rpi"
echo "The plugin's MCP server points here directly — no PATH change required."
echo "To invoke 'rpi' from your shell, add this to your shell rc:"
echo "  export PATH=\"\$HOME/.rpi/bin:\$PATH\""

# 9. Confirm.
"$HOME/.rpi/bin/rpi" version || { echo "Installed binary did not run."; exit 1; }
echo "Setup complete."
echo
echo "IMPORTANT: Restart Claude Code (close and reopen this session) so the rpi"
echo "MCP server can launch. The MCP server tried to start at session start, but"
echo "the binary wasn't installed yet — it has to be re-launched, and Claude Code"
echo "has no live MCP reload, so a session restart is required."
echo "After restart, the RPI tools will be available."
```

After step 9 prints "Setup complete." plus the restart instruction, surface the restart message to the user explicitly in your reply — RPI tool calls will fail until the session is restarted, so the user needs to act on it before continuing with any RPI workflow.

## Behavioral guarantees

- After step 1 fires with `conflict=1`, no further steps run — the binary is not downloaded.
- After step 2 succeeds (delegated upgrade), the script returns the binary's own upgrade exit code; no files under the plugin directory are modified.
- Steps 3–9 run in one shell, so the `trap 'rm -rf "$tmp"' EXIT INT TERM` set in step 5 fires on every exit path, including the SHA256 mismatch case. Splitting steps 3–9 across separate Bash invocations breaks this guarantee.
- Step 3a fails fast with an actionable install hint if `jq`, `curl`, `tar`, `awk`, `install`, `mktemp`, or `shasum`/`sha256sum` is missing — nothing under `~/.rpi/` is touched in that case.
- The only file ever written outside `$tmp` is `$HOME/.rpi/bin/rpi`.
