# Feature 4: Manual Transaction Tracking — Brainstorm

## What We Know

- Users log income and expense transactions manually.
- Transaction fields per roadmap: account space, amount & currency, date, title, description, category.
- Money pattern is implemented (integer minor units, embedded GORM columns).
- Multi-tenancy repository layer is established — all queries scoped by `userID`.
- Account Spaces (Feature 5) and Transaction Categorization (Feature 7) are **not yet implemented**.
- The existing UI is HTMX + Web Awesome; new views should follow the same pattern.

---

## Iteration 1 — Fundamental Design Decisions

### What is still unclear

1. **Transaction type** — how do we distinguish income from expense?
2. **Account space** — Feature 5 is not yet built. How do we handle the field now?
3. **Category** — Feature 7 is not yet built. How do we handle the field now?
4. **CRUD scope** — which operations are in scope for this feature?
5. **Currency** — does the user pick a currency per transaction, or inherit from their profile?
6. **Date granularity** — date-only vs. datetime?

---

### Question 1: How do we distinguish income from expense?

**Question: How is transaction type (income vs expense) represented?**
- [ ] Explicit `type` field with values `income` / `expense`
- [x] Amount sign: positive = income, negative = expense (no separate field)

---

### Question 2: Account space field — Feature 5 is not yet built

**Question: How do we handle the account space field before Feature 5 exists?**
- [x] Omit it entirely for now; add it when Feature 5 is implemented
- [ ] Include a nullable `AccountSpaceID` foreign key (no spaces exist yet, field stays NULL)
- [ ] Use a plain text field for now and migrate it to a FK in Feature 5

---

### Question 3: Category field — Feature 7 is not yet built

**Question: How do we handle the category field before Feature 7 exists?**
- [x] Omit it entirely for now; add it when Feature 7 is implemented
- [ ] Include a nullable `CategoryID` foreign key (no categories exist yet, field stays NULL)
- [ ] Use a plain text field for now and migrate it to a FK in Feature 7

---

### Question 4: CRUD scope for this feature

**Question: Which operations are in scope for Feature 4?**
- [ ] Create + Read (list) only — edit/delete deferred
- [ ] Create + Read + Delete
- [x] Full CRUD: Create, Read, Update, Delete

---

### Question 5: Currency per transaction

**Question: Can a transaction use any supported currency, or only the user's profile currency?**
- [ ] Only the user's profile currency (simpler — no currency picker on the form)
- [x] Any supported currency (more flexible — currency picker on every transaction)

---

### Question 6: Date granularity

**Question: What granularity should the transaction date have?**
- [ ] Date only (YYYY-MM-DD) — simpler; most personal finance apps use this
- [x] Full datetime (with time-of-day) — more precise

but time is optional, the transaction date should be prefilled with the current date and time

---

## Iteration 2 — UI/UX & Interaction Design

### What We Know So Far

- Full CRUD on transactions.
- Amount sign = income/expense (positive/negative).
- Any supported currency, picked per transaction.
- Datetime field: required, time optional, prefilled with now.
- Account space and category omitted.
- Fields: title, description, amount, currency, date.

### What is Still Unclear

1. **Income/expense UX** — with signed amounts, how does the user indicate direction in the form?
2. **Currency default** — which currency is pre-selected in the picker?
3. **Required vs optional fields** — is title required? Is description optional?
4. **List layout** — what columns/info is shown, and how is it sorted?
5. **Edit UX** — inline list editing, a modal, or a separate edit page?
6. **Delete confirmation** — instant delete or a confirmation step?
7. **Navigation** — where does the Transactions section live in the nav?
8. **Pagination** — how does the list handle many transactions?

---

### Question 7: Income/expense UX in the form

The amount is stored as a signed integer (positive = income, negative = expense). The UI must make this clear.

**Question: How does the user indicate income vs expense in the form?**
- [x] Toggle / segmented control (Income | Expense) that flips the sign — user always enters a positive number
- [ ] User enters a signed number directly (e.g. `-12.50`)
- [ ] Two separate amount fields (credit / debit)

---

