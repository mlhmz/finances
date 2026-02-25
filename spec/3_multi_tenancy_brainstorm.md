# Feature 3 — Multi-Tenancy: Brainstorm

## What the Feature Is

Feature 3 establishes **complete data isolation between users**: every DB query for user-owned data
must be scoped by `user_id` at the repository/handler layer so that no user can ever read or modify
another user's records.

Because Features 4–10 (transactions, account spaces, fix costs, categories, etc.) haven't been built
yet, Feature 3 is primarily **foundational infrastructure**: it sets up the pattern, helpers, and
conventions that all future handlers will follow automatically.

---

## Current State

| Area | Status |
|------|--------|
| Auth middleware | ✅ stores `c.Locals("userID")` + `c.Locals("email")` |
| Protected routes | ✅ use per-handler `authMw` middleware |
| DB queries | ✅ `db.DB.Where("user_id = ?", ...)` used in OTPToken — pattern exists |
| Repository/service layer | ❌ none — handlers call `db.DB` directly |
| Route groups | ❌ no group; auth applied per-route individually |
| UserID helper | ❌ no helper — type-assert from `c.Locals` directly |
| Home page renders user | ❌ `Index` ignores auth context entirely |

---

## Iteration 1 — Core Architectural Questions

### Q1: Scoping pattern for DB queries

Three options for how future handlers scope queries by user ID:

```
Option A — Inline GORM Where
    db.DB.Where("user_id = ?", userID).Find(&transactions)

Option B — GORM Scope functions (chainable helpers)
    func ByUser(userID string) func(*gorm.DB) *gorm.DB {
        return func(db *gorm.DB) *gorm.DB {
            return db.Where("user_id = ?", userID)
        }
    }
    db.DB.Scopes(ByUser(userID)).Find(&transactions)

Option C — Repository layer (internal/repository/)
    type TransactionRepo struct { db *gorm.DB; userID string }
    func (r *TransactionRepo) List() ([]models.Transaction, error) { ... }
```

**Question: Which scoping pattern should be used?**
- [ ] A — Inline `.Where("user_id = ?", userID)` in each handler (simple, no new layers)
- [ ] B — GORM Scope helper functions in a shared package (reusable, still no extra layer)
- [x] C — Repository layer in `internal/repository/` (more structure, more boilerplate)

---

### Q2: Route group organization

Currently each protected route applies `authMw` individually:
```go
app.Get("/", authMw, handlers.Index)
app.Get("/greet", authMw, handlers.Greet)
app.Post("/auth/logout", authMw, handlers.Logout)
```

Alternatively, use a Fiber route group so the middleware is guaranteed for all protected routes:
```go
protected := app.Group("", authMw)
protected.Get("/", handlers.Index)
protected.Get("/greet", handlers.Greet)
protected.Post("/auth/logout", handlers.Logout)
```

**Question: Should protected routes be moved into a Fiber route group?**
- [x] Yes — route group guarantees no route accidentally omits the middleware
- [ ] No — keep per-handler middleware (current style, explicit and visible)

---

### Q3: CurrentUser helper

Handlers need to extract the authenticated user's ID (and optionally email) from the context.
Currently the pattern would be a raw type assertion:
```go
userID := c.Locals("userID").(string)
```

A small helper prevents panics on missing values:
```go
// in internal/middleware or internal/handlers
func CurrentUserID(c *fiber.Ctx) string {
    id, _ := c.Locals("userID").(string)
    return id
}
```

**Question: Where should the CurrentUserID helper live?**
- [x] `internal/middleware/` — alongside the auth middleware that sets the value
- [ ] `internal/handlers/` — as a package-level helper used only by handlers (unexported ok)
- [ ] No helper needed — inline type assertion is fine given the codebase is small

---

## Iteration 2 — Visible Deliverables

Feature 3's spec says "data isolation" and "scoped by user ID at the repository layer". This is
architectural, but does it ship any visible UI change?

Possible concrete deliverables:
1. **Home page shows the user's name/initials** — the `Index` handler already receives the auth
   context; it just doesn't pass user data to the template yet.
2. **Profile / Settings page** — a new GET /profile page displaying the user's name, email,
   currency preference, with an option to update display name.
3. **Pure infrastructure only** — no new UI; just the patterns, helpers, and route groups.

### Q4: Should Feature 3 include any visible UI?

**Question: What should be the visible output of Feature 3?**
- [ ] Home page only — pass authenticated user data (name, initials, email) to the index template
- [x] Home page + Profile page (GET /profile) — displays user info, optionally allows editing name/currency
- [ ] No UI — purely architectural (patterns, helpers, route grouping only)

---

## Iteration 3 — Profile Page (if included)

If a profile page is part of this feature:

```
┌─────────────────────────────────────────────────────┐
│  ← Back                                  [Logout]   │
├─────────────────────────────────────────────────────┤
│                                                     │
│         ┌──────┐                                    │
│         │  ML  │  Malek Lastname                    │
│         └──────┘  malek@example.com                 │
│                   Currency: EUR                     │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │  Full Name  [__________________________]     │   │
│  │  Currency   [EUR ▼]                          │   │
│  │                        [Save Changes]        │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

### Q5: If profile page is included — is it read-only or editable?

**Question: Should the profile page allow editing?**
- [ ] Read-only — display name, email, initials, currency (no form)
- [x] Editable — allow changing Full Name and Currency (email is identity, stays fixed)

---

## Iteration 4 — Edge Cases & Conventions

### Scoping convention for future models

All future user-owned models (transactions, account spaces, etc.) must include:
```go
UserID string `gorm:"not null;index"`
```
This should be documented as a project convention.

### GORM AutoMigrate

Feature 3 adds no new tables. The `Connect()` function in `internal/db/db.go` will need to be
updated when future models are added. No change needed for Feature 3 itself (unless a profile
update handler writes to `users` table — which is already migrated).

### Security invariant

**Every handler that reads or writes user-owned data must:**
1. Be behind the auth middleware (guaranteed by route group if Q2 = Yes)
2. Extract `userID` from `c.Locals` (never from query params or request body)
3. Pass `userID` as a WHERE clause to every relevant query

This is the core contract of multi-tenancy in this app.

---

## Summary of Open Questions

1. **Scoping pattern** — inline Where / GORM Scopes / Repository layer?
2. **Route grouping** — Fiber Group or keep per-handler middleware?
3. **CurrentUserID helper** — middleware pkg / handlers pkg / no helper?
4. **Visible UI** — home page data / home + profile / no UI?
5. **Profile editability** — read-only or editable? (only relevant if Q4 includes profile)
