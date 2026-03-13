---
description: Verify implementation against proposal artifacts for completeness, correctness, and coherence
model: opus
---

# Verify Implementation

Validate that an implementation matches its proposal artifacts across three dimensions: completeness, correctness, and coherence. Produces a severity-classified verification report.

This command is purely advisory — it does not block anything. It can be re-run after fixes to confirm resolution.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`.
Run `rpi --help` to discover available commands and `rpi <command> --help` for detailed usage with examples.

## Step 1: Receive the input

If the user provided a path to a proposal, plan, or other artifact as command arguments, proceed to Step 3.

If no arguments were provided, proceed to Step 2 (auto-detection).

## Step 2: Auto-detect what to verify

When no path is provided, detect from recent git changes:

1. Run in parallel using `rpi`:
   - Get the list of changed files on the current branch (falls back to last 5 commits if on main)
   - Find active plans
   - Find active proposals
2. If artifacts found, announce:
   ```
   Found active plan at [path] — verifying against it.
   ```
   Then proceed to Step 3 with the discovered path(s).
3. If nothing found, ask:
   ```
   I couldn't detect an active plan or proposal from recent git changes.

   What would you like me to verify? Provide a path to:
   - A plan: `.thoughts/plans/YYYY-MM-DD-description.md`
   - A proposal: `.thoughts/proposals/YYYY-MM-DD-description.md`
   ```

## Step 3: Read referenced artifacts

Read the provided or detected artifact(s) fully. Then use `rpi` to resolve the artifact chain.

This returns the full chain (plan → proposal → research) with metadata. Read the files it identifies.

Also check:
- `.thoughts/specs/` — if this directory exists, read any specs relevant to the changed domain areas
- Use `rpi` to get the list of changed files to identify which files were actually changed in the implementation

Present a summary before proceeding:

```
Verifying against:
- Plan: [path] (if found)
- Proposal: [path] (if found)
- Specs: [paths] (if found)

Changed files: [N files from git diff]

