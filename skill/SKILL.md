---
name: scr
description: Automated terminal screenshots. Run CLI/TUI commands, script interactions, capture PNGs. Use when generating documentation, testing terminal apps, or capturing command output sequences.
license: MIT
---

Capture terminal screenshots by running commands through ttyd + headless Chrome, executing scripted keypresses, saving PNG sequences.

## Syntax

```
scr [flags] <command> [script]
```

## Flags

| Flag | Default         | Purpose             |
| ---- | --------------- | ------------------- |
| `-o` | `./screenshots` | Output directory    |
| `-i` | `500ms`         | Screenshot interval |
| `-t` | `60s`           | Timeout             |
| `-p` | `7681`          | ttyd port           |
| `-v` | `false`         | Verbose             |

## Script Actions

| Action     | Syntax             | Behavior                                                           |
| ---------- | ------------------ | ------------------------------------------------------------------ |
| Type       | `Type 'text'`      | Type text (50ms/char default); supports spaces + ASCII punctuation |
| Type speed | `Type@30ms 'text'` | Type with custom per-char delay                                    |
| Sleep      | `Sleep 500ms`      | Pause for duration (ms or s)                                       |
| Key        | `Enter`            | Press key once                                                     |
| Key repeat | `Down 3`           | Press key N times                                                  |
| Key delay  | `Enter@200ms`      | Delay before keypress                                              |
| Ctrl combo | `Ctrl+C`           | Send control character                                             |

## Supported Keys

`Enter` `Tab` `Escape` `Space` `Backspace` `Delete` `Up` `Down` `Left` `Right` `Home` `End` `PageUp` `PageDown`

---

## Common Patterns

### Static capture (no interaction)

```bash
scr "ls -la"
scr "git status"
scr "cat README.md"
```

### Type and execute

```bash
scr bash "Type 'echo hello' Enter"
scr bash "Type \"echo 'hello / ?'\" Enter"
scr bash "Type 'npm test' Enter Sleep 5s"
```

### Menu navigation (fzf, gum, etc.)

```bash
scr "seq 1 20 | fzf" "Down 5 Enter"
scr "gum choose A B C" "Down Down Enter"
```

### Editors

```bash
scr vim "Type 'i' Type 'Hello' Escape Type ':wq' Enter"
scr nano "Type 'Hello' Ctrl+X Type 'y' Enter"
```

### REPLs

```bash
scr python3 "Type 'print(1+1)' Enter Sleep 500ms Ctrl+D"
scr node "Type '2+2' Enter Sleep 300ms Type '.exit' Enter"
```

### Long-running processes

```bash
scr htop "Sleep 3s Ctrl+C"
scr -t 120s "npm run build" "Sleep 60s"
```

### Custom output

```bash
scr -o ./docs/images "git log --oneline"
scr -i 200ms bash "Type 'ls' Enter"
```

---

## Error Handling

Parse errors include position:

```
Error: parse script: position 23: expected quoted string after Type
```

Unknown key:

```
Error: unknown key 'Foo' at position 15; valid keys: Enter, Tab, ...
```

## Prerequisites

Requires `ttyd` in PATH:

```bash
# macOS
brew install ttyd

# Linux
apt-get install ttyd
```

Chrome/Chromium must be available (chromedp auto-detects).

## Output

Screenshots saved as `screenshot_001.png`, `screenshot_002.png`, etc.

Capture sequence:

1. Initial terminal state
2. Periodic snapshots (per `--interval`)
3. Final state after script completes
