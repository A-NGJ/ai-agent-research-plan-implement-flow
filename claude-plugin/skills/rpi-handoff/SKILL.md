---
name: rpi-handoff
description: Capture in-flight conversation context to a deterministic per-project temp file at the end of a session, so the next session in the same project can be told to read and consume it. Use when user says 'handoff', 'end of session', 'save session context for next time', 'wrap this up for the next session', or 'hand off to the next session'.
argument-hint: "[focus for next session]"
---

# Session Handoff

## Goal

Capture in-flight conversation context — current task, open questions, files touched, suggested next steps — to a deterministic per-project temp file at the end of a session, so the next session in the same project can pick up the thread without re-discovery.

This skill writes only. The SessionStart hook installed by project setup is what nudges the next session to read and remove the file (consume-once). There is no separate pickup-side skill — the hook plus existing read and shell tools are sufficient.

## Invariants

- **Path recipe (load-bearing — keep in lockstep with the SessionStart `claude-handoff` hook entry in `cmd/rpi/init_cmd.go`; the test `TestSessionHandoffHookRecipePinned` enforces this):**
  ```
  /tmp/claude-handoff-$(echo -n "$PWD" | shasum -a 256 | cut -c1-12).md
  ```
  shasum -a 256 is available on macOS and most Linux distros out of the box. The first 12 hex chars give a 2⁴⁸ namespace — collision risk between distinct project paths on one machine is effectively zero.
- If a handoff file already exists at the computed path, print a one-line warning to the user before writing:
  ```
  Existing handoff at <path> will be overwritten — proceeding
  ```
  The warning is user-visible by design — the user can interrupt before the write.
- Run `rpi resume` to capture the current RPI-derived session state, and embed its output **verbatim** under the `## RPI session state` section of the handoff body. Embedding (not just referencing the command) is required so the next session sees the snapshot the previous session signed off on, even if `.rpi/` artifacts drift between capture and pickup.
- Reference `.rpi/` artifacts (plans, designs, specs, research) by path under `## References` — never duplicate their bodies. The next session reads them itself.
- On a single-user workstation the sensitive-content risk is the user's responsibility; if the conversation visibly touched secrets, internal URLs, or other sensitive content, call it out before writing (mirrors the commit skill).
- Write the handoff via the Write tool. After writing, confirm with one line:
  ```
  Handoff written to <path>
  ```
  No tail summary, no ceremony.

## Body sections

Compose the handoff body in this order. Omit `## Focus for next session` when `$ARGUMENTS` is empty; omit `## Pending uncommitted work` when the working tree is clean.

1. `## Focus for next session` — populated from `$ARGUMENTS` when invoked with a focus argument; the rest of the body is framed around this.
2. `## Current task` — one paragraph: what the session was doing and where it stopped.
3. `## Open questions / unresolved decisions` — bullets the user still needs to resolve.
4. `## Files touched` — paths plus key line refs from recent edits, so the next session resumes without re-discovery.
5. `## Pending uncommitted work` — only when the working tree is dirty. Flag it and suggest the commit skill for the next session.
6. `## References` — `.rpi/` artifact paths mentioned in the session, by path only.
7. `## Suggested skills for next session` — e.g. plan, implement, verify, based on what was in flight.
8. `## RPI session state` — the embedded `rpi resume` output, verbatim.

## Pickup

This skill writes only. Pickup is handled automatically by the SessionStart hook entry that project setup installs in `.claude/settings.json` (marker `claude-handoff`). The hook checks for the file at the same per-`$PWD` path and, if present, emits a one-line nudge telling the next session to read the file and remove it after reading. The cleanup is allowlisted via `Bash(rm /tmp/claude-handoff-*.md)`, so it runs without a permission prompt.

If the next-session pickup ever stops working, first check that the path recipe in this skill matches the recipe in the `claude-handoff` hook entry — they must stay byte-identical.
