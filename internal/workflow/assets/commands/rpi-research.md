---
description: Research the codebase to understand problems before proposing solutions
model: opus
---

# Research Codebase

Investigate the codebase conversationally to understand how things work, find patterns, and surface insights — without being forced to produce artifacts.

This is the entry point for unclear or open-ended questions: "How does X work?", "What's going on with Y?", "Can we improve Z?". When insights crystallize into something actionable, suggest `/rpi-propose`.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.

## Step 1: Receive the question

If the user provided a question or topic as command arguments, proceed directly to Step 2.

If no arguments were provided, ask:
```
What would you like to research? Provide a question, topic, or file path and I'll investigate.

This works for focused questions ("how does auth work?") and open-ended exploration ("what could we improve about error handling?").
```

## Step 2: Read mentioned files and check existing context

- If the user mentions specific files, read them fully before doing anything else
- Check for existing research: run `rpi scan --type research`
  - If relevant research exists, mention it: "I found existing research on this topic: [path]. I'll build on it."
  - If none, continue

## Step 3: Assess scope and investigate

Scale investigation effort to the question:

- **Focused question** (e.g., "how does the webhook handler work?"): Read the relevant files directly. 1-2 sub-agents max.
- **Broad question** (e.g., "what's the tech debt situation?"): Full investigation with multiple parallel sub-agents across different areas.
- **Exploratory question** (e.g., "what could we improve about X?"): Broad investigation focused on surfacing opportunities and trade-offs.

**Sub-agents to use:**
- Sub-task: "Find WHERE files and components live for [topic] — return organized file listings grouped by purpose"
- Sub-task (@codebase-analyzer): Understand HOW specific code works
- Sub-task: "Find examples of existing code patterns for [topic] — return concrete snippets with file:line references"
- Sub-task: "Search .thoughts/ for existing documents about [topic] — return file listings grouped by type"

Run sub-agents in parallel when they're searching for different things.

## Step 4: Checkpoint (broad/exploratory questions only)

After initial findings return, present what was found and let the user redirect:

```
Initial findings on [topic]:
- [Area 1]: [brief summary with file:line refs]
- [Area 2]: [brief summary with file:line refs]
- [Area 3]: [brief summary with file:line refs]

Should I dig deeper into any of these, or focus somewhere specific?
```

Skip this for focused questions — just proceed to presenting findings.

## Step 5: Present findings

Synthesize all results:
- Prioritize live codebase findings as the primary source of truth
- Include specific file paths and line numbers
- Connect findings across different components
- Answer the user's specific questions with concrete evidence
- When findings reveal clear pain points or opportunities, include your assessment — but keep facts and opinions clearly separated

Ask if they have follow-up questions.

## Step 6: Optional summary save

If the user asks to save findings ("save this", "summarize what we found"), or if the investigation is substantial enough that cross-session handoff is clearly valuable:

```
Want me to save these findings to `.thoughts/research/`? This makes them available for future sessions — e.g., as input to `/rpi-propose`.
```

If they agree:
1. Run `rpi scaffold research --topic "..." --write`
2. Fill in key findings, file references, and assessment
3. Mark as active: `rpi frontmatter transition <research-path> active`
4. Present the path: "Saved to `.thoughts/research/YYYY-MM-DD-topic.md`"

**Do not force artifact creation.** The default is conversational — only save when requested or clearly useful.

## Step 7: Transition out

When insights crystallize into something actionable, suggest the next step:

```
This looks like it could turn into a concrete proposal. Want to continue with `/rpi-propose`?
```

If in the same session, the user can proceed directly and the full context carries over. If in a new session, they can pass the saved summary: `/rpi-propose .thoughts/research/YYYY-MM-DD-topic.md`.

## Guidelines

- **Scale to the question** — a simple question doesn't need five sub-agents
- **No forced artifacts** — this command is conversational by default
- **Be concrete** — always include file:line references, not vague descriptions
- **Be interactive** — checkpoint after initial findings for broad questions
- **Cross-session handoff** — when saving, include enough context that a future `/rpi-propose` can skip re-investigating documented areas
- **Facts first, opinions when warranted** — present what exists before suggesting what should change
- **Follow-ups welcome** — append to the same research doc rather than creating a new one
