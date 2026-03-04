---
description: Create high-level solution designs with architectural decisions and trade-off analysis
model: opus
---

# Solution Design

Create high-level solution designs by exploring trade-offs, making architectural decisions, and documenting the chosen approach. This stage bridges research findings and implementation planning.

This is part of the pipeline: **research → design → plan → implement**. For complex features involving many new files or major module reorganizations, `/create-structure` is available as an optional step between design and plan.

Design is where you make the hard choices — choosing between approaches, identifying risks, and ensuring decisions compose well together. A good design document captures *why* choices were made (not just what was chosen), so future readers can evaluate whether the reasoning still holds when circumstances change.

## Initial Response

**Auto-detect the mode from what's provided:**

- **Path to research doc or ticket describing a complex feature** → Comprehensive mode
- **Plain description of a focused decision** (e.g., "should we use X or Y?", "design the caching approach") → Lightweight mode
- **Path to an existing design doc** → Incremental mode (updating a previous design)
- **Nothing provided** → Ask:
  ```
  I'll help you create a solution design.

  Please provide:
  1. The task/ticket description (or reference to a ticket file)
  2. Any relevant research documents from `.thoughts/research/`
  3. Constraints, non-functional requirements, or preferences

  Tip: `/create-design .thoughts/research/2025-01-08-authentication-flow.md`
  ```

---

## Lightweight Mode

For focused decisions that don't warrant a full architecture document — choosing between two approaches, designing a single component's interface, deciding on a data model shape. The output is still a design doc, just a concise one.

### Step 1: Quick Context

1. **Read any mentioned files fully**
2. **Do proportional research** — scale to the decision's scope:
   - If the decision is localized (one module, one pattern): read the relevant files directly
   - If it touches multiple areas: spawn 1-2 targeted sub-tasks (locate-codebase, find-patterns)
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

### Step 3: Document the Decision

After the user confirms direction, write a lightweight design doc to `.thoughts/designs/YYYY-MM-DD-description.md`:

```markdown
---
date: [ISO 8601 datetime with timezone]
topic: "[Decision Topic]"
tags: [design, relevant-component-names]
status: complete
---

# Design: [Decision Topic]

## Decision
[What was decided, 1-2 sentences]

## Context
[Why this decision was needed — the problem or requirement]

## Options Considered

### [Option A]
[Brief description, key pros/cons]

### [Option B]
[Brief description, key pros/cons]

## Rationale
[Why the chosen option wins — tied to specific constraints, codebase patterns, or requirements]

## Consequences
[What this decision enables, prevents, or requires in follow-up work]

## References
- [file:line references to relevant code]
```

Then proceed to the next pipeline stage or ask if the user wants to continue.

---

## Comprehensive Mode

For features that involve multiple interacting decisions, new components, or significant architectural changes.

### Step 1: Context Gathering

Build a thorough understanding of the terrain before proposing solutions. Rushing to solutions without understanding the landscape leads to designs that fight the codebase instead of working with it.

1. **Read all mentioned files fully** (no limit/offset) before spawning sub-tasks
2. **Spawn parallel research sub-tasks** using the Task tool. Each sub-task should load the appropriate skill first, then perform its work:
   - Sub-task: "Load the `locate-codebase` skill, then find components related to [task description]"
   - Sub-task (@codebase-analyzer): Understand current architecture and patterns in use
   - Sub-task: "Load the `locate-thoughts` skill, then find existing research, designs, and plans about [topic]"
   - Sub-task: "Load the `find-patterns` skill, then find how similar problems were solved in the codebase for [topic]"
3. **Read all files identified by research tasks**
4. **Probe for non-functional requirements** — these are commonly overlooked and cause expensive rework later. Consider which of these matter for this design:
   - Performance (latency, throughput, resource usage)
   - Reliability (error handling, retry behavior, graceful degradation)
   - Security (auth, data sensitivity, injection surfaces)
   - Observability (logging, metrics, debugging)
   - Scalability (data growth, user growth, concurrency)

   Don't force-fit all of these — just surface the ones genuinely relevant to the feature.
5. **Present current understanding:**
   ```
   Based on the research and analysis:

   Current architecture:
   - [Component/system description with file:line reference]
   - [Relevant pattern or convention in use]

   Constraints I've identified:
   - [Technical constraint from codebase]
   - [Requirement from ticket/user]
   - [Non-functional requirement, if relevant]

   Questions before I explore design options:
   - [Question requiring human judgment or domain knowledge]
   ```

   If the user indicates they want you to proceed without checkpoints ("just design it", "I trust your judgment"), compress Steps 1-3 and present the synthesized design directly, then iterate from there.

### Step 2: Design Exploration

Map the decision space — what are the meaningful choices, and what are the real trade-offs? The goal isn't to list every conceivable option, but to identify the 2-3 genuinely viable approaches and make it clear why one might be better than another.

