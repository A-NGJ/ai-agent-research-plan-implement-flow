---
description: Create scoped tickets from design docs to break complex features into plannable units. Use whenever a design doc covers multiple concerns or the work is too large for a single /create-plan pass — even if the user doesn't say "tickets" explicitly, suggest this when a design would produce an unwieldy plan.
model: opus
---

# Create Tickets

Break down a design document (and optional structure document) into discrete, scoped tickets. Each ticket captures a bounded piece of work that can be independently planned with `/create-plan`.

This is part of the pipeline: **research → design → [structure] → tickets → plan → implement**. Use this when a design covers enough complexity that planning it as a single unit would produce an unwieldy plan. Tickets give you control over implementation order, scope, and parallelism.

**When to use tickets vs. straight to plan:**
- **Skip tickets** when the design is focused enough that `/create-plan` can handle it directly (1-3 phases, single concern)
- **Use tickets** when the design covers multiple independent concerns, the work spans many files across different subsystems, or you want to implement in stages over multiple sessions

## Initial Response

1. **If a design doc path is provided** → Read it and proceed to decomposition
2. **If a design doc + structure doc provided** → Read both and proceed
3. **If nothing provided** → Ask:
   ```
   I'll help you break down a design into implementation tickets.

   Please provide:
   1. Path to the design document
   2. (Optional) Path to the structure document

   Example: `/create-tickets .thoughts/designs/2025-01-08-feature-name.md`
   ```

---

## Step 1: Understand the Design

1. **Read all provided documents fully** — design doc, structure doc if available, and any linked research
2. **Identify the design's key components:**
   - What are the distinct subsystems or concerns?
   - What are the natural boundaries? (data model, business logic, API layer, UI, infrastructure)
   - What depends on what?
3. **Look for decomposition signals in the design:**
   - Separate design decisions that affect different parts of the system
   - The file structure section (or structure doc) — groups of files that change together
   - Integration points — these often mark ticket boundaries
   - Any sequencing hints the design mentions

## Step 2: Propose Ticket Breakdown

Present the proposed decomposition to the user before writing anything:

```
Based on the design, I'd break this into [N] tickets:

1. **[Short title]** — [1-line scope description]
   Files: [key files involved]
   Depends on: nothing (foundation)

2. **[Short title]** — [1-line scope description]
   Files: [key files involved]
   Depends on: #1

3. **[Short title]** — [1-line scope description]
   Files: [key files involved]
   Depends on: #1, #2

Dependency graph:
#1 ──→ #2 ──→ #3
  └──────────→ #4

Does this breakdown make sense? Want to split, merge, or reorder any tickets?
```

**Decomposition principles:**
- Each ticket should be implementable in a single `/create-plan` → `/implement-plan` cycle
- Tickets should minimize cross-ticket dependencies (prefer a DAG with few edges)
- The first ticket(s) should lay foundation that others build on (data models, core types, shared interfaces)
- Each ticket should leave the codebase in a working state when complete
- A ticket should cover one concern — don't mix unrelated changes
- If a ticket would result in more than ~4 plan phases, consider splitting it further

## Step 3: Write the Tickets

After the user confirms (or iterates on) the breakdown, write each ticket to `.thoughts/tickets/`.

**Filename format**: `prefix-NNN-descriptive-name.md`

Derive the prefix from the design doc's topic — a short (2-5 char) abbreviation that groups related tickets. For example, a design about "user authentication" gets prefix `auth`, a design about "notification system" gets prefix `notif`. This prevents collisions when multiple designs generate tickets, and makes it easy to see which tickets belong together.

Sequential numbering (`001`, `002`, `003`) reflects suggested implementation order. If implementation order is flexible for some tickets, note that in the dependency field.

### Ticket Template

````markdown
---
ticket: prefix-NNN
title: "[Short Descriptive Title]"
status: open
depends_on: []          # ticket IDs this blocks on, e.g. ["auth-001"]
design: "[path to design doc]"
structure: "[path to structure doc, or omit]"
tags: [relevant-component-names]
---

# [Title]

## Summary

[2-4 sentences: what this ticket accomplishes and why it matters in the context of the broader feature. Include enough context that someone reading just this ticket understands the goal without needing to read the full design doc.]

## Scope

