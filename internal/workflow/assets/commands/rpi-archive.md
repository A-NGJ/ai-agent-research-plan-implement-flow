---
description: Archive completed artifacts to keep .thoughts/ directory clean
model: sonnet
---

# Archive Artifacts

Move completed or superseded artifacts from `.thoughts/` to `.thoughts/archive/` to keep the active directory clean while preserving full history.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.
Run `rpi --help` to discover available commands and `rpi <command> --help` for detailed usage with examples.

## Input

This command accepts two modes:

- **Specific paths**: `/rpi-archive .thoughts/research/2026-01-15-auth-flow.md` — archive specific artifacts
- **Scan mode**: `/rpi-archive` (no arguments) — scan for archive candidates

## Step 1: Identify Candidates

### If specific paths provided:

1. Use `rpi` to read each artifact's current status from its frontmatter
2. If status is `draft` or `active`, warn immediately:
   ```
   Warning: This artifact is still [draft/active]:
   - .thoughts/research/2026-01-15-auth-flow.md (draft)

   Are you sure you want to archive it? This is unusual — draft/active artifacts are typically still in use.
   Please confirm explicitly: yes / no
   ```
3. If status is `complete` or `superseded`, proceed to Step 2

### If no paths provided (scan mode):

1. Use `rpi` to discover archivable artifacts with reference counts
2. The output returns candidates with `status: complete` or `status: superseded`, grouped by type, with `ref_count` and `superseded_by` for each.
3. Present the candidates:

```
Archive candidates:

Research (2):
- .thoughts/research/2026-01-15-auth-flow.md (complete)
- .thoughts/research/2026-02-01-api-patterns.md (complete)

Plans (1):
- .thoughts/plans/2026-02-10-add-rate-limiting.md (complete, all phases done)

Proposals (1):
- .thoughts/proposals/2026-01-20-caching-strategy.md (superseded by .thoughts/proposals/2026-03-01-caching-v2.md)

Which would you like to archive? (all / specific items / none)
```

4. If no candidates found:
   ```
   No archive candidates found. All artifacts are either draft or active.
   ```

## Step 2: Confirm Selection

Wait for the user to choose:

- **"all"**: Archive all candidates
- **Specific items**: User names specific files or numbers
- **"none"**: Cancel the operation

Never proceed without explicit confirmation. This is not optional.

## Step 3: Pre-Archive Checks

Before moving any files, perform these safety checks:

### Cross-Reference Check

For each artifact about to be archived:

1. Use `rpi` to check for active references to the artifact
2. If references are found, warn:
   ```
   Cross-reference warning:

   .thoughts/research/2026-01-15-auth-flow.md is referenced by:
   - .thoughts/proposals/2026-02-15-auth-redesign.md (line 12)
   - .thoughts/plans/2026-02-20-auth-plan.md (line 8)

   These references will become stale after archiving.
   Proceed anyway? (yes / no)
   ```

### Draft/Active Safety Gate

If any selected artifacts have `status: draft` or `status: active` (only possible when specific paths were provided):

```
Safety check: The following artifacts are still [draft/active]:
- .thoughts/plans/2026-03-01-wip-feature.md (draft)

Archiving draft/active artifacts is unusual. Are you absolutely sure? (yes / no)
```

This is a double confirmation — the user was already warned in Step 1 and must confirm again here.

## Step 4: Execute Archive

For each confirmed artifact:

1. Use `rpi` to archive the artifact (updates frontmatter and moves to archive/). Use the force option to skip the ref check warning if the user has already confirmed.
   This handles everything automatically:
   - Updates frontmatter (`status: archived`, `archived_date: YYYY-MM-DD`)
   - Creates the destination directory (`.thoughts/archive/YYYY-MM/[type]/`)
   - Moves the file

2. **Report results**:
   ```
   Archived 3 artifacts to .thoughts/archive/2026-03/:

   - research/2026-01-15-auth-flow.md
   - research/2026-02-01-api-patterns.md
   - designs/2026-01-20-caching-strategy.md
   ```

## Step 5: Specs Check (Optional)

After archiving, if `.thoughts/specs/` exists and contains spec files:

1. Check if any of the archived artifacts referenced spec files (search archived content for `.thoughts/specs/` paths)
2. If references found, prompt:
   ```
   The archived artifacts referenced these specs:
   - .thoughts/specs/auth.md
   - .thoughts/specs/api-endpoints.md

   Want me to verify these specs are still current?
   ```
3. If the user says yes, read each referenced spec and check if its content still matches the codebase (lightweight check — look at key file references, not exhaustive verification)

## Safety Rules

These rules are non-negotiable:

1. **Never auto-archive** — always present candidates and wait for explicit confirmation
2. **Never delete** — archived artifacts are moved, not deleted. They remain fully recoverable
3. **Draft/active double confirmation** — artifacts with `status: draft` or `status: active` require two explicit confirmations before archiving
4. **Cross-reference warnings** — always check for and warn about references from remaining active artifacts
5. **No bulk operations on draft/active** — if scan mode finds draft/active artifacts, exclude them from the candidate list entirely. They can only be archived via specific paths with double confirmation

## Guidelines

- Keep archive operations atomic — if something fails mid-archive, report what was archived and what wasn't
- Prefer archiving in batches by type (all research first, then designs, etc.) for cleaner output
- The archive directory is append-only — never modify or reorganize existing archived artifacts
- If an artifact has no `status` field in frontmatter, skip it and note: "Skipped [file] — no status field in frontmatter"
