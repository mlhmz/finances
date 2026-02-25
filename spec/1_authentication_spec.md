# Feature 1 — Authentication: Specification

## Overview

Passwordless authentication via a one-time code (OTP). Users log in with their email address only. On first use, the email form redirects to a registration page to collect a display name and currency. Sessions are managed with JWT access + refresh tokens stored in HttpOnly cookies.

---

## Routes

| Method | Path | Protected | Description |
|---|---|---|---|
| `GET` | `/login` | No | Render login page |
| `POST` | `/auth/request` | No | Request OTP for a known email, or redirect to `/register` for unknown emails |
| `POST` | `/auth/verify` | No | Verify OTP; issues JWT cookies on success |
| `GET` | `/register` | No | Render registration page (requires `?email=` query param) |
| `POST` | `/register` | No | Create user account, generate OTP, render code-entry |
| `GET` | `/auth/refresh` | No (reads refresh cookie) | Issue new access token from valid refresh token |
| `POST` | `/auth/logout` | Yes | Clear JWT cookies, redirect to `/login` |

All other routes require authentication. Unauthenticated requests are redirected to `/login`.

---

## Flow A — Known User Login

```
POST /auth/request  { email }
        │
        email exists in DB
        │
        ├─ delete any existing OTP for this user
        ├─ generate 8-char alphanumeric OTP (uppercase, crypto-random)
        ├─ hash OTP with SHA-256, store in otp_tokens (expires_at = now+15min, attempt_count = 0)
        ├─ print to console: [AUTH] OTP for <email>: <plaintext code> (expires in 15 minutes)
        └─ return HTMX fragment → swap #auth-message with OTP code-entry form

User types code → POST /auth/verify  { email, code }
        │
        ├─ look up otp_tokens record by user email
        ├─ check expired  → if yes: return error fragment "Code expired. Request a new one."
        ├─ check attempt_count >= 3 → return error fragment "Too many attempts. Request a new code."
        ├─ SHA-256 hash submitted code, compare to stored hash
        │        ├─ mismatch → increment attempt_count
        │        │             if attempt_count >= 3: delete OTP record
        │        │             return error fragment "Incorrect code. N attempt(s) remaining."
        │        └─ match → delete OTP record
        │                 → set cookie: access_token (JWT, HttpOnly, 1h)
        │                 → set cookie: refresh_token (JWT, HttpOnly, 7d)
        │                 → HX-Redirect: /
```

---

## Flow B — New User Registration

```
POST /auth/request  { email }
        │
        email NOT found in DB
        └─ redirect 302 → /register?email=<url-encoded email>

GET /register?email=<email>
        │
        email param missing → redirect 302 → /login
        └─ render views/register.html  (email pre-filled as read-only)

POST /register  { email, full_name, currency }
        │
        ├─ validate email not already taken (race-condition guard)
        ├─ validate full_name not blank
        ├─ validate currency is in allowed list (currently: ["EUR"])
        ├─ derive initials from full_name (see Initials Logic below)
        ├─ create users record
        ├─ generate OTP → store → print to console  (same as Flow A)
        └─ render register.html with OTP code-entry form visible

User types code → POST /auth/verify  { email, code }  (identical to Flow A)
```

---

## Flow C — Silent Token Refresh

Auth middleware runs on every protected route:

```
Read access_token cookie
        │
        ├─ valid & not expired → set user in context, continue
        │
        └─ missing or expired
                │
                Read refresh_token cookie
                │
                ├─ valid & not expired
                │       → issue new access_token cookie (1h)
                │       → set user in context, continue
                │
                └─ missing or expired
                        → clear both cookies
                        → redirect 302 → /login
```

---

## Flow D — Logout

```
POST /auth/logout
        ├─ set access_token cookie  MaxAge = -1 (delete)
        ├─ set refresh_token cookie MaxAge = -1 (delete)
        └─ redirect 302 → /login
```

---

## Data Models

### `users`

