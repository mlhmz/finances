# Feature 1 — Authentication: Brainstorm

## Status: All questions resolved ✓ — Ready to write spec

---

## Decisions Log

| Topic | Decision |
|---|---|
| OTP format | 8-char alphanumeric (e.g. `X4K9P2WR`) |
| OTP expiry | 15 minutes |
| OTP max attempts | 3 wrong attempts → OTP invalidated, must request new one |
| JWT storage | HttpOnly cookies (both access + refresh) |
| Access token TTL | 1 hour |
| Refresh token TTL | 7 days |
| New user flow | Full redirect to `/register` page |
| Route protection | All routes except `/login` and `/register` require auth |
| Password reset | Not needed — passwordless; just request a new OTP |

---

## Flows

### Flow A — Known User Login

```
GET /login
    │
    User enters email → POST /auth/request
    │
    Backend: email found in DB
    │→ generate 8-char OTP (alphanumeric, uppercase)
    │→ store OTP in DB with expiry = now + 15min, attempt_count = 0
    │→ print OTP to server console: "[AUTH] OTP for user@example.com: X4K9P2WR"
    │→ return HTMX fragment: OTP code-entry form (replaces submit area)
    │
    User sees code in console (dev mode), types it into the form
    │
    POST /auth/verify  { email, code }
    │
    Backend:
    │→ look up OTP record by email
    │→ check: not expired, attempt_count < 3, code matches (case-insensitive)
    │        ├─ fail → increment attempt_count, return error fragment
    │        │         if attempt_count >= 3 → invalidate OTP, show "request new code"
    │        └─ success → delete OTP record
    │                   → issue access JWT (1h) as HttpOnly cookie
    │                   → issue refresh JWT (7d) as HttpOnly cookie
    │                   → redirect 302 → /  (app home)
```

### Flow B — New User Registration

```
GET /login
    │
    User enters unknown email → POST /auth/request
    │
    Backend: email NOT found in DB
    │→ redirect 302 → /register?email=user@example.com
    │
    GET /register  (email pre-filled, read-only)
    │
    User fills:
      Full Name (required, text)
      Currency   (select, only "EUR" for now)
    │
    POST /register  { email, full_name, currency }
    │
    Backend:
    │→ validate: email not already registered, name not blank
    │→ create User record (id, email, full_name, currency, initials, created_at)
    │→ generate 8-char OTP → store → print to console
    │→ render /register page with OTP code-entry form visible
    │
    User enters code → POST /auth/verify  (same as Flow A from here)
```

### Flow C — Silent Token Refresh

```
Any authenticated request
    │
    Auth middleware: validate access JWT from cookie
    │
    ├─ valid → attach user to context, continue
    │
    └─ expired → check refresh JWT cookie
                    ├─ valid refresh → issue new access JWT cookie → continue
                    └─ invalid/expired → clear both cookies → redirect 302 → /login
```

### Flow D — Logout

```
POST /auth/logout
    │
    │→ clear access token cookie (MaxAge = -1)
    │→ clear refresh token cookie (MaxAge = -1)
    │→ redirect 302 → /login
```

---

## Data Model

### `users` table

| Column | Type | Notes |
|---|---|---|
| `id` | UUID (TEXT) | primary key |
| `email` | TEXT | unique, not null |
| `full_name` | TEXT | not null |
| `currency` | TEXT | e.g. "EUR" |
| `initials` | TEXT | derived from full_name on create |
| `created_at` | DATETIME | auto |
| `updated_at` | DATETIME | auto |

### `otp_tokens` table

| Column | Type | Notes |
|---|---|---|
| `id` | UUID (TEXT) | primary key |
| `user_id` | TEXT | FK → users.id |
| `code` | TEXT | 8-char alphanumeric, stored hashed (bcrypt/sha256) |
| `expires_at` | DATETIME | now + 15min |
| `attempt_count` | INTEGER | default 0 |
| `created_at` | DATETIME | auto |

> One active OTP per user at a time. Creating a new OTP replaces/deletes any existing one for that user.

---

## JWT Structure

### Access Token Claims

```json
{
  "sub": "<user_id>",
  "email": "user@example.com",
  "iat": 1234567890,
  "exp": 1234567890
}
```

### Refresh Token Claims

```json
{
  "sub": "<user_id>",
  "type": "refresh",
  "iat": 1234567890,
  "exp": 1234567890
}
```

Signing algorithm: **HS256** with a secret from app config (`JWT_SECRET` env var).

---

## Routes

| Method | Path | Auth required | Description |
|---|---|---|---|
| GET | `/login` | No | Login page |
| POST | `/auth/request` | No | Request OTP; HTMX partial or redirect |
| POST | `/auth/verify` | No | Verify OTP, issue tokens |
| GET | `/register` | No | Registration page (new users) |
| POST | `/register` | No | Create user + trigger OTP |
| POST | `/auth/logout` | Yes | Clear tokens, redirect to /login |
| GET | `/auth/refresh` | No (uses refresh cookie) | Issue new access token |

---

## Views

| File | Description |
|---|---|
| `views/login.html` | Email input form |
| `views/register.html` | Full Name + Currency form (email pre-filled) |
| `views/partials/otp_form.html` | HTMX fragment: code-entry box shown after OTP is sent |
| `views/partials/otp_error.html` | HTMX fragment: inline error (wrong code / expired) |

---

## Avatar (Initials)

- Derived from `full_name` on user creation: first letter of first word + first letter of last word, uppercase.
  - `"Malek Mustafa"` → `"MM"`
  - `"Ada"` → `"A"`
- Displayed in the app navbar after login.
- No external image service needed.

---

## Console OTP Format

```
[AUTH] OTP for user@example.com: X4K9P2WR (expires in 15 minutes)
```

---

## Open Questions

- [x] Should the OTP code be stored as a hash (bcrypt/sha256) or plaintext in the DB? → **SHA-256 hash**
- [x] What should the JWT_SECRET default be in development? → **Hardcoded dev default, overridden by `JWT_SECRET` env var**
- [x] Should `/register` accept a GET with no email param? → **No — must always arrive via redirect with `?email=` param; redirect to `/login` if param is missing**

---

## ASCII Overview

```
  ┌─────────────┐     ┌──────────────┐     ┌──────────────┐
  │  /login     │────►│ /auth/request│────►│ /auth/verify │
  │  (email)    │     │  (known?)    │     │ (code check) │
  └─────────────┘     └──────┬───────┘     └──────┬───────┘
                             │ unknown             │ success
                             ▼                     ▼
                      ┌──────────────┐     ┌──────────────┐
                      │  /register   │     │  JWT cookies │
                      │ (name+curr)  │     │  issued      │
                      └──────────────┘     └──────┬───────┘
                                                  │
                                                  ▼
                                           ┌──────────────┐
                                           │   App (/)    │
                                           │  protected   │
                                           └──────────────┘
```
