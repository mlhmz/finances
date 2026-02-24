# 1 — Initial Setup

**Date:** 2026-02-24

## What was done

- Initialized Go module `github.com/mlhmz/finances` (Go 1.25.4).
- Installed dependencies:
  - `github.com/gofiber/fiber/v2` v2.52.11
  - `github.com/gofiber/template/html/v2` v2.1.3
  - `gorm.io/gorm` v1.31.1
  - `gorm.io/driver/sqlite` v1.6.0 (uses `mattn/go-sqlite3`)
- Created `main.go` with:
  - GORM SQLite initialization
  - Fiber app with HTML template engine pointing to `views/`
  - `GET /` renders `views/index.html`
  - `GET /greet` returns an HTMX-targeted HTML fragment
- Created `views/index.html` with:
  - Web Awesome 2.0.0-alpha.10 (CDN) for UI components and theming
  - HTMX 2.0.4 (CDN) for partial updates
  - A `<wa-button>` that triggers `GET /greet` and injects the response via HTMX
- Created `docs/architecture.md` documenting the request flow and conventions.
- Updated `CLAUDE.md` with full stack reference and project structure.

## Result

Running `go run main.go` starts the server on `http://localhost:3000` with a Hello World page styled with Web Awesome.
