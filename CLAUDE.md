# Finances

A personal finance tracker built with Go, SQLite, Fiber, Gorm, HTMX, and Web Awesome.

## Stack

| Layer       | Technology                          | Version  |
|-------------|-------------------------------------|----------|
| Language    | Go                                  | 1.25.4   |
| Web Framework | Fiber v2                          | v2.52.11 |
| Template Engine | Fiber HTML template (html/v2) | v2.1.3   |
| ORM         | Gorm                                | v1.31.1  |
| Database    | SQLite (via mattn/go-sqlite3)       | v1.14.22 |
| Frontend    | HTMX                                | 2.0.4    |
| UI Components | Web Awesome                       | 2.0.0-alpha.10 |

## Project Structure

```
finances/
├── cmd/
│   └── finances/
│       └── main.go    # Application entry point (thin: config, DB init, routes, server)
├── internal/
│   ├── config/
│   │   └── config.go  # Typed config struct with defaults
│   ├── db/
│   │   └── db.go      # GORM initialization
│   ├── handlers/
│   │   └── home.go    # Route handler functions
│   └── models/
│       └── doc.go     # GORM model structs (placeholder)
├── views/
│   └── index.html     # Main HTML template (Web Awesome + HTMX)
├── docs/              # Advanced documentation
├── reports/           # Change reports
├── go.mod
├── go.sum
└── finances.db        # SQLite database (created at runtime)
```

## Running the App

```bash
go run ./cmd/finances
# Visit http://localhost:3000
```

## Architecture

- **cmd/finances/main.go**: Thin entry point — loads config, inits DB, builds Fiber app, registers routes, starts server on `:3000`.
- **internal/config**: Typed `Config` struct with `Default()` factory.
- **internal/db**: `Connect(dbPath)` wraps GORM initialization.
- **internal/handlers**: Route handler functions, one file per domain area.
- **internal/models**: GORM model structs (currently a placeholder package).
- **views/index.html**: Server-rendered HTML template using Web Awesome components and HTMX for partial-page updates without full reloads.
- **Database**: SQLite file `finances.db` created in the working directory on first run.

## Key Conventions

- Templates live in `views/` with `.html` extension.
- Routes return `c.Render("template-name", fiber.Map{...})` for full pages, or plain strings/JSON for HTMX partials.
- Reports for every change are written to `reports/{nr}_{change}.md`.
- Advanced docs (architecture decisions, API reference, etc.) go in `docs/`.
