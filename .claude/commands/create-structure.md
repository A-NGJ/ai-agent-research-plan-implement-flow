---
description: Define file structure, module boundaries, and interfaces — use for greenfield projects or major reorganizations where file layout needs dedicated attention
model: opus
---

# Structure Definition

Map a design to concrete codebase structure — files, modules, interfaces, and dependencies.

**When to use this command**: For greenfield projects creating many new files, major module reorganizations, or when the file layout itself is complex enough to warrant a dedicated discussion. For most features, the design doc's optional "File Structure" section (via `/create-design`) and the plan's file listings (via `/create-plan`) provide sufficient coverage without a separate structure stage.

**This command OWNS**: file layout, module boundaries, public interfaces/contracts, dependency graph, and naming conventions.

**This command does NOT own**: codebase research (→ `/create-research`), architectural decisions (→ `/create-design`), or implementation code and phasing (→ `/create-plan`). Define *what* each file exports — the plan stage defines *how* to implement the internals.

## Initial Response

If a file path or design document was provided, read it fully and begin analysis. Otherwise respond:
```
I'll help you define the structure for your implementation.

Please provide:
1. The design document from `.thoughts/designs/`
2. Any additional context or preferences about code organization

Tip: `/create-structure .thoughts/designs/2025-01-08-authentication-flow.md`
```

## Process Steps

### Step 1: Context Gathering

1. **Read the design document fully** (no limit/offset) before spawning sub-tasks
2. **Read any linked research documents** referenced in the design
3. **Spawn parallel research sub-tasks** using the Task tool. Each sub-task should load the appropriate skill first, then perform its work:
   - Sub-task: "Load the `locate-codebase` skill, then find existing files/modules that will be affected by [feature]"
   - Sub-task (@codebase-analyzer): Understand current file organization, module patterns, and naming conventions
   - Sub-task: "Load the `find-patterns` skill, then find how similar features are structured in the codebase (directory layout, file naming, export patterns) for [feature]"
   - Sub-task: "Load the `locate-thoughts` skill, then find any existing structure docs or conventions documentation about [topic]"
4. **Wait for ALL sub-tasks to complete**
5. **Read all files identified by research tasks** - especially module entry points, index files, and existing interfaces
6. **Present structural analysis:**
   ```
   Based on the design and codebase analysis:

   Existing structure relevant to this work:
   - [directory/module] - [what it contains, how it's organized]
   - [file pattern] - [convention used, e.g., "one component per file with co-located test"]

   Conventions I've identified:
   - Naming: [file naming pattern, e.g., kebab-case, PascalCase]
   - Modules: [how modules are organized, e.g., barrel exports, flat structure]
   - Tests: [where tests live, naming convention]
   - Types: [where types/interfaces are defined]

   Questions before I propose the structure:
   - [Question about organization preference or convention]
   ```

### Step 2: Structure Proposal

After clarifications:

1. **Create a todo list** using TodoWrite to track structural decisions
2. **Map each design component to concrete files:**
   - Which existing files need modification?
   - What new files need to be created?
   - What's the dependency direction between modules?
3. **Present the proposed structure:**
   ```
   ## Proposed File Structure

   ### Modified Files:
   - `path/to/existing-file.ext` - [what changes and why]
   - `path/to/another-file.ext` - [what changes and why]

   ### New Files:
   - `path/to/new-file.ext` - [responsibility, what it exports]
   - `path/to/new-test-file.ext` - [what it tests]

   ### Module Boundaries:
   - [Module A] depends on [Module B] via [interface/contract]
   - [Module C] is independent, no new dependencies

   ### Dependency Direction:
   [Module A] -> [Module B] -> [Module C]
   (No circular dependencies)

   Does this organization make sense?
   ```

### Step 3: Interface Definition

After structural buy-in:

1. **Spawn parallel sub-tasks** if needed:
   - Sub-task: "Load the `find-patterns` skill, then find existing interface/type patterns to match for [component]"
   - Sub-task (@codebase-analyzer): Understand existing function signatures at integration points
