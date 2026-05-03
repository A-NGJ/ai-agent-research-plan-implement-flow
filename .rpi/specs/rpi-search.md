---
domain: rpi-search
feature: rpi-search
last_updated: 2026-04-30T00:00:00+02:00
updated_by: .rpi/designs/2026-04-17-semantic-artifact-discovery.md
---

# rpi-search

## Purpose

Let AI assistants and users discover semantically relevant prior artifacts in `.rpi/` by issuing natural-language queries. When a semantic search backend is available and ready, return ranked hits reflecting the current state of the artifact corpus across both keyword and semantic matching paths. When the backend is unavailable, broken, or in a recoverable first-run state, report that distinctly so callers can fall back to non-semantic discovery or surface a fixable instruction to the user.

## Scenarios

### Ranked hits for a natural-language query
Given a semantic search backend is installed and ready, and the `.rpi/` corpus contains artifacts related to the query topic
When a caller issues `rpi_search` with a natural-language query
Then the response status is "ok" and returns hits ranked by relevance, each with path, artifact type, title, score, and a snippet

### No matching artifacts is distinct from a tool failure
Given a semantic search backend is installed and ready, and no artifact in the corpus matches the query
When a caller issues `rpi_search`
Then the response status is "empty" with no hits, distinguishable from any error or unavailability state

### Not-installed backend is reported with an install hint
Given no semantic search backend is installed
When a caller issues `rpi_search` with any query
Then the response status is "backend_unavailable", states the reason, and includes an install hint and a suggested fallback path

### Installed-but-failing backend surfaces diagnostic detail
Given a semantic search backend is installed but a refresh, query, or output-parse step fails
When a caller issues `rpi_search`
Then the response status is "backend_error" with a stage, message, and actionable hint distinct from "backend_unavailable", and the tool call itself completes successfully

### Recoverable first-run state is surfaced as fixable
Given a semantic search backend is installed but its required models or supporting daemon are not yet ready
When a caller issues `rpi_search`
Then the response status is "backend_error" with a stage indicating first-run setup is required and a hint naming the user-invoked command that completes setup, and no setup is triggered automatically

### Recent edits appear in keyword matches without manual re-indexing
Given a semantic search backend is installed and ready, and the user has just written or edited an artifact in `.rpi/`
When a caller issues `rpi_search` whose query matches the new or edited content via keywords
Then the response includes that artifact in the hits without the caller needing to trigger an index refresh

### Recent edits appear in semantic matches without manual re-indexing
Given a semantic search backend is installed and ready, and the user has just written or edited an artifact in `.rpi/`
When a caller issues `rpi_search` whose query matches the new or edited content semantically (without exact keyword overlap)
Then the response includes that artifact in the hits without the caller needing to trigger an index refresh

### Archive is included by default
Given the `.rpi/archive/` directory contains artifacts matching the query
When a caller issues `rpi_search` without opting out of archived results
Then archived artifacts may appear in the response alongside active ones, indistinguishable in shape from other hits apart from their path

### Archive is excludable on request
Given the `.rpi/archive/` directory contains artifacts matching the query
When a caller issues `rpi_search` with archive exclusion opted in
Then no archived artifact appears in the response

### Filter by artifact type
Given the `.rpi/` corpus contains artifacts of multiple types matching the query
When a caller issues `rpi_search` with a type filter
Then every hit in the response has the requested artifact type

### Hit count respects the requested limit
Given a semantic search backend is installed and ready, and many artifacts match the query
When a caller issues `rpi_search` with a specified result limit
Then the response contains no more hits than the requested limit

### Snippet and score are enough to filter without reading full files
Given the response contains hits with snippets and scores
When a caller filters by a score threshold and reads only the snippets
Then it can decide which hits warrant a full document read without re-querying or fetching additional data

### Cross-project isolation
Given multiple projects on the same machine each use RPI and have artifacts on similar topics
When a caller in one project issues `rpi_search`
Then the response contains only artifacts from that project's `.rpi/` directory

## Constraints

- The semantic search backend is an optional dependency; RPI must function fully without it.
- The tool contract is backend-agnostic — request and response shapes describe intent and hits, not backend-specific fields.
- Index freshness is guaranteed by the tool across both keyword and semantic paths; callers never need to trigger re-indexing.
- Archive inclusion is on by default; callers opt out via an explicit flag.
- The tool never installs the backend, downloads models, spawns a daemon, or otherwise mutates the user's environment without an explicit user-invoked command.

## Out of Scope

- Cross-project search (indexing other repositories).
- Returning dependency chains or link expansion for hits — composition with other tools happens at the caller level.
- LLM-based reranking beyond what the backend natively provides.
- File-watcher or daemon-based auto-indexing of file changes (a model-warming daemon is an implementation detail, not a file watcher).
