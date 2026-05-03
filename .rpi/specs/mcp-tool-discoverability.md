---
domain: mcp-tool-discoverability
feature: mcp-tool-discoverability
last_updated: 2026-04-09T00:36:58+02:00
updated_by: .rpi/designs/2026-04-09-mcp-tool-discoverability.md
---

# mcp-tool-discoverability

## Purpose

Ensure every RPI MCP tool is individually discoverable by AI assistants that use keyword-based tool search. Each tool must have a unique description, keyword-optimized search hints, and hook-invoked tools must be immediately available without a discovery step.

## Scenarios

### Every tool has a unique description
Given the RPI MCP server is running
When a client lists all available tools
Then no two tools share the same description text

### Every tool has a searchHint
Given the RPI MCP server is running
When a client lists all available tools
Then every tool includes an `_meta` field containing `anthropic/searchHint` with a non-empty string value

### Hook-invoked tools are always loaded
Given the RPI MCP server is running
When a client lists all available tools
Then `rpi_session_resume`, `rpi_context_essentials`, and `rpi_suggest_next` each include `_meta` with `anthropic/alwaysLoad` set to `true`

### Non-hook tools are not always loaded
Given the RPI MCP server is running
When a client lists all available tools
Then tools other than `rpi_session_resume`, `rpi_context_essentials`, and `rpi_suggest_next` do not set `anthropic/alwaysLoad`

### Descriptions do not contain sibling tool content
Given a tool derived from a multi-action Cobra command (git-context, frontmatter, verify, extract)
When a client reads that tool's description
Then the description only describes that specific tool's action, not the actions of its sibling tools

### Descriptions are self-contained
Given any RPI MCP tool
When a client reads that tool's description
Then it states what the tool does and what it returns in at most 3 sentences

### searchHint contains no newlines
Given any RPI MCP tool with a searchHint
When the searchHint value is inspected
Then it contains no newline characters

## Constraints
- Tool names must not change (existing hooks and skills reference them)
- CLI behavior is unaffected — Cobra command descriptions remain as-is
- `alwaysLoad` is only for tools called by hooks; other tools remain deferred to minimize context cost

## Out of Scope
- ToolAnnotations (readOnly, destructive hints)
- Changes to tool input schemas or behavior
- Adding or removing tools
