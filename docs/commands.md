# Command reference

All commands support the global `--no-color` flag to disable ANSI color output.

## ctx create

Create a new topic.

```
ctx create [title] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--tag` | string (repeatable) | Add a tag |
| `--ticket` | string | Ticket reference (e.g. PROJ-1234) |
| `--no-edit` | bool | Create without opening an editor |

If `title` is omitted, you will be prompted for one. Opens `$EDITOR` unless `--no-edit` is set.

```bash
ctx create "Database migration plan" --tag backend --ticket DB-101
ctx create "Quick note" --no-edit
```

## ctx view

View a topic.

```
ctx view <topic> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--section` | string | Print only a named section (e.g. "Next Steps") |
| `--raw` | bool | Print raw markdown without formatting |

The `<topic>` argument is resolved by slug, short ID, or fuzzy title match. See [topic resolution](topics.md#topic-resolution).

```bash
ctx view database-migration
ctx view db3f
ctx view database-migration --section "Open Questions"
ctx view database-migration --raw
```

## ctx edit

Edit a topic interactively or update specific fields.

```
ctx edit <topic> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--append` | string | Append text to the Notes section |
| `--prepend-note` | string | Prepend a timestamped note to the Notes section |
| `--set-title` | string | Update the title (slug remains unchanged) |
| `--add-tag` | string (repeatable) | Add a tag |
| `--remove-tag` | string (repeatable) | Remove a tag |
| `--set-ticket` | string | Update ticket reference |

When no flags are passed, opens `$EDITOR` for interactive editing. Multiple flags can be combined in a single call.

```bash
# Open in editor
ctx edit database-migration

# Non-interactive updates
ctx edit database-migration --append "Decided on blue-green deployment"
ctx edit database-migration --prepend-note "Migration tested on staging"
ctx edit database-migration --add-tag urgent --set-ticket DB-202
ctx edit database-migration --remove-tag draft
ctx edit database-migration --set-title "DB migration plan (v2)"
```

## ctx list

List topics.

```
ctx list [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string (repeatable) | | Filter by tag (AND logic when repeated) |
| `--archived` | bool | false | Show archived topics only |
| `--all` | bool | false | Show active and archived topics |
| `--sort` | string | `updated` | Sort by: `updated`, `created`, `title` |
| `--format` | string | `table` | Output format: `table`, `ids`, `slugs` |

```bash
ctx list
ctx list --tag backend --tag urgent
ctx list --archived --sort title
ctx list --all --format ids
ctx list --format slugs
```

## ctx search

Fuzzy search across topics.

```
ctx search <query> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--full-text` | bool | false | Search body content in addition to titles |
| `--limit` | int | 10 | Maximum number of results |

Results are ranked by match score. Archived topics are shown with an "(archived)" marker.

```bash
ctx search "migration"
ctx search "deploy" --full-text
ctx search "auth" --limit 5
```

## ctx archive

Archive a topic.

```
ctx archive <topic>
```

Moves the topic from active to archived status. Returns an error if the topic is already archived.

```bash
ctx archive database-migration
```

## ctx restore

Restore an archived topic.

```
ctx restore <topic>
```

Moves the topic from archived back to active status. Returns an error if the topic is not archived.

```bash
ctx restore database-migration
```

## ctx delete

Permanently delete a topic.

```
ctx delete <topic> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--confirm` | bool | Skip the confirmation prompt |

Deletion is irreversible. Without `--confirm`, you will be prompted to confirm.

```bash
ctx delete old-notes
ctx delete old-notes --confirm
```

## ctx show-path

Print the filesystem path to a topic's `context.md` file.

```
ctx show-path <topic>
```

Useful for piping to other tools or opening the file directly.

```bash
ctx show-path database-migration
cat $(ctx show-path database-migration)
$EDITOR $(ctx show-path database-migration)
```
