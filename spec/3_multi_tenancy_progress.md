# Feature 3: Multi-Tenancy — Progress

## Status: Done

## Plan

- [x] Add `middleware.CurrentUserID(c)` helper
- [x] Create `internal/repository/user.go` (UserRepository)
- [x] Create `internal/handlers/profile.go` (ProfilePage, ProfileUpdate)
- [x] Update `internal/handlers/home.go` (pass user data to Index)
- [x] Refactor route registration to Fiber group in `cmd/finances/main.go`
- [x] Create `views/profile.html`
- [x] Update `views/index.html` (show user name/initials in header)
- [x] Unit tests for UserRepository
- [x] Playwright E2E tests (`e2e/profile.spec.ts`)
- [x] Manual verification

## Implementation Log

### Add `middleware.CurrentUserID(c)` helper
- Files changed: `internal/middleware/auth.go`
- Notes: Added above `AuthMiddleware`; safe type-assertion returning `""` outside auth context.

### Create `internal/repository/user.go`
- Files changed: `internal/repository/user.go` (new)
- Notes: `UserRepository` struct with constructor-bound `userID`. `Get()` and `Update()` both scope to that ID — never accepts userID from caller at query time.

### Create `internal/handlers/profile.go`
- Files changed: `internal/handlers/profile.go` (new)
- Notes: `ProfilePage` (GET) and `ProfileUpdate` (POST). Simplified by code-simplifier: render closure captures `user` from scope; uses `maps.Copy` for data merging.

### Update `internal/handlers/home.go`
- Files changed: `internal/handlers/home.go`
- Notes: `Index` now fetches the authenticated user via `UserRepository` and passes `UserName`, `UserInitials`, `UserEmail` to the template.

### Refactor route registration to Fiber group
- Files changed: `cmd/finances/main.go`
- Notes: Single `app.Group("", authMw)` replaces per-handler middleware. Added `/profile` GET and POST routes.

### Views
- Files changed: `views/index.html`, `views/profile.html` (new)
- Notes: Full-width sticky nav with brand + avatar link + logout button. Dashboard shows welcome greeting. Profile page has large avatar, editable form, success/error feedback.

### Unit tests
- Files changed: `internal/repository/user_test.go` (new)
- Notes: 5 tests covering Get, not-found, isolation between users, Update, and update isolation. All pass with `CGO_ENABLED=1`.

### Playwright E2E tests
- Files changed: `e2e/profile.spec.ts` (new)
- Notes: 9 tests. Auth guard, rendering, update success, 3 validation cases, data isolation. All use `request` or `playwright.request.newContext()` (no browser binary required).

## Test Results

### Unit tests — `go test ./...`
```
ok  github.com/mlhmz/finances/internal/auth
ok  github.com/mlhmz/finances/internal/currency
ok  github.com/mlhmz/finances/internal/models
ok  github.com/mlhmz/finances/internal/money
ok  github.com/mlhmz/finances/internal/repository   (5 tests)
```
All pass.

### E2E tests — `npx playwright test e2e/profile.spec.ts`
```
9 passed (3.1s)
```
All pass.

### Full suite — `npx playwright test`
```
20 passed, 12 failed
```
The 12 failures are pre-existing ARM64 browser binary issue (all use `page`/`browser` fixtures requiring Chromium). Zero regressions introduced by this feature.

## Blockers

None.
