# The `.rpi/` Directory

All pipeline artifacts live in `.rpi/`, which is **tracked in git by default** (use `--no-track` during init to gitignore it instead):

```
.rpi/
├── research/            # Research notes (codebase or external)
├── designs/             # Solution designs with trade-off analysis
├── diagnoses/           # Bug diagnosis post-mortems
├── plans/               # Implementation plans with checkboxes
├── specs/               # Living behavioral specs for modules/domains
├── reviews/             # Code review and verification reports
├── templates/           # Scaffold templates (user-overridable)
└── archive/             # Completed artifacts (mirrors above structure)
```

Files follow the naming convention: `YYYY-MM-DD-descriptive-name.md` (specs use `domain-name.md` instead).

## Document Status Lifecycle

All pipeline artifacts use a `status` field in their YAML frontmatter to track progress:

```
draft -> active -> complete -> archived
              \        \         |
               \        \-> active (reopen)
                \-> superseded
                     (from any state)
```

- **`draft`** -- Initial state when a document is created (research, plans, tickets). Transitions to `active` or `superseded`.
- **`active`** -- Work is in progress (e.g., `/rpi-implement` sets the plan to `active` when it starts executing). Transitions to `complete` or `superseded`.
- **`complete`** -- All work described in the document is finished. Can transition to `active` (reopen), `archived`, or `superseded`.
- **`archived`** -- Moved to `.rpi/archive/` for long-term storage.
- **`superseded`** -- Replaced by a newer artifact. Reachable from `draft`, `active`, or `complete`.

`/rpi-implement` manages plan status automatically. `/rpi-archive` warns before archiving documents that aren't yet `complete`.

## Specs (`.rpi/specs/`)

Specs are living documents that describe the **current behavior** of a module or domain -- not planned changes. Each spec contains:

- **Purpose** -- what the feature does and why (1-3 sentences)
- **Scenarios** -- 5-8 Given/When/Then scenarios describing user-observable behavior. Scenarios must not reference internal structure (structs, file paths, function names) -- they describe what the user sees, not how it's built.
- **Constraints** -- boundaries and invariants
- **Out of Scope** -- what the spec intentionally does not cover

Specs are created and updated as a byproduct of research and implementation:

- `/rpi-propose` creates a spec alongside each design, with scenarios as the behavioral contract
- `/rpi-propose` can flag existing specs with `pending_changes` when a proposal will alter documented behavior
- `/rpi-verify` and `/rpi-implement` check specs by extracting scenarios and verifying each against the implementation
- Specs are updated to reflect reality *after* implementation, not during design

This directory serves as persistent context across sessions. You can read, edit, or delete any document. Claude will check for existing documents before creating new ones to avoid duplication.

## Sharing `.rpi/` with Your Team

By default, `.rpi/` is tracked in git -- research captures institutional knowledge about the codebase, designs document *why* decisions were made, and plans provide a record of what was implemented and how.

If you want pipeline artifacts to stay local, use the `--no-track` flag during initialization:

```bash
rpi init --no-track
```

This adds `.rpi/` to `.gitignore`, so pipeline artifacts stay local to your machine.

**Why share with the team:**
- **Research documents** become searchable codebase documentation that stays current -- new team members can read them to understand how systems work instead of spelunking through code.
- **Design documents** preserve decision rationale. When someone asks "why did we use Redis instead of Memcached?", the answer is in `.rpi/designs/`, not lost in a Slack thread.
- **Plans** give visibility into how features were decomposed and implemented, making code review easier and providing a template for similar future work.
- **Any team member can pick up where another left off** -- if one person does the research and design, another can run `/rpi-plan` and `/rpi-implement` using those same documents.

**When to keep it local:**
- Exploratory or throwaway research you don't want cluttering the repo
- Plans for personal refactors or experiments
- When your team prefers to keep this context in other tools (Notion, Linear, etc.)

You can also take a hybrid approach: gitignore `.rpi/` by default but selectively commit specific documents that have lasting value.
