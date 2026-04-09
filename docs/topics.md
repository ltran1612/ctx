# Topics and storage

## How topics are stored

All data lives under `~/.ctx/` in two directories:

```
~/.ctx/
  active/
    api-redesign-plan/
      context.md
    login-bug/
      context.md
  archive/
    old-sprint-notes/
      context.md
```

Each topic is a folder named after its slug, containing a single `context.md` file. The store is created automatically on first use -- there is no init command.

## File format

A `context.md` file consists of YAML frontmatter followed by a Markdown body:

```markdown
---
id: a1b2c3d4
title: API redesign plan
slug: api-redesign-plan
status: active
tags:
  - backend
  - api
ticket: PROJ-789
created: 2026-03-15T10:00:00Z
modified: 2026-03-20T14:30:00Z
---

## Summary

We're moving from the v1 REST API to a new design...

## Notes

- Decided to keep REST over GraphQL
- Target completion: end of Q2

## Next Steps

- Draft OpenAPI spec
- Review with frontend team
```

The `status` field is derived from filesystem location (`active/` or `archive/`), not the frontmatter value. When a topic is saved, the status field is automatically corrected to match its location.

## Slugs

Slugs are derived from the topic title:

- Converted to kebab-case
- Truncated to a maximum of 5 words
- Collision suffixes (`-2`, `-3`, ...) are appended if needed

The slug is permanent -- renaming a topic with `--set-title` updates the title but not the slug.

## Topic resolution

Commands that accept a `<topic>` argument use `store.Resolve()`, which tries these strategies in order:

1. **Exact slug** in `active/` -- e.g. `api-redesign-plan`
2. **Short ID** scan -- e.g. `a1b2` (prefix of the full ID)
3. **Exact slug** in `archive/` -- matches archived topics
4. **Fuzzy title match** -- partial or approximate title matches

This means you can reference topics in whichever way is most convenient:

```bash
ctx view api-redesign-plan    # full slug
ctx view a1b2                 # short ID
ctx view "API redesign"       # fuzzy match
```

## Active vs. archived

- **Active** topics live in `~/.ctx/active/` and appear in `ctx list` by default.
- **Archived** topics live in `~/.ctx/archive/` and appear with `ctx list --archived` or `ctx list --all`.
- Use `ctx archive <topic>` and `ctx restore <topic>` to move between states.
- `ctx delete <topic>` permanently removes a topic from either location.

## Sections

The Markdown body can contain sections denoted by `##` headings. Some commands operate on sections directly:

- `ctx view --section "Notes"` prints only that section
- `ctx edit --append "text"` appends to the Notes section
- `ctx edit --prepend-note "text"` prepends a timestamped entry to the Notes section
