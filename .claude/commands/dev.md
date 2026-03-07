# Skill: /dev

Implement a feature end-to-end: code, unit tests, Playwright E2E tests — iterate until everything passes.

## Usage

```
/dev <feature-number>
```

## Instructions

### Step 1 — Find the spec

Search for `/spec/{nr}_*_spec.md` matching the given feature number. Read it fully. If it does not exist, tell the user and stop.

Derive the `{summary}` slug from the filename for use in the progress file path.

Progress file: `/spec/{nr}_{summary}_progress.md`

### Step 2 — Create the progress file

Create `/spec/{nr}_{summary}_progress.md` with this structure:

```markdown
# Feature {nr}: {Title} — Progress

## Status: In Progress

## Plan
[ tasks derived from the spec, filled in after reading it ]

## Implementation Log
[ append entries as work progresses ]

## Test Results
[ append test run outputs ]

## Blockers
[ anything that needs user input ]
```

### Step 3 — Plan

Break the spec into concrete tasks. Write them as a checklist under **Plan** in the progress file:

```
- [ ] Data model / migrations
- [ ] Route handlers / business logic
- [ ] Unit tests
- [ ] Playwright E2E tests
- [ ] Manual verification
```

Adjust the list to match what the spec actually requires.

### Step 4 — Implement

Work through each task in order. For each task:
1. Write the code following existing project conventions.
2. Mark the task `[x]` in the progress file when done.
3. Append a brief log entry under **Implementation Log**:
   ```
   ### {task name}
   - Files changed: ...
   - Notes: ...
   ```

> **gopls LSP (Go files):** The `gopls-lsp` plugin provides live Go code intelligence. Use it for diagnostics, hover info, go-to-definition, and refactoring hints while editing `.go` files. If gopls flags a type error or unused import, fix it before moving on.

> **Frontend / UI tasks (views/):** When implementing HTML templates or any UI work, invoke the `frontend-designer` agent (via the Task tool with `subagent_type: frontend-designer`). It produces production-grade, visually distinctive interfaces using HTMX, Web Awesome, and Go Fiber templates — not generic boilerplate. Use it whenever you touch files under `views/`.

### Step 5 — Unit tests

Invoke the `unit-test-writer` agent (via the Task tool with `subagent_type: unit-test-writer`) to write comprehensive unit tests for all non-trivial logic added in this feature. The agent will:
- Identify recently changed handlers, models, and logic
- Write table-driven tests with happy paths and error cases
- Use in-memory SQLite for database tests
- Run `go test ./...` and iterate until all tests pass

Append the agent's test results output under **Test Results**.

### Step 6 — Simplify

Once unit tests pass, invoke the `backend-code-simplifier` agent (via the Task tool with `subagent_type: backend-code-simplifier`) to refine all backend code modified during implementation. The agent will improve clarity, remove redundancy, and enforce Go/Fiber/GORM project conventions — without changing behaviour.

If the agent proposes any changes, re-run `go test ./...` to confirm everything still passes before continuing.

### Step 7 — Playwright E2E tests

Invoke the `playwright-e2e-tester` agent (via the Task tool with `subagent_type: playwright-e2e-tester`) to write and run end-to-end tests for the feature. The agent will:
- Analyse the user flows and HTMX interactions introduced by the feature
- Write Playwright tests in `e2e/<feature>.spec.ts`
- Handle Web Awesome custom element selectors and HTMX async waits correctly
- Run all E2E tests and iterate on failures until the suite is green

> **Playwright MCP (if available):** When the Playwright MCP server is connected,
> the agent should prefer using it to interact with and inspect the running app directly
> instead of manually writing selectors. If tests are hard to get passing, the MCP can
> explore the live DOM and diagnose selector or timing issues.

Append the agent's test results output under **Test Results**.

### Step 8 — Verify

Do a final check:
- All plan items are `[x]`
- `go test ./...` passes
- `npx playwright test` passes
- Feature behaves as described in the spec's Acceptance Criteria

Update the progress file status to `## Status: Done`.

### Step 9 — Summary

Tell the user:
- What was implemented
- Test results summary
- Path to the progress file
- Any remaining blockers or deferred decisions (add to **Blockers** in the progress file and ask the user)