1. **Identify the key design dimensions** — what are the independent decisions to make? Common dimensions include:
   - Data model / storage approach
   - Component decomposition / module boundaries
   - Communication patterns (sync vs async, events vs direct calls)
   - Error handling strategy
   - API surface / interface shape
2. **Spawn parallel sub-tasks** for deeper investigation:
   - Sub-task: "Load the `find-patterns` skill, then find similar patterns in the codebase or adjacent systems for [topic]"
   - Sub-task: "Load the `analyze-thoughts` skill, then extract insights from [specific document paths] about [topic]"
   - Sub-task (when valuable): Web research for library docs, benchmarks, or architectural patterns that inform the decision

   Web research is valuable when: evaluating external libraries or tools, designing around third-party APIs, or when the codebase is greenfield with no internal patterns to follow. Don't default to it — reach for it when external context would genuinely improve the design.
3. **Wait for ALL sub-tasks to complete** before proceeding
4. **Present design options with concrete trade-off analysis:**
   ```
   ## Design Decisions

   ### Decision 1: [e.g., State Management Approach]

   **Option A: [Name]**
   - How it works: [description]
   - Pros: [concrete advantages]
   - Cons: [concrete disadvantages]
   - Fits existing patterns: [yes/no, with evidence from file:line]

   **Option B: [Name]**
   - How it works: [description]
   - Pros: [concrete advantages]
   - Cons: [concrete disadvantages]
   - Fits existing patterns: [yes/no, with evidence]

   **Recommendation**: [Option] because [reasoning tied to constraints]

   ### Decision 2: [next key decision]
   ...

   Which options align with your goals?
   ```

   Use diagrams when they clarify relationships or data flow better than prose — see the Visual Aids section below.

### Step 3: Design Synthesis

This is the most critical step. Individual decisions might each look sensible in isolation, but designs fail at the *seams* — where decisions interact in ways nobody anticipated. A fast data model paired with a synchronous API might create a bottleneck. A flexible plugin architecture paired with a strict type system might create friction.

After the user selects directions:

1. **Validate the combined choices work together:**
   - Do the selected options create any contradictions? (e.g., choosing "eventual consistency" but also "immediate validation")
   - Do they create unexpected complexity when combined? (e.g., each decision is simple alone but together they require a coordination layer nobody planned for)
   - Are there emergent properties — good or bad — from the combination?
2. **Check integration with existing systems:**
   - Where does the design touch existing code? (specific file:line references)
   - Do existing interfaces need to change, or can the design work with what's there?
   - Are there migration concerns for existing data or behavior?
3. **Identify risks that emerge from the full picture** — risks from individual decisions should already be discussed in Step 2, but some risks only appear when you see the whole design:
   - Operational risks (deployment complexity, rollback difficulty)
   - Evolution risks (will this design accommodate likely future changes?)
   - Failure mode risks (what happens when component X fails — does it cascade?)

   If you discover that the combined choices create a serious problem, surface it honestly rather than papering over it. It's better to revisit a decision now than to discover the problem during implementation.
4. **Present the cohesive design for buy-in:**
   ```
   Proposed design summary:

   ## Approach
   [1-3 sentence high-level description]

   ## Key Decisions
   1. [Decision]: [Chosen option] — [one-line rationale]
   2. [Decision]: [Chosen option] — [one-line rationale]

   ## How the Parts Connect
   [Component interaction diagram or description — data flow, dependencies, event flow]

   ## Risks & Mitigations
   - [Risk]: [Mitigation strategy]

   Does this design direction look right before I document it?
   ```

### Step 4: Write the Design Document

Save to `.thoughts/designs/YYYY-MM-DD-description.md`
- With ticket: `2025-01-08-<ticket-id>-feature-name.md`
- Without ticket: `2025-01-08-improve-error-handling.md`

**Template:**

````markdown
---
date: [Current date and time with timezone in ISO format]
topic: "[Feature/Task Name]"
tags: [design, architecture, relevant-component-names]
status: draft
related_research: [path to research doc if available]
---

# Design: [Feature/Task Name]

