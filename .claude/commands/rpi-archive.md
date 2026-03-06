---
description: Archive completed artifacts to keep .thoughts/ directory clean
model: sonnet
---

# Archive Artifacts

Move completed or superseded artifacts from `.thoughts/` to `.thoughts/archive/` to keep the active directory clean while preserving full history.

## Input

This command accepts two modes:

- **Specific paths**: `/rpi-archive .thoughts/research/2026-01-15-auth-flow.md` — archive specific artifacts
- **Scan mode**: `/rpi-archive` (no arguments) — scan for archive candidates

## Step 1: Identify Candidates

### If specific paths provided:

1. Read each file and check its frontmatter `status` field
2. If status is `draft` or `active`, warn immediately:
   ```
   Warning: This artifact is still [draft/active]:
   - .thoughts/research/2026-01-15-auth-flow.md (draft)

   Are you sure you want to archive it? This is unusual — draft/active artifacts are typically still in use.
   Please confirm explicitly: yes / no
   ```
3. If status is `complete` or `superseded`, proceed to Step 2

### If no paths provided (scan mode):

1. Scan these `.thoughts/` subdirectories for markdown files with YAML frontmatter:
   - `research/`
   - `designs/`
   - `structures/`
   - `plans/`
   - `tickets/`
   - `specs/`
2. Read the frontmatter of each file and extract the `status` field
3. Collect all artifacts with `status: complete` or `status: superseded`
4. Group candidates by directory type and present them:

```
Archive candidates:

Research (2):
- .thoughts/research/2026-01-15-auth-flow.md (complete)
- .thoughts/research/2026-02-01-api-patterns.md (complete)

Plans (1):
- .thoughts/plans/2026-02-10-add-rate-limiting.md (complete, all phases done)

Designs (1):
- .thoughts/designs/2026-01-20-caching-strategy.md (superseded by .thoughts/designs/2026-03-01-caching-v2.md)

Which would you like to archive? (all / specific items / none)
```

5. If no candidates found:
   ```
   No archive candidates found. All artifacts are either draft or active.
   ```

**For `superseded` artifacts**: Check for a `superseded_by` field in frontmatter and include it in the display so the user can see what replaced it.

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

1. Use Grep to search remaining (non-archived) `.thoughts/` files for references to the artifact's path
2. If references are found, warn:
   ```
   Cross-reference warning:

   .thoughts/research/2026-01-15-auth-flow.md is referenced by:
   - .thoughts/designs/2026-02-15-auth-redesign.md (line 12)
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

1. **Determine destination**: `.thoughts/archive/YYYY-MM/[type]/[filename]`
   - `YYYY-MM` is the current year-month (when the archive happens, not the artifact date)
   - `[type]` is the source subdirectory (research, designs, plans, etc.)
   - Example: `.thoughts/research/2026-01-15-auth-flow.md` → `.thoughts/archive/2026-03/research/2026-01-15-auth-flow.md`

2. **Create directory structure** if it doesn't exist:
   ```
   .thoughts/archive/
   └── 2026-03/
       ├── research/
       ├── designs/
       └── plans/
   ```

3. **Update frontmatter**: Before moving, update the artifact's `status` to `archived` and add an `archived_date` field:
   ```yaml
   status: archived
   archived_date: YYYY-MM-DD
   ```

4. **Move the file** to its archive destination

5. **Report results**:
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
