---
name: rpi-archive
description: Move completed or superseded .rpi/ artifacts to .rpi/archive/ with reference checks and frontmatter updates. Use when user says 'archive these artifacts', 'clean up .rpi/', or 'archive completed work'.
---

# Archive Artifacts

## Goal

Move completed or superseded artifacts from `.rpi/` to `.rpi/archive/` to keep the active directory clean while preserving full history.

Two modes:
- **Specific paths** → archive named artifacts
- **No arguments** → scan for archive candidates (complete/superseded status)

## Invariants

- **Scan mode**: discover archivable artifacts, present candidates grouped by type with reference counts — never include draft/active in scan results
- **Specific paths**: check each artifact's status; warn immediately if draft/active
- Always present candidates and wait for explicit user confirmation before archiving — never auto-archive
- Check for cross-references from remaining active artifacts before moving — warn about stale references
- **Draft/active artifacts require double confirmation**: warn at identification AND again at pre-archive check
- Archive operation: update frontmatter (status → archived, add archived_date) and move to `.rpi/archive/YYYY-MM/[type]/`
- Never delete — archived artifacts are moved, not removed
- Report results: list what was archived and where
- After archiving, check if archived artifacts referenced specs — offer to verify those specs are still current

## Principles

- Keep operations atomic — if something fails mid-archive, report what succeeded and what didn't
- The archive directory is append-only — never modify existing archived artifacts
- Skip artifacts with no status field in frontmatter