## Overview
[Brief description of what we're designing and the problem it solves]

## Context
- **Research**: [Link to research document if available]
- **Ticket**: [Link to ticket if available]
- **Current State**: [Brief description of the current system]

## Constraints & Requirements
- [Hard constraint from the system or business]
- [Non-functional requirement: performance, security, etc.]
- [Compatibility requirement]

## Design Decisions

### Decision 1: [Decision Title]
**Chosen**: [Option name]
**Alternatives considered**: [Other options briefly]
**Rationale**: [Why this option was chosen, tied to specific constraints]
**Evidence**: [File references, benchmarks, or patterns that support this]

### Decision 2: [Decision Title]
[Same structure...]

## Architecture

### Component Overview
[Description of the main components and their responsibilities]

### Data Flow
[How data moves through the system — use diagrams for clarity]

```
[Input] --> [Component A] --> [Component B] --> [Output]
                 |                   ^
                 v                   |
            [Storage] ----------> [Cache]
```

### Integration Points
[Where this design touches existing systems, with file:line references]

### API Contracts (if applicable)
[Key interfaces, function signatures, or API shapes that other components depend on]

## File Structure (if applicable)
<!-- Include this section when the design introduces new files/modules or reorganizes existing ones.
     For large-scale reorganizations or greenfield projects with many new files, consider using
     /create-structure for a dedicated deep-dive instead. -->

### New Files
- `path/to/new-file.ext` — [responsibility, key exports]

### Modified Files
- `path/to/existing-file.ext` — [what changes and why]

### Module Boundaries
- [Module A] depends on [Module B] via [interface]
- [No circular dependencies]

## Risks & Mitigations
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| [Risk description] | High/Med/Low | High/Med/Low | [Strategy] |

## What This Design Does NOT Cover
[Explicit out-of-scope items to prevent scope creep]

## Open Questions
[Any remaining questions — resolve all before marking status as complete]

## References
- Research: `[path to research doc]`
- Similar implementation: `[file:line]`
- External reference: [link if applicable]
````

### Step 5: Review & Iterate

1. Present the draft design document location
2. Iterate based on feedback
3. Resolve all open questions before marking status as `complete`
4. Continue until user is satisfied

---

## Incremental Mode

When the user provides a path to an existing design doc that needs updating:

1. **Read the existing design fully**
2. **Understand what's changing** — ask the user what prompted the update (new requirements, implementation findings, changed constraints)
3. **Assess impact** — which decisions are affected? Which still hold?
4. **Research if needed** — spawn targeted sub-tasks only for the areas that changed
5. **Propose changes** — present what you'd update and why, get buy-in before modifying
6. **Update the document** — modify in place, update the frontmatter:
   ```yaml
   status: updated
   last_updated: [YYYY-MM-DD]
   update_reason: "[Brief description of what changed]"
   ```
7. Add an `## Update Log` section at the bottom if one doesn't exist, with a dated entry explaining what changed and why

---

## Guidelines

1. **Be Opinionated** — Present recommendations with clear reasoning. The user wants opinions grounded in evidence, not a neutral menu of options. That said, flag genuine toss-ups honestly rather than pretending one option is clearly better.
2. **Be Interactive** — Get buy-in on key decisions before documenting the full design. A design that surprises the user during review means the process failed. However, if the user signals they want you to proceed autonomously, respect that and present a complete design for review.
3. **Be Evidence-Based** — Ground decisions in codebase patterns, constraints, and concrete trade-offs. "This is the standard approach" is weak; "the codebase already uses this pattern in `auth/handler.py:45` and `api/middleware.py:12`" is strong.
4. **Be Focused** — Design at the right level of abstraction — architecture and key interfaces, not implementation details. If you're specifying loop bodies, you've gone too deep; if you're just saying "add a service," you haven't gone deep enough.
5. **Resolve Open Questions** — Don't finalize with unresolved questions. They'll haunt the implementation. Either answer them through research or make a decision and document the reasoning.
6. **Respect Existing Patterns** — Prefer solutions that align with how the codebase already works. Diverging from established patterns has a real cost (cognitive overhead, inconsistency, maintenance burden) — only do it when the benefit clearly outweighs that cost.
7. **Use Diagrams** — When describing component interactions, data flow, or state transitions, an ASCII diagram often communicates in 5 lines what takes 3 paragraphs of prose. Don't force them everywhere, but reach for them when relationships are the point.

## Visual Aids

Use ASCII diagrams when they clarify the design. Common patterns:

**Component interaction:**
```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────▶│   API    │────▶│ Service  │
└──────────┘     └──────────┘     └──────────┘
                      │                 │
                      ▼                 ▼
                 ┌──────────┐     ┌──────────┐
                 │  Cache   │     │    DB    │
                 └──────────┘     └──────────┘
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
   ↑                    |
   +----revert----------+
```

## Common Design Patterns

When identifying the shape of the problem, these patterns are common starting points. Each has characteristic decisions:

- **Data Pipeline**: Source → Ingestion → Transform → Storage → Query → Presentation. Key decisions: batch vs stream, error handling per stage, idempotency, backpressure.
- **New Service/Module**: Identify boundary → Define interface → Design internals → Plan integration. Key decisions: sync vs async communication, data ownership, failure isolation.
- **Refactoring**: Map current state → Identify target state → Design migration path → Define compatibility layer. Key decisions: big bang vs incremental migration, backward compatibility duration, rollback plan.
- **API Extension**: Understand existing surface → Design new endpoints/fields → Plan versioning → Define validation. Key decisions: breaking vs non-breaking changes, deprecation strategy, backward compatibility.
- **Event-Driven**: Identify events → Define event schema → Design handlers → Plan ordering/delivery guarantees. Key decisions: at-least-once vs exactly-once, event store vs fire-and-forget, consumer independence.
