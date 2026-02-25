# Feature 4: Manual Transaction Tracking

## Overview

Introduces the core financial data entry capability: authenticated users manually record income and
expense transactions. Each transaction captures a signed amount (positive = income, negative =
expense), a currency, a datetime, a title, and an optional description. The feature delivers full
CRUD with an HTMX-powered list page, a modal (desktop) / drawer (mobile) create-and-edit form, and
inline delete confirmation — all scoped to the authenticated user via the existing repository
pattern.

---

## Goals

- Users can create, read, update, and delete their own transactions.
- Transactions are stored with signed integer minor units using the existing Money pattern.
- Income vs. expense is conveyed via a toggle in the UI; amounts are always entered as positive
  numbers.
- The transaction list is paginated (20/page), sorted newest first.
- Navigation is responsive: bottom tab bar on mobile (≤480px), collapsible sidebar on desktop.
- All data is strictly scoped to the authenticated user through the repository layer.

## Non-Goals

- Account spaces (deferred to Feature 5).
- Categories (deferred to Feature 7).
- Bulk import or CSV upload.
- Transaction search or filtering.
- Transfer transactions (money moving between accounts).

---

## Data Model

### `transactions` table

```go
// internal/models/transaction.go
type Transaction struct {
    ID          string      `gorm:"primaryKey"`
    UserID      string      `gorm:"not null;index"`
    Title       string      `gorm:"not null"`
    Description string
    Amount      money.Money `gorm:"embedded;embeddedPrefix:amount_"`
    Date        time.Time   `gorm:"not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

| Column               | Type     | Constraints              | Notes                                      |
|----------------------|----------|--------------------------|--------------------------------------------|
| id                   | TEXT     | PRIMARY KEY              | UUID v4                                    |
| user_id              | TEXT     | NOT NULL, INDEX          | FK → users.id                              |
| title                | TEXT     | NOT NULL                 |                                            |
| description          | TEXT     |                          | Nullable; empty string treated as absent   |
| amount_amount        | INTEGER  | NOT NULL                 | Signed minor units; positive=income, negative=expense |
| amount_currency_code | TEXT     | NOT NULL                 | ISO 4217 code (e.g. `"EUR"`)               |
| date                 | DATETIME | NOT NULL                 | Full datetime; time component is optional  |
| created_at           | DATETIME |                          |                                            |
| updated_at           | DATETIME |                          |                                            |

`db.AutoMigrate` adds this table alongside the existing `users` and `otp_tokens` tables.

### Convention

Follows the established multi-tenancy convention: `UserID string \`gorm:"not null;index"\`` is
present, and all queries are scoped by it in the repository.

---

## Repository

**New file:** `internal/repository/transaction.go`

```go
type TransactionRepository struct {
    db     *gorm.DB
    userID string
}

func NewTransactionRepository(userID string) *TransactionRepository

// List returns a page of transactions for the user, sorted newest first.
// Returns the slice, total record count, and any error.
func (r *TransactionRepository) List(page, pageSize int) ([]models.Transaction, int64, error)

// Create inserts a new transaction. Sets ID (UUID) before insert.
func (r *TransactionRepository) Create(t *models.Transaction) error

// GetByID returns the transaction with the given ID scoped to the user.
func (r *TransactionRepository) GetByID(id string) (*models.Transaction, error)

// Update saves changes to an existing transaction (scoped to user).
func (r *TransactionRepository) Update(t *models.Transaction) error

// Delete removes the transaction with the given ID (scoped to user).
func (r *TransactionRepository) Delete(id string) error
```

All methods filter by `r.userID` — never accepting a user ID from request input.

---

## API / Routes

Register under the existing `protected` group in `cmd/finances/main.go`.

