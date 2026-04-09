# Getting started

## Prerequisites

- Go 1.26 or later
- A terminal with a text editor configured via `$EDITOR` or `$VISUAL` (falls back to `vi`)

## Build and install

```bash
git clone https://github.com/user/ctx.git
cd ctx
go build -o ctx .
```

Move the binary to a directory on your `$PATH`:

```bash
mv ctx /usr/local/bin/
```

Verify the installation:

```bash
ctx --help
```

## Create your first topic

```bash
ctx create "Sprint 12 planning"
```

This opens your editor with a new `context.md` file pre-populated with YAML frontmatter. Write your notes in the body, save, and close the editor.

To skip the editor and create an empty topic:

```bash
ctx create "Sprint 12 planning" --no-edit
```

## Add metadata

Tags and ticket references can be set at creation time:

```bash
ctx create "Login bug" --tag bug --tag auth --ticket AUTH-456 --no-edit
```

Or added later:

```bash
ctx edit login-bug --add-tag urgent --set-ticket AUTH-456
```

## View and list topics

```bash
# List all active topics
ctx list

# View a single topic
ctx view login-bug

# View only the "Next Steps" section
ctx view login-bug --section "Next Steps"
```

## Edit without opening an editor

Append text to the Notes section:

```bash
ctx edit login-bug --append "Root cause identified: session cookie not refreshed"
```

Prepend a timestamped note:

```bash
ctx edit login-bug --prepend-note "Patch deployed to staging"
```

## Editor configuration

`ctx` checks these environment variables in order when opening an editor:

1. `$EDITOR`
2. `$VISUAL`
3. Falls back to `vi`

Set your preferred editor in your shell profile:

```bash
# ~/.bashrc or ~/.zshrc
export EDITOR="code --wait"   # VS Code
export EDITOR="nvim"          # Neovim
export EDITOR="nano"          # Nano
```

## Disabling color

Pass `--no-color` to any command, or set the `NO_COLOR` environment variable:

```bash
ctx list --no-color
NO_COLOR=1 ctx view my-topic
```

## Next steps

- See the [command reference](commands.md) for all commands and flags
- See [topics and storage](topics.md) to understand how data is stored and resolved
