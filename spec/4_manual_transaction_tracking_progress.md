# Feature 4: Manual Transaction Tracking — Progress

## Status: Done

## Plan

- [x] Data model: `internal/models/transaction.go`
- [x] Export `ParseDecimal` from `internal/money/money.go`
- [x] Repository: `internal/repository/transaction.go`
- [x] DB migration: add `Transaction` to `AutoMigrate` in `internal/db/db.go`
- [x] Handlers: `internal/handlers/transaction.go`
- [x] Routes + template functions in `cmd/finances/main.go`
- [x] Views: `views/transactions.html`, `views/partials/transaction_row.html`,
         `views/partials/transaction_form.html`, `views/partials/transaction_confirm_delete.html`
- [x] Unit tests: `internal/repository/transaction_test.go`
- [x] Playwright E2E tests: `e2e/transactions.spec.ts`
- [x] Manual verification

## Implementation Log

### Data model
- Files changed: `internal/models/transaction.go`
- Notes: Transaction struct with money.Money embedded, UserID scoped

### Export ParseDecimal
- Files changed: `internal/money/money.go`
- Notes: Exposed `parseDecimal` as `ParseDecimal` for use in handlers

### Repository
- Files changed: `internal/repository/transaction.go`
- Notes: Full CRUD scoped by userID; uses uuid for ID generation

### DB migration
- Files changed: `internal/db/db.go`
- Notes: Added models.Transaction to AutoMigrate

### Handlers
- Files changed: `internal/handlers/transaction.go`
- Notes: 7 handlers covering CRUD + confirm-delete + row partial; HTMX OOB swaps for create/update

### Routes + template functions
- Files changed: `cmd/finances/main.go`
- Notes: Added 7 transaction routes under protected group; template funcs: fmtAmountDisplay, isIncome, fmtDate, fmtDateTimeInput, absAmountStr

### Views
- Files changed: `views/transactions.html`, `views/partials/transaction_row.html`,
  `views/partials/transaction_form.html`, `views/partials/transaction_confirm_delete.html`
- Notes: Responsive layout (sidebar desktop, bottom tab bar mobile); HTMX interactions for create/edit/delete; CSS modal overlay (desktop center, mobile bottom drawer)

### Unit tests
- Files changed: `internal/repository/transaction_test.go`
- Notes: Tests for List, Create, GetByID, Update, Delete with isolation checks

## Test Results

### Unit tests
```
ok  github.com/mlhmz/finances/internal/repository  0.007s
ok  github.com/mlhmz/finances/internal/money       0.001s
ok  github.com/mlhmz/finances/internal/models      0.001s
ok  github.com/mlhmz/finances/internal/auth        (cached)
ok  github.com/mlhmz/finances/internal/currency    (cached)
All packages: PASS
```

### E2E tests
```
26 passed (3.4s) — e2e/transactions.spec.ts
All 26 transaction tests pass.
(12 pre-existing browser-binary failures in auth/home/profile/currency suites unaffected)
```

## Blockers
None