```go
type User struct {
    ID        string    `gorm:"primaryKey"`          // UUID v4
    Email     string    `gorm:"uniqueIndex;not null"`
    FullName  string    `gorm:"not null"`
    Currency  string    `gorm:"not null;default:'EUR'"`
    Initials  string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### `otp_tokens`

```go
type OTPToken struct {
    ID           string    `gorm:"primaryKey"`   // UUID v4
    UserID       string    `gorm:"not null;index"`
    CodeHash     string    `gorm:"not null"`      // SHA-256 hex of plaintext code
    ExpiresAt    time.Time `gorm:"not null"`
    AttemptCount int       `gorm:"not null;default:0"`
    CreatedAt    time.Time
}
```

Constraint: one active OTP per user. On generation, delete all existing OTP records for that user before inserting the new one.

---

## JWT

### Signing

- Algorithm: **HS256**
- Secret: read from `JWT_SECRET` environment variable; fall back to `"dev-secret-do-not-use-in-production"` if unset

### Access Token Claims

```json
{
  "sub":   "<user_id>",
  "email": "user@example.com",
  "iat":   1234567890,
  "exp":   1234567890
}
```

### Refresh Token Claims

```json
{
  "sub":  "<user_id>",
  "type": "refresh",
  "iat":  1234567890,
  "exp":  1234567890
}
```

### Cookie Settings

| Cookie | Name | HttpOnly | SameSite | Secure | MaxAge |
|---|---|---|---|---|---|
| Access token | `access_token` | Yes | Lax | false (dev) | 3600 (1h) |
| Refresh token | `refresh_token` | Yes | Lax | false (dev) | 604800 (7d) |

---

## OTP Generation

1. Generate 8 crypto-random uppercase alphanumeric characters (`[A-Z0-9]`).
2. Compute `sha256(plaintext)` → store as hex string in `otp_tokens.code_hash`.
3. Print plaintext to server console only. Never send it over the network after this point.
4. On verification: compute `sha256(submitted_code_uppercased)`, compare to stored hash.

Console format:
```
[AUTH] OTP for user@example.com: X4K9P2WR (expires in 15 minutes)
```

---

## Initials Logic

Derived from `full_name` at registration time:

- Split by whitespace.
- If 1 word: use first character of that word, uppercased.
- If 2+ words: use first character of first word + first character of last word, uppercased.

Examples:
- `"Malek Mustafa"` → `"MM"`
- `"Ada Lovelace"` → `"AL"`
- `"Ada"` → `"A"`

---

## Views

| File | Type | Description |
|---|---|---|
| `views/login.html` | Full page | Email input form; `hx-post="/auth/request"`, target `#auth-message` |
| `views/register.html` | Full page | Full Name + Currency (EUR) form; email pre-filled read-only; OTP form section hidden initially, shown after registration |
| `views/partials/otp_form.html` | HTMX fragment | Code-entry box (8-char input), submit button `POST /auth/verify`, hidden email field |
| `views/partials/otp_error.html` | HTMX fragment | Inline error text (wrong code, expired, locked) |

---

## File Structure

```
internal/
├── handlers/
│   └── auth.go          # RequestOTP, VerifyOTP, Register, Logout, Refresh
├── models/
│   ├── user.go          # User struct + initials helper
│   └── otp_token.go     # OTPToken struct
├── middleware/
│   └── auth.go          # JWT validation middleware (wraps protected routes)
└── config/
    └── config.go        # add JWTSecret, JWTAccessTTL, JWTRefreshTTL fields

views/
├── login.html
├── register.html
└── partials/
    ├── otp_form.html
    └── otp_error.html
```

---

## Dependencies to Add

| Package | Purpose |
|---|---|
| `github.com/golang-jwt/jwt/v5` | JWT signing and verification |
| `github.com/google/uuid` | UUID generation for IDs |

No third-party OTP or email library needed — OTP generation uses `crypto/rand` from the standard library. SHA-256 is from `crypto/sha256` (stdlib).

---

## Acceptance Criteria

1. A known user can log in by entering their email, receiving an OTP on the console, and entering it within 15 minutes.
2. An unknown email redirects to `/register` where the user can create an account and then complete the OTP flow.
3. After 3 wrong OTP attempts the token is invalidated and the user must request a new one.
4. An expired OTP (> 15 min) is rejected with a clear error message.
5. Access and refresh JWTs are stored exclusively in HttpOnly cookies, never exposed to JavaScript.
6. Expired access token + valid refresh token → silent re-issue of access token without redirecting.
7. Expired or missing refresh token → redirect to `/login`.
8. `POST /auth/logout` clears both cookies and redirects to `/login`.
9. Accessing any protected route without a valid session redirects to `/login`.
10. `/register` without a `?email=` param redirects to `/login`.
