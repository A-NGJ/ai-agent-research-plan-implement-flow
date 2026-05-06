---
name: rpi-explain
description: Walk through a recent diff and explain what changed and why, focusing on non-obvious decisions. Use when user says 'explain the changes', 'walk me through the diff', 'what did we just do?', or 'explain how <feature> was implemented'.
---

# Explain Implementation

## Goal

Generate a diff-scoped walkthrough of an implemented solution, explaining what changed and why — with focus on non-obvious decisions. This is an optional post-implement step: research → propose → plan → implement → **explain**.

Auto-detect the mode from input:
- **Artifact path** (path to plan, design, or spec) → resolve the artifact chain for context, then walk through the diff
- **No arguments** → auto-detect changed files from git and explain without artifact context

After the initial walkthrough, enter conversation mode — the user can ask follow-up questions to dig deeper into any part of the changes. When done, offer to save the explanation as an artifact if the user wants it.

## Invariants

- If an artifact path is provided, resolve its full chain (plan → design → research) and use as context for attributing rationale (EX-1)
- If no arguments, identify changed files from git and proceed without artifact context (EX-2)
- In no-args mode, after identifying changed files, search for related designs, plans, or diagnoses that match the diff's topic — prefer semantic search when available (default relevance threshold ~0.4), and fall back to keyword-based artifact discovery when not. Read snippets first. Use high-relevance hits as additional rationale sources, citing them as "per the design..." rather than "inferred from code". Skip this step in artifact-path mode — explanation should stay grounded in the provided artifact's chain. If semantic search reports an installed-but-failing state, surface its hint before falling back. (EX-10)
- If a provided path doesn't exist or has no linked artifacts, proceed with diff-only explanation and note the missing context (EX-3)
- Read all changed files fully before generating explanations — never use limit/offset
- Walk through changes file-by-file, providing a factual summary of what changed in each (EX-4)
- For non-obvious changes, provide explicit callouts explaining the reasoning — inferred from artifacts when available, from code context otherwise (EX-5)
- Clearly distinguish between rationale sourced from artifacts ("per the design...") vs inferred from code context ("based on the surrounding code...") (EX-6)
- Summarize straightforward changes briefly (1-2 sentences) rather than explaining the obvious (EX-7)
- After the walkthrough, invite questions — answer follow-ups by reading additional context as needed
- Do not save an artifact by default — only when the user explicitly requests it (EX-8)
- When saving, scaffold to `.rpi/reviews/` with a descriptive topic (EX-9)
- Include file:line references in all explanations

## Principles

- Explain, don't judge — no pass/fail ratings or severity classifications (that's `/rpi-verify`)
- Be interactive — the initial walkthrough is a starting point, not the final word; follow-up questions are expected
- Attribute your sources — when rationale comes from an artifact, cite it; when inferred, say so; when uncertain, flag it
- Scale to the diff — large diffs get grouped by theme; small diffs get file-by-file detail
- Prioritize the non-obvious — mechanical changes (renames, imports, formatting) get a brief mention; design decisions and behavioral changes get thorough explanation
