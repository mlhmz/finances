# Playwright E2E Tester Memory - Finances App

## Project Structure
- **Test location**: `/Users/malek/Code/finances/e2e/transactions.spec.ts`
- **Config**: `/Users/malek/Code/finances/playwright.config.ts`
- **Running tests**: `TEST_MODE=1 npx playwright test e2e/transactions.spec.ts --reporter=line`
- **App runs on**: `http://localhost:3000`

## Key Patterns & Conventions

### Authentication & Setup
- Use `registerAndLogin(request, email)` helper with unique email: `e2e-{test-name}-${TS}@example.com`
- TEST_OTP is always `"00000000"` for verification
- Must use `request` fixture (API context) for stateful auth flows
- Session/cookies persist across requests within same test

### Test Structure
- All tests use API-level requests (no browser) for speed
- Tests import: `expect, test` from `@playwright/test`
- Timestamp `TS = Date.now()` available for unique identifiers
- Tests are independent and can run in any order

### Transaction Endpoints & Headers
- **GET /transactions/new**: Returns blank form; sets `HX-Trigger: openTransactionForm`
- **GET /transactions/:id/edit**: Returns pre-filled form; sets `HX-Trigger: openTransactionForm` (added in latest sprint)
- **POST /transactions**: Success sets `HX-Trigger: closeTransactionForm`, `HX-Retarget: #transaction-list`, `HX-Reswap: afterbegin`
- **PUT /transactions/:id**: Success sets `HX-Trigger: closeTransactionForm`, `HX-Retarget: #tx-{id}`, `HX-Reswap: outerHTML`

### Form Markup & Content
- Form partial: `partials/transaction_form`
- Form fields (always present): `name="title"`, `name="amount"`, `name="currency"`, `name="date"`, `name="description"`
- New form: Contains "New Transaction" heading
- Edit form: Contains "Edit Transaction" heading

### Common Test Assertions
- Response status: `expect(res.status()).toBe(200)`
- Headers: `expect(res.headers()["hx-trigger"]).toContain("openTransactionForm")`
- Content: `expect(body).toContain("text")` or `expect(body).not.toContain("text")`
- Transaction ID extraction: `/id="tx-([^"]+)"/` regex from list page

### Data Isolation
- All transaction data is per-user (isolation tests exist and pass)
- PUT/DELETE on other user's transactions returns 404 or no-op

## Test Coverage (31 tests total)
- Auth guard (1 test)
- Transactions page (3 tests)
- Create transaction (6 tests)
- Amount storage (2 tests)
- **New transaction form (4 tests)** - ADDED
- Edit transaction (5 tests, including HX-Trigger verification)
- Delete transaction (4 tests)
- Data isolation (1 test)
- Pagination (3 tests)

## Important Notes
- Use `TEST_MODE=1` environment variable when running tests (config will auto-start server)
- Form validation errors return 200 status with form + error messages
- Blank form means no pre-fill: no specific title/amount visible
- All new tests follow existing naming and structure conventions
