---
name: find-patterns
description: Find similar implementations, usage examples, or existing code patterns that can be modeled after. Returns concrete code snippets with file:line references. Use when you need to understand how something is currently done in the codebase before implementing something new, and especially when you want to follow established conventions. Use this skill proactively whenever a task involves adding something similar to what already exists — new endpoints, new sync classes, new test files, new factories, etc.
---

## Purpose

Find existing code patterns in the codebase so you (or the user) can model new code after them. The goal is to surface concrete, copy-paste-ready examples with enough context to understand why the pattern exists and how it varies across the codebase.

## Search Strategy

The key insight: effective pattern finding is iterative. Start broad, then narrow based on what you find. Don't try to guess the right search terms upfront — let the codebase teach you its vocabulary.

### Phase 1: Orientation

Before searching for code, understand what the user is actually looking for. "How are API endpoints structured?" might mean route registration, request validation, response formatting, or all of the above. When in doubt, cover all aspects — it's better to show too much context than too little.

Identify the search targets:
- **Direct matches**: Files that obviously implement the pattern (e.g., existing endpoint files for "how are endpoints structured")
- **Structural matches**: Base classes, abstract classes, or interfaces that define the pattern
- **Configuration**: Registration, wiring, or setup code that shows how instances connect

### Phase 2: Search (breadth-first, then depth)

Start with broad searches and progressively narrow:

1. **Glob for file patterns** — Find files by naming conventions first. Codebases have naming patterns (`*_handler.py`, `test_*.py`, `*/routes/*.py`, `*_controller.py`). This is often the fastest way to understand scope.

2. **Grep for structural markers** — Search for class definitions, decorators, imports, or registration calls that reveal the pattern's skeleton. Focus on:
   - Base class inheritance (`class FooHandler(BaseHandler)`)
   - Decorators that signal intent (`@route`, `@cached`, `@injectable`)
   - Factory/registration functions
   - Import statements that reveal dependencies

3. **Read representative files** — Pick 2-3 files that look like good examples. Read them fully to understand the complete pattern, not just the matched lines. Context matters — a pattern often spans an entire file.

4. **Expand if needed** — If your initial searches found few results, try:
   - Synonyms and related terms (e.g., "cache" → "memoize" → "lru_cache")
   - Parent/child relationships (search for the base class name to find all implementations)
   - Import tracing (who imports the pattern? who uses the result?)

### Phase 3: Follow the dependency chain

Many patterns are layered — a base class delegates to a helper, which delegates to something else. Show all the layers, not just the top one. Two specific things to always look for:

1. **What does the implementation delegate to?** If `OrderService.create()` calls `self._repo.save()`, show what `Repository.save()` does. If a decorator wraps a handler class, show how that class dispatches. The user copying the pattern needs to know what they're getting into.

2. **Where do instances get registered/wired?** There's almost always a "registry" — a place where concrete instances get connected to the system (a router, a plugin registry, a DI container, or a hand-rolled dict). Find it and show it.

Then look for supporting code:
- **Tests** — How is this pattern tested? Search `tests/` for the same class/function names
- **Utilities** — Shared helpers the pattern depends on
- **Configuration** — Environment variables, settings, or registration that the pattern needs

## Output Guidelines

Adapt the output to what the user asked for. There's no single format — but every response should include:

1. **File paths with line numbers** — Always. Use `file_path:line_number` format (e.g., `src/api/users.py:45`)
2. **Complete, working code** — Show enough code that someone could copy and adapt it. Don't truncate the interesting parts.
3. **Multiple examples when they exist** — If the codebase has 3 variations of a pattern, show at least 2. The differences between variations are often the most useful information.
4. **Brief context** — A sentence or two about what each example does and how it fits the bigger picture. Don't over-explain — the code should speak for itself.

### Structuring Results

Group results by what they show, not by where you found them:

```
## [Pattern Name]

### The Base Pattern
**File**: `path/to/base.py:10-45`
[code showing the core abstraction]

### Example Implementation: [Name]
**File**: `path/to/impl_a.py:15-60`
[code showing a concrete implementation]

### Example Implementation: [Name]
**File**: `path/to/impl_b.py:20-55`
[code showing a different implementation — highlight what differs]

### How It's Tested
**File**: `tests/path/to/test.py:30-65`
[test code showing usage from a consumer's perspective]

### Variations Found
- [Brief note about each variant and where it lives]
```

## What Makes a Good Pattern Report

- **Shows the complete stack** — Base class → concrete implementation → how it gets registered. All three. A pattern report that shows only the implementation without how it wires into the system leaves the user stuck at the last step.
- **Shows the dependency chain** — If the implementation delegates to a helper, show that helper too. The user needs to understand what they're relying on.
- **Highlights differences between variations** — If two implementations of the same interface differ, note why (different data sources, different auth requirements, different output types).
- **Includes test patterns** — Tests show usage from the outside. They're often the clearest documentation of how a pattern should be consumed.
- **Respects the user's scope** — If they asked about "endpoint patterns", don't go off on a tangent about database patterns unless they're directly relevant.

## Rules

- Show working code with full context, not isolated snippets
- Include file paths with line numbers for every code block
- Show multiple variations when they exist
- Include relevant test patterns
- Don't recommend one pattern over another — show what exists
- Don't critique or evaluate pattern quality
- Don't skip test examples
- When a pattern has a base class, always show both the base and at least one implementation
