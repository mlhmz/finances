# Project Structure

This document describes the directory layout for the Finances app, following the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) conventions adapted for a single-binary web application.

---

## Directory Tree

```
finances/
├── cmd/
│   └── finances/
│       └── main.go          # Entry point: wires DB, Fiber, routes, and starts the server
├── internal/
│   ├── db/
│   │   └── db.go            # GORM initialization, connection, AutoMigrate calls
│   ├── handlers/
│   │   └── *.go             # Fiber route handlers grouped by resource (e.g. transactions.go)
│   ├── models/
│   │   └── *.go             # GORM model structs (one file per domain entity)
│   └── config/
│       └── config.go        # App configuration (port, DB path, env vars)
├── views/
│   └── *.html               # Server-rendered HTML templates (Fiber HTML engine)
├── docs/
│   ├── architecture.md      # Request flow, DB strategy, frontend approach
│   └── structure.md         # This file
├── reports/
│   └── {nr}_{change}.md     # Append-only change log, one file per feature/fix
├── go.mod
├── go.sum
└── finances.db              # SQLite database file (created at runtime, gitignored)
```

---

## Directory Rationale

### `cmd/finances/`

Holds the single entry-point binary. Following Go convention, each subdirectory of `cmd/` maps to one compiled binary (`go build ./cmd/finances`). `main.go` should be thin — its only job is to read config, call initializers from `internal/`, register routes, and start the server.

### `internal/`

All application code that must not be imported by external modules lives here. The Go compiler enforces this: packages under `internal/` can only be imported by code within the same module. This is the primary home for all business logic, data access, and HTTP handling.

- **`internal/db/`** — Opens the GORM/SQLite connection and runs `AutoMigrate`. Exposes a `*gorm.DB` instance. Keeping DB setup here prevents `main.go` from growing into a grab-bag file.
- **`internal/handlers/`** — Fiber route handler functions, grouped by resource (e.g. `transactions.go`, `accounts.go`). Handlers should be thin: parse input → call a model/service → render a template or return HTML fragment. Keep business logic out of handlers.
- **`internal/models/`** — GORM model structs. Each struct maps to a DB table. No business logic here — only field definitions, GORM tags, and table-name overrides.
- **`internal/config/`** — Typed configuration struct loaded from environment variables or a config file. Centralising config avoids scattered `os.Getenv` calls throughout the codebase.

### `views/`

HTML templates consumed by Fiber's `html/v2` template engine. The engine is initialised with a root path (`./views`) and file extension (`.html`), so this directory must be co-located with the working directory when the binary runs (typically the project root during development, or embedded via `go:embed` for production builds).

### `docs/`

Long-form documentation: architectural decisions, API references, and diagrams. Not executable — purely for human readers and code review context.

### `reports/`

Append-only change log in markdown. Each feature or bugfix gets its own file named `{sequential-nr}_{short-description}.md`. This creates a lightweight audit trail without relying on commit messages alone.

---

## Key Conventions

| Rule | Reason |
|------|--------|
| `internal/` for all app code | Prevents accidental external imports; enforced by the compiler |
| Thin `main.go` | Easy to test `internal/` packages in isolation |
| One file per domain entity in `models/` | Avoids a monolithic `models.go` that becomes hard to navigate |
| Handlers return HTML fragments for HTMX routes | Keeps the frontend/backend contract explicit |
| No `pkg/` directory | This is a single-binary app with no public library surface; `pkg/` adds unnecessary indirection |
| No `util/` or `helpers/` packages | Grab-bag names obscure intent — put code in the package closest to where it is used |

---

## References

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) — widely adopted community layout reference
- [go.dev: Organizing a Go module](https://go.dev/doc/modules/layout) — official Go module layout guide
- [Alex Edwards: 11 tips for structuring Go projects](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects) — practical advice for real-world Go apps