### Question 8: Default currency in the transaction form

**Question: What is the default selected currency in the transaction form?**
- [x] User's profile currency (pre-selected, user can change it)
- [ ] No default — user must pick a currency explicitly

---

### Question 9: Required vs optional fields

**Question: Which fields are required when creating/editing a transaction?**
- [x] Title required, description optional
- [ ] Both title and description optional (only amount, currency, and date are required)

---

### Question 10: List sort order

**Question: How is the transaction list sorted by default?**
- [x] Newest first (date descending)
- [ ] Oldest first (date ascending)

---

### Question 11: Edit UX

**Question: How does the user edit an existing transaction?**
- [ ] Separate edit page (navigates to `/transactions/{id}/edit`)
- [ ] Inline editing within the list row
- [x] Modal/drawer overlay

Take a mobile first approach, drawer for mobile is important, normal modal for browser is good
(Breakpoint: same as existing app — 480px)

---

### Question 12: Delete confirmation

**Question: How is deletion handled?**
- [ ] Instant delete (no confirmation)
- [x] Inline confirmation (e.g. "Are you sure?" appears in the row)
- [ ] Browser confirm dialog

---

### Question 13: Pagination

**Question: How does the transaction list handle large volumes?**
- [ ] No pagination — show all (simple, fine for personal use)
- [x] Fixed page size (e.g. 20 per page) with previous/next controls
- [ ] Infinite scroll / load more

---

## Iteration 3 — Navigation, List Design & Data Model Details

### What We Know So Far

- Toggle (Income | Expense) + positive number input; stored as signed integer.
- Profile currency pre-selected in picker; any supported currency allowed.
- Title required, description optional (textarea).
- Sorted newest first; paginated 20/page.
- Create/Edit via modal (desktop) or drawer (mobile); breakpoint 480px.
- Delete with inline "Are you sure?" confirmation.

### What is Still Unclear

1. **Navigation placement** — where does the Transactions link appear?
2. **List row content** — what data is visible per row?
3. **Transaction ID** — UUID or auto-increment integer?
4. **Description in list** — shown in the row or only inside the modal?
5. **Pagination controls** — prev/next arrows only, or also page numbers?

---

### ASCII sketch of candidate list row layouts

```
Option A — Compact (date + title + signed amount)
┌─────────────────────────────────────────────────────────┐
│ 25 Feb 2026  Grocery shopping           − €42.50  [···] │
│ 24 Feb 2026  Freelance invoice          + €800.00 [···] │
└─────────────────────────────────────────────────────────┘

Option B — Two-line (adds description preview below title)
┌─────────────────────────────────────────────────────────┐
│ 25 Feb 2026  Grocery shopping           − €42.50  [···] │
│              REWE, Berlin                                │
│ 24 Feb 2026  Freelance invoice          + €800.00 [···] │
│              Client: Acme Corp                           │
└─────────────────────────────────────────────────────────┘
```

---

### Question 14: Navigation placement

**Question: Where does the Transactions section appear in the navigation?**
- [ ] Top navbar link (alongside the avatar/profile link)
- [x] Dedicated bottom tab bar (mobile-style, always visible)
- [x] Sidebar (desktop only, collapsible)

bottom tab bar, mobile, sidebar desktop

---

### Question 15: List row content

**Question: Which layout for the transaction list rows?**
- [ ] Option A — compact single line: date · title · signed amount · action menu
- [x] Option B — two lines: date · title · signed amount · action menu, + description preview below

---

### Question 16: Description in the list

**Question: Is the description shown in the list row at all?**
- [ ] No — description only visible when opening the modal/drawer
- [x] Yes — shown as a short preview below the title (Option B above)

---

### Question 17: Transaction primary key

**Question: What primary key type should Transaction use?**
- [x] UUID string (consistent with User model)
- [ ] Auto-increment integer (simpler, shorter URLs)

---

### Question 18: Pagination controls

**Question: What does the pagination UI look like?**
- [ ] Previous / Next buttons only (no page numbers)
- [x] Previous / Next + current page indicator (e.g. "Page 2 of 7")

---
