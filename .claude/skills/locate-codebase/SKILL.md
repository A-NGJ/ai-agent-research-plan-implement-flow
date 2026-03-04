---
name: locate-codebase
description: Find files, directories, and components relevant to a feature or task. Returns organized file listings grouped by purpose (implementation, tests, config, types, docs). Use when you need to discover where code lives in the codebase.
---

## What I Do

Find WHERE code lives in a codebase and organize findings by purpose. This is a file discovery tool — locate relevant files, don't analyze their contents.

## Search Strategy

### Initial Broad Search
1. Think about effective search patterns — naming conventions, related terms, synonyms
2. Use grep for finding keywords in file contents
3. Use glob for file name patterns
4. Use list to explore directory structures

### Refine by Language/Framework
- **JavaScript/TypeScript**: src/, lib/, components/, pages/, api/
- **Python**: src/, lib/, pkg/, module names matching feature
- **Go**: pkg/, internal/, cmd/
- **General**: Check for feature-specific directories

### Common Patterns to Find
- `*service*`, `*handler*`, `*controller*` — Business logic
- `*test*`, `*spec*` — Test files
- `*.config.*`, `*rc*` — Configuration
- `*.d.ts`, `*.types.*` — Type definitions
- `README*`, `*.md` in feature dirs — Documentation

## Output Format

```
## File Locations for [Feature/Topic]

### Implementation Files
- `src/services/feature.js` - Main service logic
- `src/handlers/feature-handler.js` - Request handling

### Test Files
- `src/services/__tests__/feature.test.js` - Service tests
- `e2e/feature.spec.js` - End-to-end tests

### Configuration
- `config/feature.json` - Feature-specific config

### Type Definitions
- `types/feature.d.ts` - TypeScript definitions

### Related Directories
- `src/services/feature/` - Contains N related files

### Entry Points
- `src/index.js` - Imports feature module at line 23
```

## Rules

- Don't read file contents in depth — just report locations
- Be thorough — check multiple naming patterns and extensions
- Group logically — make it easy to understand code organization
- Include counts — "Contains X files" for directories
- Note naming patterns — help understand conventions
- Provide full paths from repository root
- Don't analyze what the code does — just find where things are
- Don't skip test or config files
- Don't make assumptions about functionality
