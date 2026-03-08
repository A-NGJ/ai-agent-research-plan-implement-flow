---
description: Investigate the codebase — fact-finding with optional assessment
model: opus
---

# Research Codebase

Conduct research across the codebase to answer user questions by spawning parallel sub-agents and synthesizing their findings.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or use `claude-init` to set it up.

## Facts first, opinions when warranted

This command is the first stage of a pipeline: **research → design → plan → implement**. Its output feeds directly into `/rpi-design`. Your primary job is cartography: map the terrain accurately — describe what exists, where it lives, how components connect, and what patterns are in use.

But good research doesn't stop at raw facts. When your findings reveal clear pain points, opportunities, or trade-offs, include an **Assessment** section with your opinionated take. The key distinction: facts come first and stand on their own. Assessment is clearly separated and optional — include it when the findings warrant it, skip it when they don't.

If you notice something interesting or unusual, document it as an observation in the findings ("this module uses pattern X while the rest of the codebase uses pattern Y"). If you have a view on what should be done about it, put that in the Assessment.

## Step 1: Receive the query

If the user provided a research question or topic as command arguments, proceed directly to Step 2 — no need to prompt them.

If no arguments were provided, ask:
```
What would you like me to research? Provide a question, topic, or file path and I'll investigate thoroughly.

This works for both focused questions ("how does auth work?") and open-ended exploration ("what could we improve about error handling?").
```

## Step 2: Read mentioned files and check for existing research

- If the user mentions specific files (tickets, docs, JSON), read them fully (no limit/offset) before doing anything else. You need this context before you can decompose the research.
- Check for existing research on the same topic: run `rpi scan --type research`
  - If relevant research exists (match by topic), tell the user: "I found existing research on this topic: [path]. Want me to build on it, or start fresh?"
  - If no relevant research exists, continue.

## Step 3: Assess scope and decompose

Not every question needs the same level of investigation. Before spawning sub-agents, assess the query:

- **Focused query** (e.g., "how does the categorization pipeline work?", "where is authentication handled?"): 1-2 targeted sub-agents are enough. Skip the full orchestration — just find the relevant code and explain it.
- **Broad query** (e.g., "document the entire data model", "what's the tech debt situation?"): Full decomposition with multiple parallel sub-agents across different areas.
- **Exploratory query** (e.g., "what could we improve about X?", "is it worth migrating to Y?"): Broad investigation focused on surfacing opportunities and trade-offs. The Assessment section becomes especially important here.

For broad/exploratory queries, break the question into composable research areas:
- Identify specific components, patterns, or concepts to investigate
- Consider which directories, files, or architectural patterns are relevant
- Create a task list to track subtasks

## Step 4: Spawn research sub-agents

Create Task sub-agents to research different aspects concurrently. Each sub-agent should load the appropriate skill, then perform its work.

**For codebase research:**
- Sub-task: "Load the `locate-codebase` skill, then find WHERE files and components live for [topic]"
- Sub-task (@codebase-analyzer): Understand HOW specific code works
- Sub-task: "Load the `find-patterns` skill, then find examples of existing patterns for [topic]"

**For thoughts directory:**
- Sub-task: "Load the `locate-thoughts` skill, then discover what documents exist about [topic]"
- Sub-task: "Load the `analyze-thoughts` skill, then extract key insights from [specific document paths]"

**For web research (optional):**
Only spawn web research sub-agents when:
1. The project is greenfield (little or no existing codebase)
2. The user explicitly asks for it (e.g., "also search the web", "what do the docs say")

When triggered, spawn 1-3 general-purpose sub-agents focused on different angles:
- Official documentation, APIs, and reference material
- Community knowledge — blog posts, Stack Overflow, GitHub issues
- Changelogs, migration guides, version-specific docs (if a library/framework is involved)

Each web research sub-agent should return findings with source URLs as markdown links.

**Skill usage strategy:**
- Start with locate skills (`locate-codebase`, `locate-thoughts`) to find what exists
- Then use analyzer skills (`analyze-thoughts`, `@codebase-analyzer`) on the most promising findings
- Run sub-agents in parallel when they're searching for different things
- For focused queries, you may only need 1-2 of these — use judgment

