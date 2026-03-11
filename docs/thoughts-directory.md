# The `.thoughts/` Directory

All pipeline artifacts live in `.thoughts/`, which is **gitignored by default**:

```
.thoughts/
├── PIPELINE.md          # This workflow reference guide
├── research/            # Codebase research documents
├── proposals/           # Solution proposals with trade-off analysis
├── plans/               # Implementation plans with checkboxes
├── specs/               # Living behavioral specs for modules/domains
├── prs/                 # PR descriptions
├── reviews/             # Code review and verification reports
└── archive/             # Completed artifacts (mirrors above structure)
```

Files follow the naming convention: `YYYY-MM-DD-descriptive-name.md` (specs use `domain-name.md` instead).

## Document Status Lifecycle

All pipeline artifacts use a `status` field in their YAML frontmatter to track progress:

```
draft -> active -> complete
```

- **`draft`** -- Initial state when a document is created (research, plans, tickets)
- **`active`** -- Work is in progress (e.g., `/rpi-implement` sets the plan to `active` when it starts executing)
- **`complete`** -- All work described in the document is finished

`/rpi-implement` manages plan status automatically. `/rpi-archive` warns before archiving documents that aren't yet `complete`.

## Specs (`/.thoughts/specs/`)

Specs are living documents that describe the **current behavior** of a module or domain -- not planned changes. They're created and updated as a byproduct of research and implementation:

- `/rpi-explore` can optionally create or update a spec when it documents a module's behavior comprehensively
- `/rpi-propose` can flag existing specs with `pending_changes` when a proposal will alter documented behavior
- Specs are updated to reflect reality *after* implementation, not during design

This directory serves as persistent context across sessions. You can read, edit, or delete any document. Claude will check for existing documents before creating new ones to avoid duplication.

## Sharing `.thoughts/` with Your Team

By default, `.thoughts/` is added to `.gitignore` -- useful when you want pipeline artifacts to stay local. But these documents can be valuable to the whole team: research captures institutional knowledge about the codebase, designs document *why* decisions were made, and plans provide a record of what was implemented and how.

To track `.thoughts/` in git, use the `--track-thoughts` flag during initialization:

```bash
rpi init --track-thoughts
```

This skips adding `.thoughts/` to `.gitignore`, so all pipeline artifacts get committed alongside your code.

**Why share with the team:**
- **Research documents** become searchable codebase documentation that stays current -- new team members can read them to understand how systems work instead of spelunking through code.
- **Proposal documents** preserve decision rationale. When someone asks "why did we use Redis instead of Memcached?", the answer is in `.thoughts/proposals/`, not lost in a Slack thread.
- **Plans** give visibility into how features were decomposed and implemented, making code review easier and providing a template for similar future work.
- **Any team member can pick up where another left off** -- if one person does the research and design, another can run `/rpi-plan` and `/rpi-implement` using those same documents.

**When to keep it local:**
- Exploratory or throwaway research you don't want cluttering the repo
- Plans for personal refactors or experiments
- When your team prefers to keep this context in other tools (Notion, Linear, etc.)

You can also take a hybrid approach: gitignore `.thoughts/` by default but selectively commit specific documents that have lasting value.
