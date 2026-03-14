---
description: Research the codebase to understand problems before proposing solutions
model: opus
---

# Research Codebase

Investigate the codebase conversationally to understand how things work, find patterns, and surface insights — without being forced to produce artifacts.

This is the entry point for unclear or open-ended questions: "How does X work?", "What's going on with Y?", "Can we improve Z?". When insights crystallize into something actionable, suggest `/rpi-propose`.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.
Run `rpi --help` to discover available commands and `rpi <command> --help` for detailed usage with examples.

## Step 1: Receive the question

If the user provided a question or topic as command arguments, proceed directly to Step 2.

If no arguments were provided, ask what they'd like to research — this works for focused questions ("how does auth work?") and open-ended exploration ("what could we improve about error handling?").

## Step 2: Read mentioned files and check existing context

- If the user mentions specific files, read them fully before doing anything else
- Run `rpi scan --type research` to check for existing research on this topic
  - If relevant research exists, mention it and build on it

## Step 3: Assess scope and investigate

Scale investigation effort to the question:

- **Focused question** (e.g., "how does the webhook handler work?"): Read the relevant files directly — minimal research needed.
- **Broad or exploratory question** (e.g., "what's the tech debt situation?"): Investigate multiple areas in parallel.

To investigate (adapt to scope — parallelize when possible):

1. Run `rpi index query "[topic]"` to locate relevant files, then read them
2. Understand how the key code works — read the implementation and trace the logic
3. Find existing patterns related to the topic — look for similar code with file:line refs
4. Run `rpi scan` to surface any existing documents about this topic in `.thoughts/`

For focused questions, skip tasks that aren't relevant.

## Step 4: Checkpoint (broad/exploratory questions only)

After initial findings return, present what was found and let the user redirect. Skip for focused questions — just proceed to presenting findings.

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

Offer to save to `.thoughts/research/`. If they agree:
1. Use `rpi` to scaffold and save a research artifact for this topic
2. Fill in key findings, file references, and assessment
3. Use `rpi` to transition the research artifact to active status

**Do not force artifact creation.** The default is conversational — only save when requested or clearly useful.

## Step 7: Transition out

When insights crystallize into something actionable, suggest continuing with `/rpi-propose`. If in the same session, the full context carries over. If in a new session, they can pass the saved summary: `/rpi-propose .thoughts/research/YYYY-MM-DD-topic.md`.

## Guidelines

- **Scale to the question** — a simple question doesn't need broad parallel research
- **No forced artifacts** — this command is conversational by default
- **Be concrete** — always include file:line references, not vague descriptions
- **Be interactive** — checkpoint after initial findings for broad questions
- **Cross-session handoff** — when saving, include enough context that a future `/rpi-propose` can skip re-investigating documented areas
- **Facts first, opinions when warranted** — present what exists before suggesting what should change
- **Follow-ups welcome** — append to the same research doc rather than creating a new one
