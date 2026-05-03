---
domain: rpi-wiki
feature: rpi-wiki
last_updated: 2026-04-30T00:00:00+02:00
---

# rpi-wiki

## Purpose

Persist enduring domain knowledge — stakeholder decisions, domain rules, project glossary — that has no other home in the existing project artifacts. Capture happens silently during normal conversation when Claude detects qualifying signals; consumption happens on demand via semantic search. Per-project (`.rpi/wiki/`) and global (`~/.rpi/wiki/`) wikis exist in parallel and are searched together by default.

## Scenarios

### Authority signal captured silently
Given the user states a stakeholder decision in conversation ("Alan confirmed CI takes precedence over base MMIT")
When Claude responds to the message
Then a wiki entity page on that decision exists in the project wiki, the page records the attribution, and the user is not prompted

### Domain rule captured silently
Given the user states a declarative domain fact whose source of truth lives nowhere else ("CI data only changes twice a year")
When Claude responds to the message
Then a wiki entity page on that domain rule exists in the project wiki and the user is not prompted

### Anti-signal content not captured
Given the user states something whose source of truth lives in the codebase, git history, eval logs, metrics, or an external tracker (e.g., "we created class Foo in foobar.py", "Opus beats Kimi on current data")
When Claude responds to the message
Then no wiki page is created, regardless of how decision-like the statement appears

### Contradicting fact interrupts the user
Given a wiki entity page already exists asserting fact X
When the user states a fact that contradicts X
Then Claude asks the user to choose between overwriting the page, superseding it with a history note, or skipping the new fact, naming the affected page

### Non-conflicting refinement updates the page in place
Given a wiki entity page already exists describing concept C
When the user states a related fact about C that does not contradict the existing content
Then Claude updates the existing page in place silently — no new page is created and the user is not prompted

### Wiki search returns ranked entries from both wikis
Given the project wiki and the global wiki both contain entity pages relevant to a topic
When a caller queries `wiki_search` with a natural-language query about that topic and no scope restriction
Then the response status is "ok" and returns ranked hits from both wikis, each hit including the page path, the source ("project" or "global"), score, and snippet

### Manual addition writes a wiki page
Given the user invokes `/wiki-add` with a fact (optionally with `--global`)
When the command completes
Then a wiki page containing that fact exists at the named scope (project by default, global with `--global`) and the user is not asked to classify the content

### Manual forget removes a page from search
Given a wiki page exists for a topic
When the user invokes `/wiki-forget` naming that topic or path
Then subsequent `wiki_search` queries no longer return that page

## Constraints

- Capture for new pages is fully silent — no prompts, no questions, no end-of-turn confirmations. The only break in silence is the contradiction interrupt.
- A single anti-signal applies absolutely: if the source-of-truth for a fact lives in code, git history, eval logs, metrics, or an external tracker, the wiki never captures it — regardless of how decision-like the statement sounds.
- Project wiki at `.rpi/wiki/` and global wiki at `~/.rpi/wiki/` are isolated indexing collections. Cross-project wiki leakage is impossible.
- Each entity page concerns one concept, refined in place over time as new related facts arrive.
- Default capture target is the project wiki when in a project. Global wiki is written explicitly via the `--global` flag on manual commands.
- The wiki is read on demand; no wiki content is auto-loaded into session context.
- The `wiki_search` tool follows the same four-state status contract as `rpi_search` (`ok` / `empty` / `backend_error` / `backend_unavailable`).

## Out of Scope

- Medium-confidence batch tier and `/wiki-review` command (v2).
- `raw/` directory for immutable source documents and `/wiki-ingest` command (v2).
- `/wiki-lint` command for orphan and stale-claim detection (v2).
- Automatic project → global promotion (v2; `--global` flag only in v1).
- Hook-based capture outside the active conversation (v2 hardening).
- Cross-project wiki search.
- Page versioning beyond git.
