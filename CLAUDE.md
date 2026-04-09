# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build -o ctx .          # build binary
go test ./...              # run all tests
go test ./internal/store/... -run TestCreate  # run a single test by name
go test ./... -v           # verbose test output with names
```

## Architecture

`ctx` is a project-scoped CLI (like `.git/`) where `ctx init` creates a `.ctx/active/` and `.ctx/archive/` directory in the working directory. Each topic is a folder containing one `context.md` with YAML frontmatter.

### Data flow

Every command calls `cmd/helpers.go:openStore()` → `store.Open()` (walks up the directory tree to find `.ctx/`) → returns a `*store.Store`. Commands then call store methods which read/write `context.md` files via the `internal/frontmatter` package.

### Package responsibilities

- **`internal/frontmatter`** — parse/serialize the `---YAML---` block and body of `context.md` files. Also owns section extraction (`Section`) and in-place text mutation (`AppendToSection`, `PrependToSection`) used by `ctx edit --append/--prepend-note`.
- **`internal/slug`** — derives a kebab-case slug from a title (max 5 words) and handles collision suffixes (`-2`, `-3`).
- **`internal/fuzzy`** — thin wrapper around `sahilm/fuzzy` for title and full-text search.
- **`internal/store`** — all filesystem operations: `Init`, `Open`, `Create`, `All`, `Resolve`, `Save`, `Reload`, `Archive`, `Restore`, `Delete`, `Search`. Source of truth for topic status is filesystem location (`active/` vs `archive/`), not the frontmatter `status` field — `Save` auto-corrects the field.
- **`internal/output`** — terminal formatting helpers (table, colors, markdown rendering). All user-visible output goes through here.
- **`cmd/`** — one file per subcommand, wired together in `cmd/root.go`. Pure helper functions (`containsTag`, `removeTags`, `filterByTags`, `hasAllTags`, `sortTopics`) live in `cmd/helpers_test.go` alongside their tests.

### Topic resolution

Every command that takes a `<topic>` argument calls `store.Resolve()`, which tries in order: exact slug in `active/` → short ID scan → exact slug in `archive/` → fuzzy title match. `--json` mode does not exist; AI agents read the same plain text output as humans.

### Editor integration

`ctx create` and `ctx edit` open `$EDITOR` (fallback: `$VISUAL`, then `vi`). After the editor closes, `edit` calls `store.Reload(t)` before `store.Save(t)` to avoid overwriting in-editor changes with stale in-memory state.
