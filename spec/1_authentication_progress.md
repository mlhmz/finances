# Feature 1: Authentication — Progress

## Status: Done

## Plan

- [x] Add JWT dependency (`github.com/golang-jwt/jwt/v5`)
- [x] Update `internal/config/config.go` — add `JWTSecret`, `JWTAccessTTL`, `JWTRefreshTTL`
- [x] Create `internal/models/user.go` — `User` struct + initials helper
- [x] Create `internal/models/otp_token.go` — `OTPToken` struct
- [x] Update `internal/db/db.go` — auto-migrate models, expose global `*gorm.DB`
- [x] Create `internal/auth/jwt.go` — JWT issuance/parsing + OTP generation/hashing
- [x] Create `internal/handlers/auth.go` — `RequestOTP`, `VerifyOTP`, `Register` (GET+POST), `Logout`, `Refresh`, `TestLastOTP`
- [x] Create `internal/middleware/auth.go` — JWT validation middleware (access + silent refresh)
- [x] Update `cmd/finances/main.go` — register all auth routes, protect `/` per-handler
- [x] Create `views/login.html`
- [x] Create `views/register.html`
- [x] Create `views/partials/otp_form.html`
- [x] Create `views/partials/otp_error.html`
- [x] Write unit tests (`internal/auth/jwt_test.go`, `internal/models/user_test.go`)
- [x] Write Playwright E2E tests (`e2e/auth.spec.ts`)
- [x] All tests pass

## Implementation Log

### Dependencies
- Files changed: `go.mod`, `go.sum`
- Added: `github.com/golang-jwt/jwt/v5 v5.3.1`

### Config
- Files changed: `internal/config/config.go`
- Added `JWTSecret` (env `JWT_SECRET`, fallback dev value), `JWTAccessTTL` (3600s), `JWTRefreshTTL` (604800s)

### Models
- Files changed: `internal/models/user.go`, `internal/models/otp_token.go`, `internal/models/doc.go`
- User: UUID PK, email unique, full name, currency, initials, timestamps
- OTPToken: UUID PK, user FK, SHA-256 code hash, expiry, attempt count
- `DeriveInitials` helper: 1 word → 1 char, 2+ words → first+last initial

### DB
- Files changed: `internal/db/db.go`
- Auto-migrates `User` and `OTPToken`; exposes package-level `DB` var

### Auth utilities
- Files changed: `internal/auth/jwt.go`
- HS256 JWT issuance/parsing for access (1h) and refresh (7d) tokens
- `GenerateOTP` (8-char crypto-random uppercase alphanumeric) + `HashOTP` (SHA-256 hex)

### Handlers
- Files changed: `internal/handlers/auth.go`
- Full auth flow: RequestOTP, VerifyOTP, RegisterPage, RegisterSubmit, Logout, Refresh
- Per-email OTP map (`lastOTPByEmail` + mutex) for test backdoor isolation
- `TestLastOTP` endpoint gated by `TEST_MODE=1`

### Middleware
- Files changed: `internal/middleware/auth.go`
- Access token → validate → continue
- Expired access + valid refresh → silent reissue → continue
- No valid session → clear cookies → 302 /login

### Routes
- Files changed: `cmd/finances/main.go`
- Auth middleware applied per-handler (not via `Group("/", ...)`) to avoid intercepting all routes
- Test backdoor registered only when `TEST_MODE=1`

### Views
- Files changed: `views/login.html`, `views/register.html`, `views/partials/otp_form.html`, `views/partials/otp_error.html`

### Playwright config
- Files changed: `playwright.config.ts`
- webServer command updated to `TEST_MODE=1 go run ./cmd/finances`

## Test Results

### Unit tests
```
ok  github.com/mlhmz/finances/internal/auth   (9 tests)
ok  github.com/mlhmz/finances/internal/models (6 tests)
```

### E2E tests (Playwright — Chromium)
```
16 passed (4.6s)
```

All 16 auth E2E tests pass covering:
- Auth guards (unauthenticated redirects)
- Login page rendering
- Register page rendering + validation
- Full register → OTP → session flow (API)
- Known-user login → OTP → cookies + HX-Redirect
- Logout clears session
- OTP lockout after 3 wrong attempts

## Blockers

None.
