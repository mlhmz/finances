# Skill: /spec

Brainstorm and write a specification for a feature from the roadmap.

## Usage

```
/spec <feature-number>
```

## Instructions

You are helping design and specify a feature. Follow these steps exactly:

### Step 1 — Load the roadmap

Read `/spec/roadmap.md`. Find the feature matching the given number. If it does not exist, tell the user and stop.

### Step 2 — Determine file names

Derive a short snake_case summary from the feature title (3–5 words, lowercase, underscores).

Files you will write:
- `/spec/{nr}_{summary}_brainstorm.md`
- `/spec/{nr}_{summary}_spec.md`

### Step 3 — Iterative brainstorm

Think step by step. Each iteration:
1. Identify open questions, edge cases, and design choices.
2. Visualize relevant flows, layouts, or data structures in ASCII-Art where it helps.
3. List what you know vs. what is still unclear.

For every decision or question, present the options as Markdown checkboxes so the user can check their choice:

```
**Question: How should amounts be formatted?**
- [ ] Integer cents (e.g. 1099)
- [ ] Decimal string (e.g. "10.99")
- [ ] Float (e.g. 10.99)
```

Continue iterating until all decisions are captured as checkbox lists.

Write the full brainstorm — all iterations, ASCII diagrams, and checkbox questions — into `/spec/{nr}_{summary}_brainstorm.md`.

After writing the brainstorm file, show the user all unchecked checkbox questions and ask them to check their choices. Once the user returns with checked boxes (`[x]`), update the brainstorm file with the selections and proceed.

### Step 4 — Write the spec

Once all questions are resolved, write `/spec/{nr}_{summary}_spec.md` with the following sections:

```
# Feature {nr}: {Title}

## Overview
One-paragraph summary.

## Goals
Bulleted list of what this feature achieves.

## Non-Goals
What is explicitly out of scope.

## Data Model
Tables / fields / relationships (use ASCII tables or code blocks).

## API / Routes
HTTP method, path, request shape, response shape.

## UI / UX
Screen layout in ASCII-Art. Interaction flow description.

## Acceptance Criteria
Numbered, testable statements of done.

## Open Questions
Any remaining deferred decisions (move here if not yet resolved).
```

Only include sections that are relevant. Omit sections that do not apply to the feature.

### Step 5 — Summary

Tell the user:
- Path to the brainstorm file
- Path to the spec file
- Any remaining open questions