Spawning verification sub-agents...
```

## Step 4: Verify across three dimensions

Spawn 3 sub-agents in parallel, one per dimension. Each sub-agent receives the full context from Step 3.

### Sub-agent 1: Completeness

Check whether everything that was planned has been done.

- Sub-task: "Verify completeness of the implementation. You are checking whether all planned work was actually completed.

  Context:
  - [Include the plan content, ticket content, and list of changed files]

  Use `rpi` for mechanical checks:
  - Run completeness checks on the plan — returns checkbox counts (checked vs total) and file coverage (planned files vs actually changed files)
  - Run marker scans on changed files — scans for TODO/FIXME/HACK markers

  Then check each of these using your own judgment:
  1. **Plan phases**: Are all phases checked off? List any unchecked items.
  2. **Acceptance criteria**: If a ticket exists, are all acceptance criteria met? Read the actual implementation files to verify — don't trust checkboxes alone.
  3. **TODO/FIXME/HACK markers**: Report any markers found by the marker scan in new or modified code.
  4. **Test coverage**: Do tests exist for new functionality? Check that new public functions, components, or endpoints have corresponding test files or test cases.
  5. **File coverage**: Were all files mentioned in the plan actually created or modified? Use the file coverage output from the completeness check.

  For each item, report:
  - PASS: [what was satisfied]
  - MISSING (BLOCKER): [critical gap]
  - MISSING (WARNING): [non-critical gap]

  Return your findings as a structured list."

### Sub-agent 2: Correctness

Check whether the implementation matches the design intent.

- Sub-task: "Verify correctness of the implementation against the proposal. You are checking whether what was built matches what was specified.

  Context:
  - [Include the proposal content, plan content, specs content if available, and list of changed files]

  Check each of these:
  1. **Approach alignment**: Does the implementation follow the chosen approach from the proposal? Read the key implementation files and compare against proposal decisions.
  2. **API contracts**: If the proposal specifies interfaces, APIs, or data contracts, verify the implementation matches. Check function signatures, data shapes, and endpoint definitions.
  3. **Edge cases**: Were edge cases identified in the proposal actually handled in code? Read the relevant code sections to verify.
  4. **Silent deviations**: Are there files changed that weren't in the plan? Are there approaches that differ from what was proposed? Flag any undocumented divergence.
  5. **Spec consistency**: If `.thoughts/specs/` files were provided, verify the implementation is consistent with the behavioral specs described there.

  For each item, report:
  - PASS: [what matches]
  - DEVIATION (BLOCKER): [critical mismatch]
  - DEVIATION (WARNING): [minor mismatch, may be intentional]
  - DEVIATION (NOTE): [trivial difference, likely fine]

  Return your findings as a structured list."

### Sub-agent 3: Coherence

Check whether the implementation fits the existing codebase.

- Sub-task: "Verify coherence of the implementation with the existing codebase. You are checking whether the new code fits in naturally.

  Context:
  - [Include the list of changed files and their contents]

  Find existing code patterns in the codebase to compare against, then check each of these:
  1. **Naming conventions**: Do new files, functions, variables, and classes follow the naming patterns used in the rest of the codebase? Find similar existing code to compare.
  2. **Error handling**: Does error handling follow established patterns? Compare with similar code in the project.
  3. **Code reuse**: Does the new code use existing utilities, helpers, or shared code rather than reinventing? Check for duplication with existing code.
  4. **Dependencies**: Were any unnecessary dependencies introduced? Check for new imports or packages that duplicate existing capabilities.
  5. **Test patterns**: Do the new tests follow the existing test style? Compare with nearby or similar test files for assertion style, setup patterns, and organization.
  6. **File organization**: Do new files follow the project's conventions for directory structure, imports, and module boundaries?

  For each item, report:
  - PASS: [what's consistent]
  - INCONSISTENCY (WARNING): [pattern mismatch]
  - INCONSISTENCY (NOTE): [minor style difference]

  Return your findings as a structured list."

## Step 5: Synthesize verification report

After all three sub-agents complete, synthesize their findings into a single report.

1. **Determine overall status**:
   - **Pass**: No blockers, no warnings (or only notes)
   - **Pass with warnings**: No blockers, but some warnings exist
   - **Issues found**: At least one blocker exists

2. **Classify all findings by severity**:
   - **Blockers (must fix)**: Missing acceptance criteria, critical design deviations, broken contracts
   - **Warnings (should fix)**: Incomplete test coverage, minor deviations, inconsistent patterns
   - **Notes (consider fixing)**: Style nits, minor naming mismatches, suggestions

3. **Write the report**: Use `rpi` to scaffold and save a verification report for this feature.
   This creates the report file at `.thoughts/reviews/` with frontmatter and section headers. Fill in each section with the sub-agent findings.

4. **Present the summary** to the user:
   ```
   Verification complete: [Pass / Pass with warnings / Issues found]

   - Completeness: [pass / warnings / issues]
   - Correctness: [pass / warnings / issues]
   - Coherence: [pass / warnings / issues]

   [N blockers, N warnings, N notes]

   Full report: .thoughts/reviews/YYYY-MM-DD-verify-[topic].md
   ```

   If blockers exist, list them directly in the summary so the user sees them immediately.

## Guidelines

- **Purely advisory** — this command never blocks anything. Report findings and let the user decide.
- **Re-runnable** — can be run again after fixes. Each run produces a new report file.
- **Works standalone** — does not require the full RPI pipeline to have been used. Can verify any code against any artifact.
- **Read actual code** — don't trust checkboxes or summaries. Sub-agents must read the implementation files to verify claims.
- **Be specific** — every finding should include a `file:line` reference where possible. Vague findings are not actionable.
- **Severity matters** — distinguish genuine blockers from style nits. Don't inflate severity.
- **Check specs** — if `.thoughts/specs/` exists and contains relevant specs, verify behavioral consistency.
- **Scale effort** — if the implementation is small (1-2 files), the sub-agents can be lighter. If it spans many files, be thorough.
