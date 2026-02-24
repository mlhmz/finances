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

### Step 5 — Unit tests

Write unit tests for all non-trivial logic. Run them:

```bash
go test ./...
```

Append the output under **Test Results**. If tests fail, fix the code or tests and re-run. Repeat until all pass.

### Step 6 — Playwright E2E tests

Write Playwright tests in `e2e/` that exercise the feature through the browser.
Test files follow the pattern `e2e/<feature>.spec.ts`.

Run all E2E tests:

```bash
npx playwright test
```

Useful flags:

```bash
npx playwright test --headed          # watch tests run in the browser
npx playwright test --ui              # interactive Playwright UI
npx playwright test e2e/feature.spec.ts  # run a single spec
npx playwright show-report            # open the HTML report after a run
```

> **Playwright MCP (if available):** When the Playwright MCP server is connected,
> prefer using it to interact with and inspect the running app directly instead of
> manually writing selectors. If Playwright tests are hard to get passing — e.g.
> selectors can't be found or async timing is flaky — fall back to the Playwright
> MCP to explore the live DOM and diagnose the issue before retrying.

Append the output under **Test Results**. If tests fail, fix the code or tests and re-run. Repeat until all pass.

### Step 7 — Verify

Do a final check:
- All plan items are `[x]`
- `go test ./...` passes
- `npx playwright test` passes
- Feature behaves as described in the spec's Acceptance Criteria

Update the progress file status to `## Status: Done`.

### Step 8 — Summary

Tell the user:
- What was implemented
- Test results summary
- Path to the progress file
- Any remaining blockers or deferred decisions (add to **Blockers** in the progress file and ask the user)
