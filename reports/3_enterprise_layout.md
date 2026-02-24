# Report 3: Enterprise Layout Restructure

## Summary

Migrated the project from a single root `main.go` to a Go enterprise layout with `cmd/` and `internal/` packages.

## Changes

### Files Created

| File | Purpose |
|------|---------|
| `cmd/finances/main.go` | Thin entry point: config, DB init, Fiber setup, routes, server start |
| `internal/config/config.go` | Typed `Config` struct with `Default()` factory |
| `internal/db/db.go` | `Connect(dbPath)` wraps GORM initialization |
| `internal/handlers/home.go` | `Index` and `Greet` route handlers |
| `internal/models/doc.go` | Placeholder package for future GORM model structs |

### Files Deleted

| File | Reason |
|------|--------|
| `main.go` (root) | Replaced by `cmd/finances/main.go` |

### Files Updated

| File | Change |
|------|--------|
| `CLAUDE.md` | Updated project structure tree and run command |

## Run Command Change

```bash
# Before
go run main.go

# After
go run ./cmd/finances
```

## Notes

- `views/` remains at the project root; Fiber resolves `./views` relative to the working directory at runtime — always run from project root.
- Module name `github.com/mlhmz/finances` is unchanged.
- The stale `finances` binary at root is superseded by `go build -o finances ./cmd/finances`.
