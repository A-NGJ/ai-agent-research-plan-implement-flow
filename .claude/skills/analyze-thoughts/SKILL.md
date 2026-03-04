---
name: analyze-thoughts
description: Extract high-value insights from thoughts documents by filtering aggressively for decisions, constraints, technical specs, and actionable information. Use when you need to deeply analyze a specific document from the .thoughts/ directory.
---

## What I Do

Deeply analyze thoughts documents to extract only the most relevant, actionable information. I'm a curator of insights, not a document summarizer — I filter aggressively to return what actually matters.

## Analysis Strategy

### Step 1: Read with Purpose
- Read the entire document first
- Identify the document's main goal
- Note the date and context
- Understand what question it was answering

### Step 2: Extract Strategically
Focus on finding:
- **Decisions made**: "We decided to..."
- **Trade-offs analyzed**: "X vs Y because..."
- **Constraints identified**: "We must..." / "We cannot..."
- **Lessons learned**: "We discovered that..."
- **Action items**: "Next steps..." / "TODO..."
- **Technical specifications**: Specific values, configs, approaches

### Step 3: Filter Ruthlessly
Remove:
- Exploratory rambling without conclusions
- Options that were rejected (unless the rejection reasoning is valuable)
- Temporary workarounds that were replaced
- Personal opinions without backing
- Information superseded by newer documents

## Output Format

```
## Analysis of: [Document Path]

### Document Context
- **Date**: [When written]
- **Purpose**: [Why this document exists]
- **Status**: [Still relevant / implemented / superseded?]

### Key Decisions
1. **[Decision Topic]**: [Specific decision made]
   - Rationale: [Why this decision]
   - Impact: [What this enables/prevents]

### Critical Constraints
- **[Constraint Type]**: [Specific limitation and why]

### Technical Specifications
- [Specific config/value/approach decided]

### Actionable Insights
- [Something that should guide current implementation]
- [Pattern or approach to follow/avoid]

### Still Open/Unclear
- [Questions that weren't resolved]

### Relevance Assessment
[1-2 sentences on whether this information is still applicable and why]
```

## Quality Filters

### Include Only If:
- It answers a specific question
- It documents a firm decision
- It reveals a non-obvious constraint
- It provides concrete technical details
- It warns about a real gotcha/issue

### Exclude If:
- It's just exploring possibilities without conclusion
- It's personal musing without conclusion
- It's been clearly superseded
- It's too vague to action
- It's redundant with better sources

## Rules

- Be skeptical — not everything written is valuable
- Think about current context — is this still relevant?
- Extract specifics — vague insights aren't actionable
- Note temporal context — when was this true?
- Highlight decisions — these are usually most valuable
