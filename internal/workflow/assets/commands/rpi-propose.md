---
description: Investigate, analyze, and propose solutions — from quick decisions to complex features
model: opus
---

# Solution Proposal

Investigate the codebase, analyze trade-offs, and produce a proposal document that captures what we learned, what we decided, and why. This merges investigation, design, and structural planning into a single command with human checkpoints at every meaningful juncture.

This is part of the pipeline: **research → propose → plan → implement**. Propose is where the hard choices happen — understanding the terrain, weighing options, and committing to an approach. The output is a proposal document that Plan consumes directly.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.

## Initial Response

**Auto-detect the mode from what's provided:**

- **Plain text description of a focused decision** (e.g., "should we use X or Y?", "design the caching approach") → **Quick mode**
- **Path to research doc, or complex feature description** → **Full mode**
- **Path to an existing proposal doc** → **Incremental mode** (updating a previous proposal)
- **Nothing provided** → Ask:
  ```
  I'll help you investigate and propose a solution.

  You can use this in several ways:

  **Quick decision** (focused):
  `/rpi-propose should we use Redis or Memcached for session caching?`

  **Complex feature** (full investigation):
  `/rpi-propose add real-time notifications to the dashboard`

  **From prior exploration**:
  `/rpi-propose .thoughts/research/2026-03-10-notifications.md`

  **Update existing proposal**:
  `/rpi-propose .thoughts/proposals/2026-03-10-notifications.md`
  ```

---

## Quick Mode

For focused decisions that don't warrant a full investigation — choosing between two approaches, designing a single component's interface, deciding on a data model shape.

### Step 1: Quick Context

1. **Read any mentioned files fully**
2. **Do proportional research** — scale to the decision's scope:
   - If the decision is localized (one module, one pattern): read the relevant files directly
   - If it touches multiple areas: spawn 1-2 targeted sub-tasks to locate relevant files and find existing patterns
3. **Present the decision frame:**
   ```
   Here's what I understand:

   The decision: [what needs to be decided]
   Context: [relevant codebase state, with file:line refs]
   Constraints: [what limits our options]

   Let me explore the options.
   ```
   If the user already specified the options, skip straight to analysis.

### Step 2: Trade-off Analysis

Present options with concrete trade-offs:
```
## [Decision Topic]

**Option A: [Name]**
- How it works: [brief description]
- Pros: [concrete, tied to constraints]
- Cons: [concrete, with severity]
- Codebase fit: [does it match existing patterns? file:line evidence]

**Option B: [Name]**
- [same structure]

**Recommendation**: [Option] because [reasoning tied to specific constraints and codebase context]
```

### Step 3: Write Proposal

After the user confirms direction:

Run: `rpi scaffold propose --topic "..." --write`

This creates `.thoughts/proposals/YYYY-MM-DD-description.md` with frontmatter pre-populated.

Fill in the sections — for Quick mode, focus on: Summary, Constraints & Requirements, Design Decisions (often just one), References. Skip sections that don't apply (Architecture, File Structure, etc.).

**Mark proposal as active**: `rpi frontmatter transition <proposal-path> active`

**Transition upstream artifacts**: If this proposal was created from a research doc:
1. Re-read the research doc's key findings, suggested next steps, and open questions
2. Verify the proposal addresses them — check each finding was incorporated or explicitly scoped out, each question was answered or deferred with rationale
3. If all points are covered: `rpi frontmatter transition <research-path> complete`
4. If gaps remain, note them:
   ```
   Research doc has unaddressed items:
   - [item not covered in proposal]

   Mark research as complete anyway, or leave it active?
   ```

### Step 4: Create/Update Specs

**This step is explicit, not optional.**

1. Check `.thoughts/specs/` for specs covering affected modules
2. If specs exist: review for accuracy against investigation findings, update if stale
3. If no spec exists for a significantly affected module: create one documenting current behavior
   - Run `rpi scaffold spec --topic "..." --write`
4. Present created/updated specs: "These specs will guide implementation — look right?"

Then suggest the next step:
```
Proposal saved. Ready to plan the implementation?
→ /rpi-plan .thoughts/proposals/YYYY-MM-DD-description.md
```

---

## Full Mode

For features that involve multiple interacting decisions, new components, or significant architectural changes.

### Step 1: Investigate Codebase

Build a thorough understanding of the terrain before proposing solutions.

