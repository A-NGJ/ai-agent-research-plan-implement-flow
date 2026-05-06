---
name: rpi-research
description: Explore the codebase to understand how something works. Use when user says 'how does X work?', 'how could X be improved?', 'investigate Y', or 'explore Z', even if they don't say 'research'. Do NOT invoke for broken behavior (use rpi-diagnose) or concrete changes with tradeoffs (use rpi-propose).
---

# Codebase Research

## Goal

Investigate the codebase conversationally to understand how things work, find patterns, and surface insights. This is the entry point: **research** → propose → plan → implement.

When insights crystallize into something actionable, suggest → `/rpi-propose` (with the research artifact path if one was saved).

## Invariants

- Always interview before investigating — ask about motivation, prior attempts, constraints, and success criteria (1-2 questions at a time, adapt based on answers)
- Reflect back a concise problem statement and get confirmation before codebase investigation
- After problem confirmation, search for prior research artifacts whose content semantically matches the problem statement — prefer semantic search when available (default relevance threshold ~0.4), and fall back to name/keyword artifact discovery when not. Read snippets first; only read full files for hits with strong relevance. For promising hits, expand to see whether downstream design or plan artifacts already exist (answers "has this been picked up already?"). Surface findings to the user before fresh codebase investigation. If semantic search reports an installed-but-failing state, surface its hint before falling back.
- Read all mentioned files fully before investigating further
- Scale investigation to the question — focused questions need minimal research; broad questions need parallel investigation across multiple areas
- Include file:line references in all findings — no vague descriptions
- Checkpoint after initial findings for broad/exploratory questions — let the user redirect
- Do not force artifact creation — save to `.rpi/research/` only when asked or clearly valuable for cross-session handoff
- If saving: scaffold a research artifact, fill in findings, and transition to active

## Principles

- Be interactive — stop interviewing when you have enough; ask more if findings raise new questions
- Facts first, opinions when warranted — present what exists before suggesting what should change
- Follow-ups welcome — append to existing research docs rather than creating new ones