### In
- [Concrete deliverable 1 — files, components, or behaviors]
- [Concrete deliverable 2]

### Out
- [Explicitly excluded items that someone might expect to be here]
- [Items covered by other tickets — reference ticket ID]

## Design Context

[Summarize the relevant design decisions from the design doc that apply to this ticket. Include the decision, the chosen approach, and the key rationale. Don't just link to the design doc — extract the parts that matter for this scope so `/create-plan` has everything it needs.]

### Key Interfaces (if applicable)
[Interfaces this ticket must implement or respect — from the design or structure doc. Include signatures or shapes so the implementer doesn't need to reference another doc.]

## Acceptance Criteria

- [ ] [Specific, verifiable criterion]
- [ ] [Another criterion]
- [ ] [Tests: what should be tested]

## Files

[List the primary files this ticket touches — new and modified. Pull from the design doc's file structure section or the structure doc.]

- `path/to/new-file.ext` (new) — [responsibility]
- `path/to/existing-file.ext` (modify) — [what changes]

## Notes

[Implementation hints, gotchas, edge cases to watch for — anything that would save the implementer time. Keep it brief. Omit this section if there's nothing non-obvious to say.]

## References
- Design: `[path]` — [relevant sections]
- Structure: `[path]` (if applicable)
- Related tickets: [ticket IDs]
````

### What makes a good ticket

The ticket template is a tool, not a form to fill out mechanically. The purpose is to give `/create-plan` everything it needs to produce a focused, accurate plan. That means:

- **The Summary should stand alone.** Someone reading just this ticket should understand what they're building and why, without opening the design doc. Copy the relevant context — don't just link.
- **Design Context is the most important section.** This is where you extract the design decisions, chosen approaches, and interface contracts that are relevant to this ticket's scope. The implementer will rely on this to make the right calls. Be specific — include code shapes, data structures, and patterns from the design doc.
- **Acceptance Criteria should be verifiable.** Each criterion should be checkable with a test, a command, or a specific observable behavior. "Works correctly" is not a criterion; "returns 404 for unknown user IDs" is.
- **Scope boundaries prevent ticket creep.** The "Out" section is just as important as "In" — it tells the implementer where to stop and which adjacent concerns belong to other tickets.

## Step 4: Write the Index

After writing all tickets, create (or update) an index file at `.thoughts/tickets/index.md`:

```markdown
# Tickets: [Feature Name]

Source: `[path to design doc]`
Created: [YYYY-MM-DD]

## Tickets

| # | Title | Status | Depends On |
|---|-------|--------|------------|
| prefix-001 | [Title] | open | — |
| prefix-002 | [Title] | open | prefix-001 |
| prefix-003 | [Title] | open | prefix-001, prefix-002 |

## Dependency Graph

```
[ASCII diagram showing ticket dependencies]
```

## Implementation Order

[Recommended sequence, noting where parallelism is possible]
```

If an `index.md` already exists with tickets from another design, append a new section rather than overwriting.

## Step 5: Present Summary

```
Created [N] tickets in `.thoughts/tickets/`:

1. `prefix-001-description.md` — [title]
2. `prefix-002-description.md` — [title]
3. `prefix-003-description.md` — [title]

Index: `.thoughts/tickets/index.md`

Recommended implementation order: #prefix-001 → #prefix-002 → #prefix-003
(#prefix-002 and #prefix-003 can run in parallel after #prefix-001)

To plan a specific ticket:
  /create-plan .thoughts/tickets/prefix-001-description.md
```

---

## Guidelines

1. **Self-contained tickets** — Each ticket should include enough context from the design doc that `/create-plan` can produce a good plan without reading the full design. Extract, don't just link.
2. **Right-sized scope** — Too big and you lose the benefit of decomposition. Too small and you create overhead. A good ticket is roughly one PR / one focused work session.
3. **Clear boundaries** — "In" and "Out" should eliminate ambiguity about where one ticket's work ends and another's begins.
4. **Dependency honesty** — Don't over-constrain with unnecessary dependencies, but don't hide real ones. If ticket 3 can't start until ticket 2 is done, say so. If they can run in parallel, make that clear.
5. **Respect the design** — Tickets decompose a design into implementation units; they don't redesign. If the decomposition reveals a design issue, flag it to the user rather than silently working around it.