1. **Read all mentioned files fully** before spawning sub-tasks
2. **Validate upstream status** if a research doc was provided:
   Run: `rpi frontmatter get <research-path> status`
   - If `active`: proceed — this is the expected state
   - If `draft`: warn the user:
     ```
     Warning: Research doc is still in draft — it may not be finalized.
     Consider running `/rpi-research` to complete it first.
     Proceed anyway? (yes / no)
     ```
   - If `complete`: warn the user:
     ```
     Warning: Research doc is already marked complete — it may have already been consumed by a previous proposal.
     Proceed anyway? (yes / no)
     ```
3. **Resolve the artifact chain** if a research doc was provided:
   Run: `rpi chain <input-path>`
   Read all files it identifies.
3. **Check for existing proposals** on the same topic:
   Run: `rpi scan --type proposal`
4. **Spawn parallel research sub-tasks:**
   - Sub-task: "Find components and files related to [feature] — return organized file listings grouped by purpose"
   - Sub-task (@codebase-analyzer): Understand current architecture and patterns in use
   - Sub-task: "Search .thoughts/ for existing research, proposals, and plans about [topic] — return file listings grouped by type"
   - Sub-task: "Find how similar problems are solved in the codebase for [topic] — return concrete code snippets with file:line references"
5. **Read all files identified by research tasks**
6. **Probe for non-functional requirements** — consider which of these matter:
   - Performance (latency, throughput, resource usage)
   - Reliability (error handling, retry behavior, graceful degradation)
   - Security (auth, data sensitivity, injection surfaces)
   - Observability (logging, metrics, debugging)

   Don't force-fit all of these — just surface the genuinely relevant ones.

7. **Present findings + questions → human checkpoint:**
   ```
   Based on investigation:

   Current architecture:
   - [Component/system description with file:line reference]
   - [Relevant pattern or convention in use]

   Constraints I've identified:
   - [Technical constraint from codebase]
   - [Requirement from user]
   - [Non-functional requirement, if relevant]

   Questions before I explore design options:
   - [Question requiring human judgment or domain knowledge]
   ```

   If the user signals they want you to proceed without checkpoints ("just design it", "I trust your judgment"), compress steps and present the synthesized proposal directly.

### Step 2: Explore Design Options → Human Checkpoint

Map the decision space — what are the meaningful choices, and what are the real trade-offs?

1. **Identify the key design dimensions** — independent decisions to make:
   - Data model / storage approach
   - Component decomposition / module boundaries
   - Communication patterns (sync vs async, events vs direct calls)
   - Error handling strategy
   - API surface / interface shape
2. **Spawn parallel sub-tasks** for deeper investigation if needed:
   - Sub-task: "Find similar patterns in the codebase for [topic] — return concrete code snippets with file:line references"
   - Sub-task (when valuable): Web research for library docs, benchmarks, or architectural patterns
3. **Wait for ALL sub-tasks to complete**
4. **Present design options:**
   ```
   ## Design Decisions

   ### Decision 1: [e.g., State Management Approach]

   **Option A: [Name]**
   - How it works: [description]
   - Pros: [concrete advantages]
   - Cons: [concrete disadvantages]
   - Fits existing patterns: [yes/no, with evidence from file:line]

   **Option B: [Name]**
   - [same structure]

   **Recommendation**: [Option] because [reasoning tied to constraints]

   Which options align with your goals?
   ```

   Use diagrams when they clarify relationships or data flow better than prose.

### Step 3: Synthesize → Human Checkpoint

After the user selects directions, validate the combined choices:

1. **Check composition** — do the selected options work together?
   - Any contradictions? (e.g., "eventual consistency" + "immediate validation")
   - Unexpected complexity when combined?
   - Emergent properties — good or bad?
2. **Check integration with existing systems:**
   - Where does the proposal touch existing code? (specific file:line refs)
   - Do existing interfaces need to change?
   - Migration concerns for existing data or behavior?
3. **Identify risks from the full picture** — risks that only appear when seeing the whole design
4. **Define file structure** — when the feature involves new files or module reorganization:
   - New files with responsibilities and exports
   - Modified files with what changes
   - Module boundaries and dependency direction
5. **Present the cohesive proposal:**
   ```
   Proposed approach:

   ## Summary
   [1-3 sentence elevator pitch]

   ## Key Decisions
   1. [Decision]: [Chosen option] — [one-line rationale]
   2. [Decision]: [Chosen option] — [one-line rationale]

   ## How the Parts Connect
   [Component interaction diagram or description]

   ## File Structure (if applicable)
   [New/modified files with responsibilities]

   ## Risks & Mitigations
   - [Risk]: [Mitigation strategy]

   Does this look right before I write the full proposal?
   ```

