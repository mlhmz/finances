# Feature 2: Multi-Currency with Money Pattern — Progress

## Status: Done

## Plan
- [x] Create `internal/currency/currency.go` — Currency type, Registry, Get, Supported, driver.Valuer/sql.Scanner
- [x] Create `internal/currency/currency_test.go` — unit tests
- [x] Create `internal/money/money.go` — Money type, New, Add, Subtract, Format, IsZero, MarshalJSON, UnmarshalJSON
- [x] Create `internal/money/money_test.go` — unit tests
- [x] Update `internal/models/user.go` — remove AllowedCurrencies
- [x] Update `internal/handlers/auth.go` — use currency.Supported()
- [x] Update `views/register.html` — show "Name (Code)" in picker
- [x] Run unit tests — go test ./...
- [x] Run code simplifier
- [x] Create `e2e/currency.spec.ts` — Playwright E2E tests
- [x] Run Playwright tests

## Implementation Log

### internal/currency package
- Files: `internal/currency/currency.go`, `internal/currency/currency_test.go`
- Currency struct with Code/Name/Symbol/Exponent; Registry map; Get/Supported functions
- driver.Valuer + sql.Scanner on Currency for GORM single-column persistence

### internal/money package
- Files: `internal/money/money.go`, `internal/money/money_test.go`
- Money{Amount int64, Currency currency.Currency `gorm:"column:currency_code"`}
- Add/Subtract return error on currency mismatch
- Format() uses golang.org/x/text/currency + strips locale-inserted spaces → "€10.99"
- MarshalJSON/UnmarshalJSON: decimal string wire format {"amount":"10.99","currency":"EUR"}
- parseDecimal helper: integer-only conversion avoids float precision issues

### Registration wiring
- Files: `internal/handlers/auth.go`, `internal/models/user.go`, `views/register.html`
- Removed models.AllowedCurrencies (replaced by currency.Supported())
- Handler passes []currency.Currency to template; validation uses currency.Get()
- Code-simplifier extracted renderRegister closure in RegisterSubmit
- Template renders "Euro (EUR)" with value="EUR"

### golang.org/x/text promotion
- Was indirect; importing subpackages in money.go + go mod tidy made it direct

### E2E tests
- File: `e2e/currency.spec.ts`
- 2 API tests: EUR accepted, XYZ rejected — both pass
- 2 browser tests: picker option count + required attr — fail on Linux arm64 (no Chromium, pre-existing env limitation shared by all page-based tests)

## Test Results

### go test ./...
```
ok  github.com/mlhmz/finances/internal/auth
ok  github.com/mlhmz/finances/internal/currency
ok  github.com/mlhmz/finances/internal/models
ok  github.com/mlhmz/finances/internal/money
```

### npx playwright test e2e/currency.spec.ts
- 2 passed (API tests: EUR accepted, XYZ rejected)
- 2 failed (browser/page tests: Chromium not available on Linux arm64 — same pre-existing issue as auth.spec.ts browser tests)

## Blockers

None. Browser-based Playwright tests require Chromium on Linux arm64 — install with `npx playwright install chromium` when running on a supported platform.