## Step 5: Checkpoint (broad queries only)

For broad research, after the initial locate sub-agents return but before deep analysis:

Present a brief summary of what was found and ask the user if they want to redirect focus:
```
Initial scan found these areas related to [topic]:
- [Area 1]: N files in path/to/...
- [Area 2]: N files in path/to/...
- [Area 3]: N files in path/to/...

Should I dig into all of these, or focus on specific areas?
```

Skip this checkpoint for focused queries — just proceed to synthesis.

## Step 6: Synthesize findings

Wait for all sub-agents to complete, then:

- Compile all results
- Prioritize live codebase findings as the primary source of truth
- Use .thoughts/ findings as supplementary historical context
- If web research was performed, connect codebase findings to external docs and patterns
- Connect findings across different components
- Include specific file paths and line numbers
- Answer the user's specific questions with concrete evidence

## Step 7: Write the research document

Ask the user if they want to save the findings:
```
Want me to save these findings to `.thoughts/research/YYYY-MM-DD-description.md`?
```

If the user declines, that's fine — research can be purely conversational. If they agree (or if the research is substantial enough that saving is clearly useful), write the document.

**Create the file**: Run `rpi scaffold research --topic "..." --write`
This creates `.thoughts/research/YYYY-MM-DD-description.md` with frontmatter pre-populated (date, researcher, git_commit, branch, repository, topic, tags, status).

**Fill in the document sections** — only include sections that have actual findings, skip empty sections rather than leaving placeholders:
- Research Question
- Summary (high-level answer)
- Detailed Findings (with file:line references)
- Code References
- Architecture (patterns, conventions found)
- Assessment (when findings warrant it — pain points, opportunities, quick wins)
- Suggested Next Steps (when assessment points to clear actions)
- Web Research Findings (only if web research was performed)
- Historical Context (only if .thoughts/ had relevant documents)
- Open Questions

## Step 8: Present and follow up

- Present a concise summary of findings to the user
- Include key file references for easy navigation
- If the findings warrant it, include your assessment and suggested next steps
- Ask if they have follow-up questions

**For follow-ups:** Append to the same research document rather than creating a new one. Add a `## Follow-up: [brief description]` section and update the frontmatter:
```
rpi frontmatter set <file> last_updated "<YYYY-MM-DD>"
rpi frontmatter set <file> last_updated_by "<Researcher name>"
rpi frontmatter set <file> last_updated_note "Added follow-up research for [brief description]"
```

## Step 9: Update specs (optional)

If the research revealed behavioral details about a module or domain, check `.thoughts/specs/` for an existing spec file:

- **If a spec exists** and the research contains new or corrected information, offer to update it
- **If no spec exists** and the research is comprehensive enough to describe a module's behavior, offer to create one

Ask: "This research documents [module] behavior. Want me to create/update a spec at `.thoughts/specs/[domain].md`?"

**To create a spec**: Run `rpi scaffold spec --topic "..." --write`
This creates the spec file with frontmatter pre-populated. Fill in the sections: Purpose, Behavior, Key Components, Interfaces, Constraints.

**Rules:**
- Only update specs with confirmed facts from the research — not assumptions or recommendations
- Never force spec creation — this step is always optional
- Specs describe current behavior, not planned changes

## Guidelines

- Use parallel Task agents to maximize efficiency
- Always run fresh codebase research — existing .thoughts/ documents supplement but don't replace live findings
- Focus on concrete file paths and line numbers
- Research documents should be self-contained
- Keep the main agent focused on synthesis; delegate deep file reading to sub-agents
- Explore all of .thoughts/, not just the research subdirectory
- Read mentioned files fully before spawning sub-tasks
- Wait for all sub-agents to complete before synthesizing
- Never write the research document with placeholder values
- Use snake_case for multi-word frontmatter fields
- Scale investigation effort to the topic — a simple question doesn't need five sub-agents
- Be opinionated in the Assessment section — that's where your view belongs
- Keep facts and opinions clearly separated — facts in Findings, opinions in Assessment
