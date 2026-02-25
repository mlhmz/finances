# Feature 2: Multi-Currency with Money Pattern — Brainstorm

## Context

The roadmap says:
- All amounts stored as **integer minor units** (e.g. cents for EUR).
- **Euro (EUR)** is the first implemented currency.
- The currency model must accommodate additional currencies in future iterations.

This feature is foundational — every subsequent feature that touches money (transactions, fix costs, dashboard) will build on the types and patterns established here.

---

## Iteration 1 — Core Concepts

### What we know
- Minor-unit integer storage avoids floating-point precision bugs.
- EUR has 2 decimal places (100 cents = 1 euro).
- Other currencies have different exponents: JPY=0, KWD=3, etc.
- A `Money` value always carries both an amount **and** a currency.

### Open questions surfaced
1. Where is the currency list authoritative — code, DB, or both?
2. How does Go represent a `Money` value?
3. How are two columns stored per money field in the DB?
4. What arithmetic operations does Money support?
5. How is Money serialized for the HTTP API and for HTML display?
6. How does the currency picker in Feature 1 registration interact with this feature?

---

## Iteration 2 — Currency Registry

Two schools of thought:

### Option A — Currency hard-coded in Go

```
internal/currency/
├── currency.go    // Currency type, registry map, Exponent(), Format()
└── currency_test.go
```

A `var Registry = map[string]Currency{...}` holds all supported currencies.
The DB `currencies` table is just a view/reference — or not needed at all.

Pros: No DB query needed to know EUR has 2 decimals. Fast. Simple.
Cons: Adding a currency requires a code deploy.

### Option B — Currencies table in DB

```sql
CREATE TABLE currencies (
  code     TEXT PRIMARY KEY,  -- "EUR"
  name     TEXT NOT NULL,      -- "Euro"
  symbol   TEXT NOT NULL,      -- "€"
  exponent INT  NOT NULL       -- 2
);
```

The Go layer loads supported currencies from the DB at startup.

Pros: Adding a currency is a migration + data change only. More extensible.
Cons: Extra DB dependency for a simple lookup. Overkill for a personal finance app.

### Option C — Hard-coded Go + seeded DB table (hybrid)

Code defines the canonical list; DB has a `currencies` table that is seeded from code at startup (or migration). DB table acts as a foreign key target for other tables.

Pros: Referential integrity via FK, but logic stays in Go.
Cons: Slightly more moving parts.

**Question: How should currencies be managed?**
- [ ] A — Hard-coded in Go only (no DB table)
- [ ] B — DB table, loaded at startup
- [ ] C — Hard-coded in Go + seeded DB table (FK integrity)

---

## Iteration 3 — The Go Money Type

### Option A — Struct with string currency code

```go
type Money struct {
    Amount   int64
    Currency string // "EUR"
}
```

Simple, no FK coupling. Currency code validated at the boundary.

### Option B — Struct with Currency value type

```go
type Currency struct {
    Code     string
    Exponent int
    Symbol   string
}

type Money struct {
    Amount   int64
    Currency Currency
}
```

Richer. Currency carries its own formatting logic.

### Option C — Struct with separate Amount/Currency fields that GORM maps to two columns

```go
type Money struct {
    Amount       int64
    CurrencyCode string
}
// Embedded in GORM models via: `gorm:"embedded;embeddedPrefix:amount_"`
```

This maps cleanly to two DB columns (`amount_amount`, `amount_currency_code` or custom names).

**Question: What shape should the Go Money struct take?**
- [ ] A — Simple `{Amount int64, Currency string}`
- [ ] B — Rich `{Amount int64, Currency Currency}` with Currency as a value type
- [ ] C — GORM-embeddable struct with embedded prefix support

---

## Iteration 4 — DB Storage Pattern

For a transaction row that has an amount, two approaches:

### Option A — Separate named columns

```sql
amount_cents    INTEGER NOT NULL
currency_code   TEXT    NOT NULL REFERENCES currencies(code)
```

Very explicit. Easy to query with plain SQL.

### Option B — GORM embedded struct

```go
type Transaction struct {
    gorm.Model
    Amount Money `gorm:"embedded;embeddedPrefix:amount_"`
}
// Produces columns: amount_amount (int64), amount_currency_code (text)
```

