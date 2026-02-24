# Architecture

## Request Flow

```
Browser
  │
  ├─ Full page load ──► GET /  ──► Fiber renders views/index.html ──► Browser
  │
  └─ HTMX partial  ──► GET /greet ──► Fiber returns HTML fragment ──► HTMX swaps target element
```

## Database

GORM is configured with the SQLite driver. The database file `finances.db` is created automatically in the working directory on startup.

Migration is performed via `db.AutoMigrate(&Model{})` — add model structs and call AutoMigrate in `main.go`.

## Frontend

Web Awesome components (`<wa-button>`, `<wa-icon>`, etc.) are loaded via CDN (early.webawesome.com). HTMX is also loaded from CDN (unpkg). No build step is required.

HTMX attributes on elements send requests to the Go backend; Fiber responds with HTML fragments that HTMX swaps into the DOM.

## Adding Features

1. Define a GORM model struct in a `models/` package.
2. Call `db.AutoMigrate` in `main.go`.
3. Add Fiber routes (GET/POST) that query/mutate the DB.
4. Return HTML fragments for HTMX endpoints, full renders for page routes.
5. Write a report in `reports/{nr}_{change}.md`.
