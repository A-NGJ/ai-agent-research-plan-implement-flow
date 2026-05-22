---
name: rpi-research
description: Investigate a question conversationally — codebase exploration or external systems/libraries/frameworks. Use when user says 'how does X work?', 'how could X be improved?', 'investigate Y', 'explore Z', 'what frameworks exist for X', 'survey the X space', or 'what's the state of X', even if they don't say 'research'. Do NOT invoke for broken behavior (use rpi-diagnose) or concrete changes with tradeoffs (use rpi-propose).
---

# Research

## Goal

Investigate conversationally to understand how things work, find patterns, and surface insights — whether the question is about the codebase or about external systems, libraries, or frameworks.

When insights crystallize into something actionable, suggest → `/rpi-propose` (with the research artifact path if one was saved).

## Invariants

- Always interview before investigating — ask about motivation, prior attempts, constraints, and success criteria (1-2 questions at a time, adapt based on answers)
- Reflect back a concise problem statement and get confirmation before investigating
- Before investigating, search for prior research artifacts on the topic; surface findings to the user before opening new research
- Read all mentioned files fully before investigating further
- Scale investigation to the question — focused questions need minimal research; broad questions need parallel investigation across multiple sources or codebase areas
- Cite the source of every finding — file:line for codebase claims, URL or quoted documentation for external claims. No vague descriptions.
- For external investigation, prefer authoritative sources (project README, official docs, release notes); flag findings drawn from blog posts or forum threads as such.
- Checkpoint after initial findings for broad/exploratory questions — let the user redirect
- Do not force artifact creation — save to `.rpi/research/` only when asked or clearly valuable for cross-session handoff
- If saving: scaffold a research artifact, fill in findings, and transition to active

## Principles

- Be interactive — stop interviewing when you have enough; ask more if findings raise new questions
- Facts first, opinions when warranted — present what exists before suggesting what should change
- Follow-ups welcome — append to existing research docs rather than creating new ones
