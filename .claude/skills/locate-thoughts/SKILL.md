---
name: locate-thoughts
description: Search the .thoughts/ directory to discover relevant documents (tickets, research, plans, PRs, notes). Returns organized file listings grouped by type. Use when you need to find historical context or documentation.
---

## What I Do

Search the .thoughts/ directory structure to find relevant documents and categorize them by type. This is a discovery tool — find where documents live, don't analyze their contents.

## Directory Structure

```
.thoughts/
├── research/      # Research documents
├── plans/         # Implementation plans
├── designs/       # Design documents
├── structures/    # Structure documents
├── tickets/       # Ticket documentation
├── prs/           # PR descriptions
├── reviews/       # Code review reports
```

## Search Strategy

1. **Think about search terms first** — consider synonyms, technical terms, component names, and related concepts
2. **Use grep** for content searching across .thoughts/
3. **Use glob** for filename patterns (e.g., `.thoughts/**/*rate-limit*`)
4. **Check all subdirectories**

## Search Patterns

- Ticket files: often named `eng_XXXX.md` or `ENG-XXXX-description.md`
- Research files: often dated `YYYY-MM-DD-topic.md`
- Plan files: often named `YYYY-MM-DD-feature-name.md`
- Design files: similar dating convention
- PR descriptions: often `{number}_description.md`
- Review reports: often `YYYY-MM-DD-branch-name-code-review-report.md`

## Output Format

```
## Thought Documents about [Topic]

### Tickets
- `.thoughts/tickets/eng_1234.md` - Brief description from title

### Research Documents
- `.thoughts/research/2024-01-15-topic.md` - Brief description

### Implementation Plans
- `.thoughts/plans/2024-01-15-topic.md` - Brief description

### Designs
- `.thoughts/designs/2024-01-15-topic.md` - Brief description

### Reviews
- `.thoughts/reviews/2024-01-15-branch-review.md` - Brief description

Total: N relevant documents found
```

## Rules

- Don't read full file contents — just scan for relevance
- Be thorough — check all subdirectories
- Group logically — make categories meaningful
- Note date patterns in filenames
- Use multiple search terms (technical terms, component names, related concepts)