### Step 4: Write Proposal

**Create the proposal doc**:
- Without research: `rpi scaffold propose --topic "..." --write`
- With research: `rpi scaffold propose --topic "..." --research <path> --write`

This creates `.thoughts/proposals/YYYY-MM-DD-description.md` with frontmatter pre-populated.

**Fill in all proposal sections:**
- Summary (elevator pitch)
- Investigation Findings (what we learned — file:line refs)
- Constraints & Requirements
- Design Decisions (chosen option, alternatives, rationale, evidence for each)
- Architecture (component overview, data flow with diagrams, integration points)
- File Structure (new/modified files, module boundaries — when applicable)
- Risks & Mitigations (table with impact/likelihood/strategy)
- What This Proposal Does NOT Cover
- Open Questions (resolve all before marking complete)
- References

**Mark proposal as active**: `rpi frontmatter transition <proposal-path> active`

**Transition upstream artifacts**: If this proposal was created from a research doc:
1. Re-read the research doc's key findings, suggested next steps, and open questions
2. Verify the proposal addresses them — check each finding was incorporated or explicitly scoped out, each question was answered or deferred with rationale
3. If all points are covered: `rpi frontmatter transition <research-path> complete`
4. If gaps remain, note them:
   ```
   Research doc has unaddressed items:
   - [item not covered in proposal]

   Mark research as complete anyway, or leave it active?
   ```

### Step 5: Create/Update Specs

**This step is explicit, not optional.**

1. Check `.thoughts/specs/` for specs covering affected modules
2. If specs exist: review for accuracy against investigation findings, update if stale
3. If no spec exists for a significantly affected module: create one documenting current behavior
   - Run `rpi scaffold spec --topic "..." --write`
4. Present created/updated specs: "These specs will guide implementation — look right?"

### Step 6: Review & Iterate

1. Present the draft proposal location
2. Iterate based on feedback
3. Resolve all open questions before marking status as `complete`

Then suggest the next step:
```
Proposal saved. Ready to plan the implementation?
→ /rpi-plan .thoughts/proposals/YYYY-MM-DD-description.md
```

---

## Incremental Mode

When the user provides a path to an existing proposal that needs updating:

1. **Read the existing proposal fully**
2. **Understand what's changing** — ask what prompted the update (new requirements, implementation findings, changed constraints)
3. **Assess impact** — which decisions are affected? Which still hold?
4. **Research if needed** — spawn targeted sub-tasks only for the areas that changed
5. **Propose changes** — present what you'd update and why, get buy-in before modifying
6. **Update the document** — modify in place, update the frontmatter:
   ```
   rpi frontmatter set <proposal> status updated
   rpi frontmatter set <proposal> last_updated "<YYYY-MM-DD>"
   rpi frontmatter set <proposal> update_reason "<Brief description of what changed>"
   ```
7. Add an `## Update Log` section at the bottom if one doesn't exist, with a dated entry explaining what changed and why
8. **Update affected specs** if the changes alter documented behavior

---

## Guidelines

1. **Be Opinionated** — Present recommendations with clear reasoning. The user wants opinions grounded in evidence, not a neutral menu.
2. **Be Interactive** — Get buy-in at each checkpoint before proceeding. A proposal that surprises the user during review means the process failed.
3. **Be Evidence-Based** — Ground decisions in codebase patterns, constraints, and concrete trade-offs. "The codebase already uses this pattern in `auth/handler.py:45`" is strong.
4. **Be Focused** — Design at the right level of abstraction — architecture and key interfaces, not implementation details.
5. **Resolve Open Questions** — Don't finalize with unresolved questions. Either answer them through research or make a decision and document the reasoning.
6. **Respect Existing Patterns** — Prefer solutions that align with how the codebase already works. Diverging has a real cost — only do it when the benefit clearly outweighs.
7. **Use Diagrams** — When describing component interactions, data flow, or state transitions, ASCII diagrams often communicate more efficiently than prose.
8. **Specs Are Not Optional** — Every proposal must end with a spec review/creation step. Specs document baseline behavior and serve as implementation guidelines.

## Visual Aids

Use ASCII diagrams when they clarify the proposal:

**Component interaction:**
```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────▶│   API    │────▶│ Service  │
└──────────┘     └──────────┘     └──────────┘
```

**Data flow:**
```
User → API → Validate → Process → Store → Respond
                ↓ (on failure)
           Return error
```

**State transitions:**
```
[Draft] --publish--> [Active] --expire--> [Archived]
```