| Method | Path                              | Handler                   | Description                                         |
|--------|-----------------------------------|---------------------------|-----------------------------------------------------|
| GET    | /transactions                     | `TransactionsPage`        | Full page render with paginated list                |
| POST   | /transactions                     | `CreateTransaction`       | Create; returns updated list partial (HTMX)         |
| GET    | /transactions/:id/edit            | `EditTransactionForm`     | Returns form partial pre-filled (HTMX, modal/drawer)|
| PUT    | /transactions/:id                 | `UpdateTransaction`       | Update; returns updated row partial (HTMX)          |
| GET    | /transactions/:id/confirm-delete  | `ConfirmDeleteTransaction`| Returns inline confirmation partial (HTMX)          |
| DELETE | /transactions/:id                 | `DeleteTransaction`       | Delete; returns empty response + removes row (HTMX) |

### GET /transactions

Query params:

| Param | Type | Default | Description         |
|-------|------|---------|---------------------|
| page  | int  | 1       | 1-based page number |

Response: renders `views/transactions.html` with:
```go
fiber.Map{
    "Transactions": []models.Transaction,
    "Page":         int,
    "TotalPages":   int,
    "UserCurrency": string,  // user's profile currency code, for form default
}
```

### POST /transactions

Form fields:

| Field       | Required | Validation                                     |
|-------------|----------|------------------------------------------------|
| type        | yes      | `"income"` or `"expense"`                      |
| amount      | yes      | decimal string, > 0                            |
| currency    | yes      | must be in `currency.Registry`                 |
| title       | yes      | non-empty                                      |
| date        | yes      | datetime string; time defaults to 00:00 if absent |
| description | no       | any string                                     |

Amount is stored as: `+minorUnits` if type=income, `-minorUnits` if type=expense.

- **Success (HX-Request):** return HTMX partial that prepends the new row to the list and closes
  the modal/drawer.
- **Error:** return the form partial with validation errors.

### GET /transactions/:id/edit

Returns the modal/drawer form partial pre-filled with the transaction's current values.
404 if transaction does not belong to the current user.

### PUT /transactions/:id

Same form fields as POST. Updates the transaction.

- **Success:** return the updated row partial (replaces the row in the list).
- **Error:** return the form partial with validation errors.

### GET /transactions/:id/confirm-delete

Returns an inline confirmation partial for the row.

### DELETE /transactions/:id

Deletes the transaction. Returns HTTP 200 with an empty body (HTMX removes the row via
`hx-swap="outerHTML"` targeting the row element).

---

## UI / UX

### Navigation

**Desktop (>480px) — collapsible sidebar on the left:**

```
┌────────────────────────────────────────────────────────────────────────┐
│ ┌────────────┐  ┌────────────────────────────────────────────────────┐ │
│ │            │  │ Transactions                  [+ New Transaction]  │ │
│ │ [⌂] Home   │  │────────────────────────────────────────────────────│ │
│ │ [≡] Trans  │  │ 25 Feb 2026  Grocery shopping        − €42.50  [⋮] │ │
│ │ [◯] Profile│  │              REWE, Berlin                          │ │
│ │            │  │────────────────────────────────────────────────────│ │
│ └────────────┘  │ 24 Feb 2026  Freelance invoice       + €800.00 [⋮] │ │
│                 │              Client: Acme Corp                      │ │
│                 │────────────────────────────────────────────────────│ │
│                 │                  ‹  Page 2 of 7  ›                  │ │
│                 └────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────┘
```

**Mobile (≤480px) — bottom tab bar:**

```
┌─────────────────────────────────┐
│ Transactions          [+ New]   │
│─────────────────────────────────│
│ 25 Feb 2026                     │
│ Grocery shopping    − €42.50    │
│ REWE, Berlin                    │
│─────────────────────────────────│
│ 24 Feb 2026                     │
│ Freelance invoice   + €800.00   │
│ Client: Acme Corp               │
│─────────────────────────────────│
│          ‹ Page 2 of 7 ›        │
├─────────────────────────────────┤
│  [⌂ Home]  [≡ Trans]  [◯ Prof] │
└─────────────────────────────────┘
```

### Transaction List Row

Each row occupies two visual lines:

