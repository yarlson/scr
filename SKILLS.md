# scr

Capture TUI screenshots with scripted interactions.

```
scr [flags] <command> [script]
```

## Flags

| Flag | Default         | Description         |
| ---- | --------------- | ------------------- |
| `-o` | `./screenshots` | Output directory    |
| `-i` | `500ms`         | Screenshot interval |
| `-t` | `60s`           | Timeout             |
| `-p` | `7681`          | ttyd port           |
| `-v` | `false`         | Verbose             |

## Script Actions

| Command    | Syntax             | Description                     |
| ---------- | ------------------ | ------------------------------- |
| Type       | `Type 'text'`      | Type text (50ms/char default)   |
| Type speed | `Type@30ms 'text'` | Type with custom per-char delay |
| Sleep      | `Sleep 500ms`      | Pause for duration (ms or s)    |
| Key        | `Enter`            | Press key once                  |
| Key repeat | `Down 3`           | Press key N times               |
| Key delay  | `Enter@200ms`      | Delay before keypress           |
| Ctrl combo | `Ctrl+C`           | Send control character          |

## Supported Keys

`Enter`, `Tab`, `Escape`, `Space`, `Backspace`, `Delete`, `Up`, `Down`, `Left`, `Right`, `Home`, `End`, `PageUp`, `PageDown`

## Usage Examples

### Basic static capture

```bash
scr "ls -la"
```

### Type and execute command

```bash
scr bash "Type 'echo hello world' Enter"
```

### Interactive selection with fzf

```bash
scr "seq 1 20 | fzf" "Down 5 Enter"
```

### Complex script with timing

```bash
scr bash "Sleep 1s Type 'date' Enter Sleep 500ms Type 'whoami' Enter"
```

### Custom typing speed and Ctrl combo

```bash
scr bash "Type@20ms 'fast input' Enter Ctrl+C"
```

## Error Patterns

Parse errors include position info:

```
Error: parse script: position 23: expected quoted string after Type
```

Invalid key error:

```
Error: unknown key 'Foo' at position 15; valid keys: Enter, Tab, ...
```
