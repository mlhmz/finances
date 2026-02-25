# Report 4: UI Navigation & Pagination Improvements

## Summary

Consolidated repeated navigation chrome into a shared layout template, unified the sidebar design across all pages, and improved the transactions page with a configurable page-size selector and a scrollable table body.

---

## Changes

### Centralized Layout (`views/layouts/app.html`)

Extracted the sidebar, bottom tab bar, user dropdown, and all shared CSS into a single Fiber layout template. All protected pages now pass `"layouts/app"` as the third argument to `c.Render()` and inject `"ActivePage"` for active-state highlighting.

**Before:** Each page (`index.html`, `transactions.html`, `profile.html`) was a full `<!DOCTYPE html>` document duplicating sidebar HTML and CSS.
**After:** Pages are bare content fragments. The layout provides the shell; `{{embed}}` receives the page content.

| File | Change |
|------|--------|
| `views/layouts/app.html` | **Created** — shared shell with sidebar, user dropdown, bottom nav, all layout CSS and JS |
| `views/index.html` | Rewritten as content fragment (welcome section + card grid) |
| `views/transactions.html` | Rewritten as content fragment (page-specific CSS/JS kept inline) |
| `views/profile.html` | Rewritten as content fragment — now uses the sidebar via the layout |
| `internal/handlers/home.go` | Added `"layouts/app"`, `"ActivePage": "home"`, `"Title": "Home"` |
| `internal/handlers/transaction.go` | Added `"layouts/app"`, `"ActivePage": "transactions"`, `"Title": "Transactions"` |
| `internal/handlers/profile.go` | Added `"layouts/app"`, `"ActivePage": "profile"`, `"Title": "Profile"` |

### Unified Sidebar User Menu

Replaced the per-page mix of "Profile" nav link + separate sign-out button with a single user dropdown button at the bottom of the sidebar. The dropdown shows **Profile** and **Sign out** options. This is now defined once in `layouts/app.html`.

- The `Profile` nav entry is removed from the sidebar nav links on all pages.
- Profile is accessible via the user dropdown and the mobile bottom tab bar.

### Profile Page Gets Sidebar

`views/profile.html` previously used a top navigation bar. It now inherits the sidebar layout and mobile bottom tab bar through `layouts/app.html`, matching the rest of the app.

---

### Per-Page Size Selector (Transactions)

Added a combobox to control how many transactions are shown per page: **5 / 10 / 15 / 20 / 25**, defaulting to **10**.

| File | Change |
|------|--------|
| `internal/handlers/transaction.go` | Reads `?pageSize=` query param; validates against allowed set; passes `PageSize` to template |
| `views/transactions.html` | Selector in the pagination footer; pagination prev/next links preserve `pageSize` |

The selector resets to page 1 on change (`?page=1&pageSize=N`) to avoid stale offsets. Invalid or missing values fall back to 10.

---

### Scrollable Transaction Table

Made the transaction list scroll internally so the pagination bar is always visible without page-level scrolling.

| Element | CSS |
|---------|-----|
| `.content` | `height: 100vh; overflow: hidden` — pins the content area to the viewport |
| `.list-card` | `flex: 1; min-height: 0; display: flex; flex-direction: column` — fills remaining space |
| `#transaction-list` | `flex: 1; min-height: 0; overflow-y: auto` — scrollable row area |
| `.pagination` | `flex-shrink: 0` — always pinned at the bottom of the card |

The pagination bar (page-size selector on the left, prev/next on the right) is visible on both desktop and mobile without scrolling.

---

## Commits

| Hash | Message |
|------|---------|
| `f4dea74` | `refactor(views): centralize nav into shared layout template` |
| `3a7bd46` | `feat(transactions): add per-page size selector (5/10/15/20/25)` |
| `73bb94b` | `fix(transactions): move page size selector to pagination footer` |
| `b45d772` | `feat(transactions): make transaction list scroll internally` |