```
┌──────────────────────────────────────────────────────────┐
│  25 Feb 2026   Grocery shopping          − €42.50   [⋮]  │
│                REWE, Berlin                               │
└──────────────────────────────────────────────────────────┘
```

- Date formatted as `DD Mon YYYY`.
- Amount prefix: `+` for income (green), `−` for expense (red).
- Description preview: single line, truncated with ellipsis if long; hidden if empty.
- `[⋮]` action menu: **Edit** and **Delete** options.

### Create / Edit — Modal (desktop >480px)

```
┌───────────────────────────────────────────┐
│  New Transaction                     [✕]  │
│───────────────────────────────────────────│
│                                           │
│    [ Income ]     [ Expense ]             │
│                                           │
│  Amount   [              ] [ EUR  ▼ ]    │
│  Title    [                           ]   │
│  Date     [ 2026-02-25  14:30         ]   │
│  Desc.    [                           ]   │
│           [                           ]   │
│                                           │
│                  [Cancel]   [Save]        │
└───────────────────────────────────────────┘
```

### Create / Edit — Drawer (mobile ≤480px, slides up from bottom)

```
┌─────────────────────────────────┐
│            ▬▬▬▬▬                │  ← drag handle
│  New Transaction                │
│─────────────────────────────────│
│  [ Income ]   [ Expense ]       │
│                                 │
│  Amount  [          ] [ EUR ▼ ] │
│  Title   [                   ]  │
│  Date    [ 2026-02-25  14:30 ]  │
│  Desc.   [                   ]  │
│          [                   ]  │
│                                 │
│      [Cancel]        [Save]     │
└─────────────────────────────────┘
```

### Inline Delete Confirmation

The row is replaced by a confirmation strip:

```
│  Delete "Grocery shopping"?          [Cancel]  [Delete]  │
```

`[Cancel]` restores the original row. `[Delete]` fires `DELETE /transactions/:id`.

### Pagination Controls

```
                    ‹  Page 2 of 7  ›
```

`‹` (previous) is disabled on page 1. `›` (next) is disabled on the last page.

---

## Acceptance Criteria

1. A logged-in user can reach `/transactions` via the Transactions item in the collapsible sidebar
   (desktop, >480px) or via the Transactions tab in the bottom tab bar (mobile, ≤480px).
2. The page lists the authenticated user's transactions only — no cross-user data is accessible.
3. Transactions are sorted newest first and paginated at 20 per page.
4. Each row displays: formatted date, title, signed and formatted amount (`+ €X` or `− €X`),
   description preview (truncated if long; absent if empty), and an action menu (⋮).
5. Clicking `[+ New Transaction]` opens a modal (desktop) or a bottom drawer (mobile).
6. The create form contains: Income/Expense toggle, positive amount field, currency picker
   (defaults to user's profile currency), required title field, required datetime field (prefilled
   with current date and time), and an optional description textarea.
7. Submitting a valid create form inserts the transaction and prepends the new row to the list
   without a full page reload; the modal/drawer closes.
8. Submitting an invalid create form (missing required fields, zero or negative amount, unknown
   currency) returns validation errors inside the form; no transaction is created.
9. The action menu (⋮) offers "Edit" and "Delete".
10. Clicking "Edit" opens the modal/drawer pre-filled with the transaction's current values.
11. Submitting a valid edit form updates the transaction and replaces the list row with updated
    data without a full page reload.
12. Clicking "Delete" replaces the row with an inline confirmation:
    `Delete "{title}"? [Cancel] [Delete]`.
13. Confirming deletion removes the transaction from the database and removes the row from the DOM
    without a full page reload.
14. Pagination controls display "‹ Page N of M ›"; `‹` is disabled on page 1, `›` is disabled on
    the last page.
15. Amounts are stored as signed integers in minor units: income is positive, expense is negative.
16. The transaction datetime stores both date and time; when only a date is supplied, time defaults
    to midnight (00:00).
17. `TransactionRepository` always scopes queries by the constructor-injected `userID`; user ID is
    never accepted from request input.

---

## Open Questions

None — all design decisions resolved.
