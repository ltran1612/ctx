# ctx

`ctx` is a lightweight CLI for managing project-scoped context. It stores topics as Markdown files with YAML frontmatter in a `~/.ctx/` directory, making them easy to read, search, and edit with any tool.

Designed for developers and AI agents that need to track decisions, notes, and running context across tickets and projects.

## Features

- **Markdown-native** -- each topic is a `context.md` file you can read with any editor
- **Fuzzy resolution** -- reference topics by slug, short ID, or partial title
- **Tagging and tickets** -- organize with tags and link to external trackers
- **Section editing** -- append or prepend notes to specific sections without opening an editor
- **Archive lifecycle** -- archive and restore topics instead of deleting them
- **No init required** -- the store is created automatically on first use

## Installation

Requires Go 1.26+.

```bash
git clone https://github.com/user/ctx.git
cd ctx
go build -o ctx .
```

Move the binary somewhere on your `$PATH`:

```bash
mv ctx /usr/local/bin/
```

## Quick start

```bash
# Create a topic (opens your $EDITOR)
ctx create "API redesign plan"

# Create without opening an editor
ctx create "Bug triage notes" --tag bug --ticket PROJ-1234 --no-edit

# List active topics
ctx list

# View a topic (by slug, short ID, or partial title)
ctx view api-redesign-plan

# Append a note without opening an editor
ctx edit api-redesign --append "Decided to use REST over GraphQL"

# Prepend a timestamped note
ctx edit api-redesign --prepend-note "Sync with backend team complete"

# Search across topics
ctx search "redesign"

# Archive when done
ctx archive api-redesign-plan
```

## Documentation

- [Getting started](docs/getting-started.md) -- setup, first topic, editor configuration
- [Command reference](docs/commands.md) -- all commands, flags, and examples
- [Topics and storage](docs/topics.md) -- how topics are stored and resolved

## Running tests

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) for details.
