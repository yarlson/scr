# scr

Automated terminal screenshots. Script interactions, capture PNGs.

```bash
scr bash "Type 'echo hello' Enter"
```

## What it does

Runs any CLI or TUI command in a real terminal (ttyd + headless Chrome), executes your scripted keypresses, captures PNG screenshots at each step. Works with static commands, interactive prompts, full-screen TUIs, REPLs, editors.

## Installation

### Prerequisites

- **Go 1.21+**
- **ttyd** — terminal sharing daemon
- **Chrome/Chromium** — for headless screenshot capture

### Install ttyd

```bash
# macOS
brew install ttyd

# Ubuntu/Debian
apt-get install ttyd

# Fedora/RHEL
dnf install ttyd
```

### Install scr

```bash
go install github.com/yarlson/scr@latest
```

Or build from source:

```bash
git clone https://github.com/yarlson/scr.git
cd scr
go build -o scr ./cmd
```

## Quick Start

```bash
# Capture static output
scr "git --help"

# Type and execute a command
scr bash "Type 'ls -la' Enter"

# Navigate a menu
scr "seq 100 | fzf" "Down 3 Enter"
```

## Usage

```
scr [options] <command> [script]
```

| Argument    | Description                     |
| ----------- | ------------------------------- |
| `<command>` | Shell command to run (required) |
| `[script]`  | Actions to perform (optional)   |

### Options

| Flag         | Short | Default         | Description         |
| ------------ | ----- | --------------- | ------------------- |
| `--out`      | `-o`  | `./screenshots` | Output directory    |
| `--interval` | `-i`  | `500ms`         | Screenshot interval |
| `--timeout`  | `-t`  | `60s`           | Max execution time  |
| `--port`     | `-p`  | `7681`          | ttyd server port    |
| `--verbose`  | `-v`  | `false`         | Debug output        |

## Script Actions

| Action             | Description                    | Example                   |
| ------------------ | ------------------------------ | ------------------------- |
| `Type 'text'`      | Type text (50ms between chars) | `Type 'hello world'`      |
| `Type@30ms 'text'` | Type with custom speed         | `Type@30ms 'fast'`        |
| `Sleep <duration>` | Pause                          | `Sleep 500ms`, `Sleep 2s` |
| `Enter`            | Press Enter                    | `Enter`                   |
| `<Key> N`          | Press key N times              | `Down 3`                  |
| `<Key>@<duration>` | Press key after delay          | `Enter@200ms`             |
| `Ctrl+<key>`       | Control combo                  | `Ctrl+C`, `Ctrl+D`        |

### Supported Keys

`Enter` `Tab` `Escape` `Space` `Backspace` `Delete` `Up` `Down` `Left` `Right` `Home` `End` `PageUp` `PageDown`

## Examples

### Static Output

Capture command output with no interaction:

```bash
scr "git --help"
scr "ls -la"
scr "cat README.md"
```

### Interactive Bash

```bash
# Type and execute
scr bash "Type 'echo hello world' Enter"

# With timing
scr bash "Sleep 1s Type 'date' Enter Sleep 500ms"
```

### Menu Navigation

```bash
# fzf: move down 3, select
scr "seq 1 100 | fzf" "Down 3 Enter"

# fzf: search then select
scr "find . -type f | fzf" "Type 'main' Sleep 200ms Enter"

# gum choose
scr "gum choose Apple Banana Cherry" "Down Down Enter"
```

### Text Editors

```bash
# vim: insert, type, save, quit
scr vim "Type 'i' Type 'Hello World' Escape Type ':wq' Enter"

# nano: type and exit
scr nano "Type 'Hello World' Ctrl+X Type 'y' Enter"
```

### REPLs

```bash
# Python
scr python3 "Type 'print(\"hello\")' Enter Sleep 500ms Ctrl+D"

# Node
scr node "Type '2 + 2' Enter Sleep 300ms Type '.exit' Enter"
```

### Long-Running Processes

```bash
# Watch then exit
scr htop "Sleep 3s Ctrl+C"

# Custom timeout
scr -t 10s top "Sleep 5s Type 'q'"
```

### Custom Output

```bash
# Different directory
scr -o ./demo bash "Type 'whoami' Enter"

# Faster screenshots
scr -i 200ms bash "Type 'ls' Enter"
```

## Output

Screenshots are saved as `screenshot_001.png`, `screenshot_002.png`, etc.

Capture sequence:

1. Initial terminal state
2. Periodic snapshots (based on `--interval`)
3. Final state after all actions complete

## Troubleshooting

### ttyd not found

```
Error: ttyd binary not found
```

Install ttyd and ensure it's in PATH:

```bash
brew install ttyd  # macOS
which ttyd         # verify
```

### Port already in use

```
Error: listen tcp :7681: address already in use
```

Use a different port:

```bash
scr -p 8080 bash "Type 'hello' Enter"
```

### Blank screenshots

1. Increase interval: `scr -i 1s ...`
2. Add initial sleep: `scr bash "Sleep 1s Type 'hello' Enter"`
3. Run with `-v` to debug

### Timeout errors

Increase timeout for slow commands:

```bash
scr -t 120s slow-command "..."
```

### Orphaned processes

If ttyd or Chrome processes remain after interruption:

```bash
pkill -f ttyd
pkill -f chrome
```

### Parse errors

```
Error: expected quoted string after Type
```

Text must be quoted:

```bash
# Wrong
scr bash "Type hello"

# Right
scr bash "Type 'hello'"
```

## License

[MIT](LICENSE)