Keeps Go code clean; column names less readable in raw SQL.

### Option C — JSON column

```sql
amount  TEXT NOT NULL  -- stores `{"amount":1099,"currency":"EUR"}`
```

Flexible but opaque to SQL queries and indexing.

**Question: How should Money be stored in DB columns?**
- [ ] A — Separate explicit named columns (`amount_cents`, `currency_code`)
- [ ] B — GORM embedded struct with prefix
- [ ] C — JSON column (not recommended for queries/indexing)

---

## Iteration 5 — Display Formatting

```
ASCII: Money formatting examples

  EUR, 1099 minor units  →  "€10.99"  or  "10.99 €"  or  "EUR 10.99"
  JPY, 500  minor units  →  "¥500"
  KWD, 1234 minor units  →  "KD1.234"
```

### Option A — Custom formatter in Go

```go
func (m Money) Format() string {
    // divide by 10^exponent, prefix symbol
}
```

Full control. No external dependency.

### Option B — Use `golang.org/x/text/currency`

Standard library extension. Handles locale-aware formatting.
Heavier dependency. May be overkill for a personal finance app with one currency.

**Question: How should Money be formatted for display?**
- [ ] A — Custom formatter in Go (no external dependency)
- [ ] B — `golang.org/x/text/currency` package

---

## Iteration 6 — Arithmetic Operations

What should the Money type support?

### Arithmetic rules
- Adding two `Money` values of different currencies should be an error (no implicit conversion).
- Subtraction: same rule.
- Multiplication by a scalar (e.g. for future scheduled amounts): `Money × int`.
- Division: likely not needed for now.

**Question: What arithmetic should Money support at this stage?**
- [ ] Add / Subtract (same currency only, error on mismatch)
- [ ] Add / Subtract + Multiply by integer scalar
- [ ] No arithmetic methods now — add only when needed by a consuming feature

---

## Iteration 7 — HTTP / Template Serialization

When Money appears in an API response or HTML template:

### HTTP JSON representation
- [ ] Integer-only: `{"amount": 1099, "currency": "EUR"}`
- [ ] Decimal string: `{"amount": "10.99", "currency": "EUR"}`
- [ ] Two-field object with both: `{"minor_units": 1099, "formatted": "€10.99", "currency": "EUR"}`

### HTML template representation
- [ ] Display formatted string only (`€10.99`)
- [ ] Pass the struct to the template; template calls `.Format()`

---

## Iteration 8 — Registration Currency Picker (tie-in with Feature 1)

Feature 1 says: "Currency picker (only implemented currencies are offered)."

With only EUR now:
- [ ] Show the picker but only with EUR (user sees one option — confirms the pattern is wired up)
- [ ] Default silently to EUR, skip the picker until a second currency exists
- [ ] Show the picker with EUR pre-selected but still visible

---

## Summary of All Open Questions

**Q1: Currency registry**
- [x] A — Hard-coded in Go only
- [ ] B — DB table, loaded at startup
- [ ] C — Hard-coded in Go + seeded DB table

**Q2: Go Money struct shape**
- [ ] A — Simple `{Amount int64, Currency string}`
- [x] B — Rich `{Amount int64, Currency Currency}` value type
- [ ] C — GORM-embeddable struct

**Q3: DB storage**
- [ ] A — Separate named columns (`amount_cents`, `currency_code`)
- [x] B — GORM embedded struct with prefix
- [ ] C — JSON column

**Q4: Display formatting**
- [ ] A — Custom formatter in Go
- [x] B — `golang.org/x/text/currency`

**Q5: Arithmetic operations**
- [x] Add / Subtract only
- [ ] Add / Subtract + Multiply by scalar
- [ ] No arithmetic methods yet

**Q6: HTTP JSON shape**
- [ ] Integer-only `{"amount": 1099, "currency": "EUR"}`
- [x] Decimal string `{"amount": "10.99", "currency": "EUR"}`
- [ ] Rich object with minor_units + formatted + currency

**Q7: Registration currency picker**
- [x] Show picker with EUR as sole option
- [ ] Default silently to EUR, hide picker
- [ ] Show picker with EUR pre-selected
