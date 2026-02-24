# Finances — Product Roadmap

## Overview

This roadmap lists the planned features for the personal finance tracker in implementation order. Each feature builds on the previous ones.

---

## Feature List

| # | Feature                          | Status  |
|---|----------------------------------|---------|
| 1 | Authentication                   | Planned |
| 2 | Multi-Currency with Money Pattern | Planned |
| 3 | Multi-Tenancy                    | Planned |
| 4 | Manual Transaction Tracking      | Planned |
| 5 | Account Spaces                   | Planned |
| 6 | Fix Costs                        | Planned |
| 7 | Transaction Categorization       | Planned |
| 8 | Dashboard & Reporting            | Planned |
| 9 | CSV Export                       | Planned |
| 10 | User Audit Log                  | Planned |

---

## Feature Details

### 1 — Authentication

Passwordless login via magic link:

- Login form accepts **email only**.
- If the email is unknown, the form extends with **Full Name** and a **Currency picker** (only implemented currencies are offered).
- A magic link is generated and, for now, **printed to the server console** instead of sending a real email.
- Sessions are managed with **JWT + refresh tokens**.
- Profile pictures use an **initials-based avatar** (no external image service required).

---

### 2 — Multi-Currency with Money Pattern

Safe, lossless handling of monetary values:

- All amounts are stored as **integer minor units** (e.g. cents for EUR).
- **Euro (EUR)** is the first implemented currency.
- The currency model is designed to accommodate additional currencies in future iterations.

---

### 3 — Multi-Tenancy

Complete data isolation between users:

- Every user can only see and modify **their own data**.
- All database queries are **scoped by user ID** at the repository layer — no cross-user data leakage.

---

### 4 — Manual Transaction Tracking

Core feature for recording financial activity:

- Users log income and expense transactions manually.
- Transaction fields:
  - Account space
  - Amount & currency
  - Date
  - Title
  - Description
  - Category

---

### 5 — Account Spaces

Organize finances across multiple accounts:

- Users can create any number of named account spaces (e.g. "Checking", "Savings", "Cash").
- Each account space has a configurable **display color** for visual distinction.

---

### 6 — Fix Costs

Recurring expense management:

- A fix cost entry has: **title**, **amount**, **start date**, and a **schedule** (e.g. monthly on day N, every X months).
- On the scheduled date, a **pending transaction** is created automatically.
- The user must **confirm the pending transaction** before it is booked as a regular transaction.

---

### 7 — Transaction Categorization

Flexible, user-driven categorization:

- Categories are created **lazily**: if a typed name does not exist yet, it is created on first use.
- An **LLM can suggest a category** based on the transaction title and description.
- The LLM provider will be decided in the feature-level spec.

---

### 8 — Dashboard & Reporting

At-a-glance financial overview:

- Displays **total balance** aggregated across all account spaces.
- Shows **spending trends over time**; the exact period options and granularity will be decided in the feature-level spec.

---

### 9 — CSV Export

Data portability:

- Allows users to export their transactions as a CSV file.
- Exact export scope and filter options (date range, account space, category, etc.) will be decided in the feature-level spec.

---

### 10 — User Audit Log

Transparency and traceability:

- Records changes to financial data (transactions, account spaces, fix costs, categories, etc.).
- Exact event types and retention policy will be decided in the feature-level spec.
