# catls

Recursively list files and print their contents in XML, Markdown, or JSON — handy for feeding a codebase to an LLM or producing a single-document snapshot of a project.

## Features

- XML (default), Markdown, or JSON output
- Recursive scan with glob include/exclude filters
- Sensible default ignore list (`node_modules`, `.direnv`, `vendor`, `.git`, `dist`, `build`, …)
- Interactive file picker (bubbletea TUI) for selecting a subset before printing
- Optional line numbers, binary-file skipping, line-pattern filtering
- Configurable base path for relative output paths

## Install

### With Nix (flake)

```sh
nix build
./result/bin/catls --help
```

Or run without installing:

```sh
nix run github:connerohnesorge/catls -- --help
```

### With Go

```sh
go install github.com/connerohnesorge/catls@latest
```

### Dev shell

```sh
nix develop
```

Provides Go 1.25, `golangci-lint`, `gopls`, `gotestsum`, `air`, `goreleaser`, formatters, and helper scripts (`lint`, `tests`, `dx`).

## Usage

```
catls [directory] [files...] [flags]
```

### Flags

| Flag | Description |
| --- | --- |
| `-a, --all` | Include hidden files |
| `-r, --recursive` | Recurse into subdirectories |
| `-n, --line-numbers` | Prefix each line with its line number |
| `-f, --format` | Output format: `xml` (default), `json`, `markdown` |
| `-I, --interactive` | Launch TUI to pick files before printing |
| `--globs` | Include-only glob (repeatable) |
| `--ignore-globs` | Exclude glob (repeatable) |
| `--ignore-dir` | Directory names to skip (repeatable) |
| `--pattern` | Only print lines matching this glob |
| `--omit-bins` | Skip binary files entirely |
| `--relative-to` | Base path for the paths shown in output |
| `--debug` | Print debug info to stderr |

### Examples

Print the current project as XML:

```sh
catls -r .
```

Only Go files, with line numbers, as Markdown:

```sh
catls -r --globs '*.go' -n -f markdown .
```

JSON output, skipping tests and binaries:

```sh
catls -r --ignore-globs '*_test.go' --omit-bins -f json .
```

Interactive selection from a recursive scan:

```sh
catls -r -I .
```

Interactive keys: `↑/↓` or `k/j` to move, `space/x` to toggle, `a` select all, `A` deselect all, `enter` confirm, `q`/`esc` cancel.

## Output formats

- **xml** — `<files><file path="…"><type>…</type><content>…</content></file></files>`, with binary files marked via `<binary>true</binary>`
- **markdown** — fenced code blocks per file with language inferred from file type
- **json** — structured array of file objects; easy to post-process

## License

MIT
