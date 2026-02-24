# 2 — Structure Doc Update

**Date:** 2026-02-24

## What was done

- Rewrote `docs/structure.md` to align with Go enterprise project layout conventions.
- Expanded the bare directory skeleton into a fully annotated tree with rationale for each directory.
- Added a conventions table explaining why `internal/` is used, why `pkg/` and `util/` are omitted, and how `views/` ties into the Fiber template engine.
- Added a References section pointing to golang-standards/project-layout, the official go.dev layout guide, and Alex Edwards' practical structuring tips.

## Result

`docs/structure.md` now serves as an authoritative layout reference for contributors, covering directory purpose, Go compiler enforcement of `internal/`, and conventions specific to the Fiber + HTMX stack.
