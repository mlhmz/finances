# Feature 2: Multi-Currency with Money Pattern

## Overview

Introduce a safe, lossless monetary value type (`Money`) used everywhere amounts appear in the application. Amounts are stored as integer minor units (e.g. cents for EUR) to avoid floating-point precision bugs. Currencies are hard-coded in Go with Euro (EUR) as the first entry. The design deliberately accommodates adding currencies in future iterations without structural changes.

## Goals

- Define a `Currency` value type (code, symbol, exponent) and a registry of supported currencies in `internal/currency`.
- Define a `Money` value type (`Amount int64` + `Currency`) with Add, Subtract, and Format methods.
- Store Money in GORM models via embedded struct with a configurable column prefix.
- Format Money for display using `golang.org/x/text/currency`.
- Serialize Money over HTTP as `{"amount": "10.99", "currency": "EUR"}` (decimal string).
- Wire the registration currency picker (Feature 1) to the Go currency registry so it lists only supported currencies.

## Non-Goals

- No DB `currencies` table — the Go registry is the single source of truth.
- No arithmetic beyond Add and Subtract (no multiplication, division, or multi-currency conversion).
- No locale detection — formatting uses a fixed locale (en-US style symbol prefix).

## Data Model

### `internal/currency/currency.go`

```go
type Currency struct {
    Code     string  // ISO 4217, e.g. "EUR"
    Name     string  // e.g. "Euro"
    Symbol   string  // e.g. "€"
    Exponent int     // decimal places, e.g. 2
}

// Registry is the single source of truth for supported currencies.
var Registry = map[string]Currency{
    "EUR": {Code: "EUR", Name: "Euro", Symbol: "€", Exponent: 2},
}

// Get returns the Currency for a code and a boolean indicating if it was found.
func Get(code string) (Currency, bool)

// Supported returns all currencies from the registry, sorted by code.
func Supported() []Currency
```

### `internal/money/money.go`

```go
type Money struct {
    Amount   int64
    Currency currency.Currency
}

// GORM-embeddable. Consumer embeds with:
//   Amount Money `gorm:"embedded;embeddedPrefix:amount_"`
// Produces columns: amount_amount (INTEGER), amount_currency_code (TEXT)

func New(amount int64, currency currency.Currency) Money
func (m Money) Add(other Money) (Money, error)      // error if currencies differ
func (m Money) Subtract(other Money) (Money, error) // error if currencies differ
func (m Money) Format() string                       // uses golang.org/x/text/currency
func (m Money) IsZero() bool

// JSON serialization: {"amount": "10.99", "currency": "EUR"}
func (m Money) MarshalJSON() ([]byte, error)
func (m *Money) UnmarshalJSON(data []byte) error
```

### GORM embedding convention

```
Model field:   Amount Money `gorm:"embedded;embeddedPrefix:amount_"`
DB columns:    amount_amount        INTEGER NOT NULL
               amount_currency_code TEXT    NOT NULL
```

Every GORM model that holds a monetary value uses this pattern. No raw `int64` or `string` columns for money elsewhere in the schema.

## API / Routes

No new HTTP routes in this feature. The Money type affects the JSON shape of existing and future routes:

```
Money JSON shape:
  { "amount": "10.99", "currency": "EUR" }

Field descriptions:
  amount   — decimal string representation (e.g. minor units ÷ 10^exponent)
  currency — ISO 4217 currency code
```

## UI / UX

### Registration currency picker (update to Feature 1)

The existing registration form gains a `<select>` populated from `currency.Supported()`.
With only EUR in the registry the picker shows one option.

```
┌─────────────────────────────────────────────┐
│  Create account                             │
│                                             │
│  Full name   [                           ]  │
│                                             │
│  Currency    [ Euro (EUR)             ▼ ]  │
│                                             │
│              [       Sign up           ]    │
└─────────────────────────────────────────────┘
```

The selected currency code is stored on the user record (future use by Multi-Tenancy and Transaction features).

## Acceptance Criteria

1. `currency.Get("EUR")` returns the Euro entry; `currency.Get("XYZ")` returns `false`.
2. `currency.Supported()` returns a slice containing exactly EUR (until more are added).
3. `money.New(1099, EUR).Format()` returns `"€10.99"`.
4. `money.New(1099, EUR).Add(money.New(50, EUR))` returns `Money{1149, EUR}` with no error.
5. `money.New(1099, EUR).Add(money.New(50, USD))` returns a non-nil error.
6. `money.New(1099, EUR).Subtract(money.New(99, EUR))` returns `Money{1000, EUR}` with no error.
7. JSON marshaling of `Money{1099, EUR}` produces `{"amount":"10.99","currency":"EUR"}`.
8. JSON unmarshaling of `{"amount":"10.99","currency":"EUR"}` produces `Money{1099, EUR}`.
9. A GORM model with `Amount Money \`gorm:"embedded;embeddedPrefix:amount_"\`` migrates to two columns (`amount_amount`, `amount_currency_code`).
10. The registration form lists every currency returned by `currency.Supported()` and no others.
11. All unit tests in `internal/currency` and `internal/money` pass under `go test ./...`.

## Package Layout

```
internal/
├── currency/
│   ├── currency.go       # Currency type, Registry, Get, Supported
│   └── currency_test.go
├── money/
│   ├── money.go          # Money type, New, Add, Subtract, Format, JSON marshal
│   └── money_test.go
```

## Dependencies

- `golang.org/x/text/currency` — add to `go.mod` via `go get golang.org/x/text`.

## Verification Steps

1. `go get golang.org/x/text` — adds the dependency.
2. `go test ./internal/currency/... ./internal/money/...` — all unit tests pass.
3. `go build ./...` — no compile errors.
4. Run the app, navigate to registration, confirm EUR appears in the currency picker.