2. **Define the key interfaces and contracts:**
   ```
   ## Key Interfaces

   ### [Interface/Contract 1]
   Location: `path/to/file.ext`
   ```[language]
   // The public API surface this module exposes
   [type definition or function signature]
   ```

   ### [Interface/Contract 2]
   Location: `path/to/file.ext`
   ```[language]
   [type definition or function signature]
   ```

   ### Integration Points:
   - [Existing file:line] will call [new interface] for [purpose]
   - [New file] will import from [existing module] for [purpose]

   Do these interfaces look right?
   ```

### Step 4: Write the Structure Document

Save to `.thoughts/structures/YYYY-MM-DD-ENG-XXXX-description.md`
- Format: `YYYY-MM-DD-ENG-XXXX-description.md`
- Without ticket: `2025-01-08-improve-error-handling.md`

**Template:**

````markdown
---
date: [Current date and time with timezone in ISO format]
author: [Author name]
git_commit: [Current commit hash]
branch: [Current branch name]
repository: [Repository name]
topic: "[Feature/Task Name]"
tags: [structure, modules, relevant-component-names]
status: draft
related_design: [path to design doc]
related_research: [path to research doc if available]
last_updated: [Current date in YYYY-MM-DD format]
last_updated_by: [Author name]
---

# Structure: [Feature/Task Name]

## Overview
[Brief description of the structural changes needed to implement the design]

## Context
- **Design**: [Link to design document]
- **Research**: [Link to research document if available]

## Conventions Applied
- **File naming**: [Convention followed, e.g., kebab-case.ts]
- **Module pattern**: [Pattern followed, e.g., barrel exports via index.ts]
- **Test placement**: [Convention followed, e.g., co-located __tests__/ directory]
- **Type definitions**: [Convention followed, e.g., types.ts per module]

## File Changes

### Modified Files

#### `path/to/existing-file.ext`
- **Current role**: [What this file does now]
- **Changes needed**: [What will be added/modified]
- **Reason**: [Why this file needs to change, tied to design decision]

#### `path/to/another-file.ext`
[Same structure...]

### New Files

#### `path/to/new-file.ext`
- **Responsibility**: [Single responsibility description]
- **Exports**: [Key exports - functions, types, constants]
- **Depends on**: [What it imports from]
- **Depended on by**: [What will import from it]

#### `path/to/new-test-file.ext`
[Same structure...]

## Module Boundaries

### [Module/Component Name]
- **Directory**: `path/to/module/`
- **Public API**: [What it exposes to other modules]
- **Internal**: [What stays private to the module]
- **Dependencies**: [External modules it depends on]

## Key Interfaces

### [Interface Name]
**Location**: `path/to/file.ext`
**Used by**: [Components that depend on this interface]

```[language]
// Interface definition
[type definition, function signature, or contract]
```

### [Interface Name 2]
[Same structure...]

## Dependency Graph

```
[Visual representation of module dependencies]
[Module A] --> [Module B] --> [Module C]
                           --> [Module D]
```

**Dependency rules:**
- [Rule, e.g., "UI components never import directly from data layer"]
- [Rule, e.g., "All cross-module communication goes through the service interface"]

## Migration Notes (if modifying existing structure)
- [Step to safely migrate, e.g., "Add new interface alongside old one before removing"]
- [Backward compatibility consideration]

## References
- Design: `[path to design doc]`
- Research: `[path to research doc]`
- Similar structure: `[existing module path used as reference]`
````

### Step 5: Review & Iterate

1. Present the draft structure document location
2. Iterate based on feedback
3. Ensure all interfaces are concrete enough for the planning stage
4. Mark status as `complete` when finalized

## Guidelines

1. **Follow Existing Conventions**: Match the codebase's naming, organization, and patterns exactly
2. **Be Interactive**: Get buy-in on file layout before defining interfaces
3. **Be Concrete**: Every file should have a clear responsibility and defined exports
4. **Be Minimal**: Don't create files or abstractions that aren't justified by the design
5. **No Circular Dependencies**: Ensure the dependency graph is a DAG
6. **No Open Questions**: Resolve all structural questions before finalizing

## Common Structure Patterns

**Feature Module:** `feature/index.ts` (barrel) + `feature/types.ts` + `feature/[component].ts` + `feature/__tests__/`

**API Endpoint:** Route definition + handler + validation + tests (mirror existing endpoint structure)

**Refactoring:** New structure alongside old -> Migrate consumers -> Remove old structure
