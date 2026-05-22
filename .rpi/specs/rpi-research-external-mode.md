---
domain: skills
feature: rpi-research-external-mode
last_updated: 2026-05-22T12:28:50+02:00
updated_by: .rpi/designs/2026-05-22-expand-rpi-research-to-cover-external-investigation.md
---

# rpi-research external mode

## Purpose

Define the behavioral contract for `rpi-research` when the question is about systems, libraries, frameworks, or patterns that live outside the current codebase. Covers when the skill fires, how findings cite their sources, how authoritative-vs-anecdotal sources are flagged, and the disambiguation boundary against `rpi-propose` (decisions) and `rpi-diagnose` (broken behavior).

## Scenarios

### External-survey questions fire research
Given a developer asks an informational question about systems outside the current codebase — for example, *"what agentic frameworks exist that I could use for a data analytics tool"*, *"what's the state of vector-database options in 2026"*, or *"survey the auth-service-mesh space"*
When the prompt is processed in a fresh session with all RPI skills installed
Then `rpi-research` fires automatically without slash-command invocation

### Decision-flavored external questions defer to propose
Given a developer asks an external question that requires picking between options — for example, *"should we use LangGraph or the Claude SDK for our agent"*, *"is Postgres or Mongo a better fit for our event store"*
When the prompt is processed
Then `rpi-propose` fires, not `rpi-research`, because the question seeks a decision rather than a survey

### Broken-external-system questions defer to diagnose
Given a developer reports broken behavior in an external system they depend on — for example, *"why is LangGraph dropping our messages"*, *"the Stripe webhook is firing twice on retry"*
When the prompt is processed
Then `rpi-diagnose` fires, not `rpi-research`, because the user reports broken behavior

### Findings cite their sources by mode
Given the skill produces findings during an external-investigation conversation
When the findings are presented to the user (and optionally saved to `.rpi/research/`)
Then each external claim carries a URL or a quoted snippet from documentation, and each codebase claim carries a `file:line` reference — no claim is presented without an anchor the reader can verify

### Authoritative sources are preferred and weaker sources are flagged
Given the skill is gathering external evidence
When findings draw from project README files, official documentation, release notes, or RFCs alongside findings drawn from blog posts, forum threads, or social-media posts
Then the authoritative findings are presented as primary and the weaker findings are explicitly flagged as such (e.g., *"per a 2025 blog post by …"*) so the user can weigh confidence

### Mixed questions investigate both modes
Given a developer asks a question that spans both the codebase and an external system — for example, *"how does our auth middleware compare to industry patterns"*, *"could we replace our queue with a managed service"*
When the skill investigates
Then it explores the relevant codebase paths (file:line refs) and surveys the external space (URL/quote refs) in the same conversation, weaving both anchor styles into the findings

### External findings flow to the pipeline through propose
Given an external-research conversation produces actionable insights the user wants to move on
When the skill suggests the next step
Then it suggests `→ /rpi:rpi-propose` (the same handoff as codebase research), optionally passing the saved research-artifact path so propose can ground its tradeoff analysis in the external survey

## Constraints

- The skill description includes both codebase and external-investigation trigger phrasings while preserving the existing negative gates against `rpi-propose` and `rpi-diagnose`.
- The citation rule applies to all findings: `file:line` for codebase, URL or quoted documentation for external. No vague summaries are acceptable in either mode.
- The skill does not pin specific external tools (`WebSearch`, `WebFetch`, docs MCP, etc.) by name in its body; tool selection is delegated to the host agent, consistent with the no-tool-name policy in `internal/workflow/workflow_test.go:346-351`.
- External-research findings are saved (when saved) to `.rpi/research/` using the same `YYYY-MM-DD-<slug>.md` naming convention — no new artifact directory.
- The 20-prompt manual eval defined in `.rpi/specs/skill-descriptions.md:42-46` is extended with at least one positive external-research probe and one disambiguation probe against `rpi-propose`. The acceptance threshold remains ≥80% pass rate after the extension.
- Skill body word count does not increase materially from the current baseline (`rpi-research` was 410 words per the 2026-05-18 conciseness research). Net change target: ≤ 0 words.

## Out of Scope

- A canonical per-skill spec retroactively documenting `rpi-research`'s full codebase behavior.
- Renaming the skill (`rpi-investigate`, `rpi-explore`, etc.) — migration cost not justified by a description rewrite alone.
- Creating a separate `rpi-discover` skill — explicitly rejected during the design's grill phase.
- A separate artifact directory (`.rpi/discoveries/`) for external findings.
- Tooling-level changes (no new MCP tools, no CLI subcommands, no host-agent tool list changes).
- Enforcing citation density (e.g., "every paragraph must carry a URL") — left as a future tightening only if drift is observed.
