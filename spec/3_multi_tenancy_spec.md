# Feature 3: Multi-Tenancy

## Overview

Establishes complete data isolation between users. Every DB query for user-owned data is scoped by
`user_id` at a repository layer. Feature 3 also ships the first visible use of that pattern: the
home page renders the authenticated user's identity, and a new Profile page lets users view and
update their name and currency.

---

## Goals

- All protected routes guaranteed to have auth middleware (Fiber route group).
- `middleware.CurrentUserID(c)` helper for clean, panic-safe user ID extraction.
- `internal/repository/` package established with `UserRepository` as the reference implementation.
- Home page (`/`) displays the authenticated user's name and initials.
- Profile page (`GET /profile`) shows user info; `POST /profile` allows editing Full Name and Currency.
- Enforced convention: every future user-owned model carries `UserID string \`gorm:"not null;index"\``.

## Non-Goals

- Roles, permissions, or admin access.
- Multi-user sharing of data.
- Changing a user's email address (email is identity).
- Profile picture upload.

---

## Data Model

No new tables. The existing `users` table (already migrated) is the only table touched.

### Convention for all future user-owned models

```go
type SomeModel struct {
    ID     string `gorm:"primaryKey"`
    UserID string `gorm:"not null;index"` // required on every user-owned model
    // ...
}
```

---

## Repository Layer

**New package:** `internal/repository/`

### `UserRepository` — `internal/repository/user.go`

```go
type UserRepository struct {
    db     *gorm.DB
    userID string
}

func NewUserRepository(userID string) *UserRepository {
    return &UserRepository{db: db.DB, userID: userID}
}

func (r *UserRepository) Get() (*models.User, error)
func (r *UserRepository) Update(fullName, currency string) error
```

`userID` is set at construction time from `c.Locals`; it is never accepted from request input.
All methods implicitly scope to that user.

---

## Middleware Helper

**File:** `internal/middleware/auth.go` (extend existing file)

```go
// CurrentUserID returns the authenticated user's ID stored by AuthMiddleware.
// Returns "" if called outside an authenticated context.
func CurrentUserID(c *fiber.Ctx) string {
    id, _ := c.Locals("userID").(string)
    return id
}
```

---

## Routes

### Route group (cmd/finances/main.go)

```go
protected := app.Group("", middleware.AuthMiddleware(cfg.JWTSecret, cfg.JWTAccessTTL))
protected.Get("/", handlers.Index)
protected.Get("/profile", handlers.ProfilePage)
protected.Post("/profile", handlers.ProfileUpdate)
protected.Post("/auth/logout", handlers.Logout)
// /greet can be removed or kept; keep if desired
```

### Profile routes

| Method | Path       | Handler           | Description                          |
|--------|------------|-------------------|--------------------------------------|
| GET    | /profile   | `ProfilePage`     | Render profile with current user data |
| POST   | /profile   | `ProfileUpdate`   | Update Full Name and/or Currency     |

**GET /profile response:** renders `views/profile.html` with user data.

**POST /profile request (form):**

| Field       | Required | Validation                          |
|-------------|----------|-------------------------------------|
| full_name   | yes      | non-empty string                    |
| currency    | yes      | must exist in `currency.Supported()` |

**POST /profile response:**
- Success: re-render profile template with success message (HTMX partial or full render).
- Validation error: re-render profile template with error message and current field values.

---

## UI / UX

### Home page — header area

The existing index template should display the user's initials avatar and full name:

```
┌─────────────────────────────────────────────┐
│  Finances              [ML] Malek Lastname   │
└─────────────────────────────────────────────┘
```

Data passed from `Index` handler: `UserName`, `UserInitials`.

### Profile page — `views/profile.html`

```
┌──────────────────────────────────────────────────────┐
│  ← Home                                  [Logout]    │
├──────────────────────────────────────────────────────┤
│                                                      │
│     ┌──────┐                                         │
│     │  ML  │   Malek Lastname                        │
│     └──────┘   malek@example.com                     │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │  Full Name   [_____________________________]   │  │
│  │  Currency    [EUR                          ▼]  │  │
│  │                                                │  │
│  │                          [Save Changes]        │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ✓ Profile updated.   ← shown on success             │
│  ✗ Full name is required.  ← shown on error          │
└──────────────────────────────────────────────────────┘
```

Uses Web Awesome components consistent with existing views.

---

## Acceptance Criteria

1. All routes in `cmd/finances/main.go` that require authentication are registered under a single Fiber route group with `AuthMiddleware`.
2. `middleware.CurrentUserID(c)` returns the authenticated user's ID when called inside a protected handler, and returns `""` otherwise.
3. `internal/repository/UserRepository` exists; `Get()` returns only the record matching the constructor `userID`; `Update()` writes only to that record.
4. `GET /` passes `UserName` and `UserInitials` to the index template and the template renders them.
5. `GET /profile` renders the authenticated user's full name, email, initials, and currency.
6. `POST /profile` with valid `full_name` and `currency` updates the user record and re-renders the page with a success message.
7. `POST /profile` with an empty `full_name` or invalid `currency` re-renders the form with an error message and does not modify the DB.
8. No handler accepts `user_id` from query params or request body — user ID is always sourced from `c.Locals("userID")` via `middleware.CurrentUserID`.
9. E2E tests cover: profile page renders correct user data, successful update, validation errors.

---

## Open Questions

None — all design decisions resolved.
