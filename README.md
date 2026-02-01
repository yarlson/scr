# TUI Screen Capture CLI

## Overview

TUI Screen Capture CLI is a Go-based command-line tool that automates interactive terminal user interface (TUI) testing and documentation generation. It runs a specified CLI command through ttyd (a terminal sharing daemon) in headless Chrome, sends a sequence of keyboard inputs with precise timing delays, and captures PNG screenshots of the terminal output at specified intervals.

This tool enables reproducible, scriptable automation of TUI interactions for testing interactive command-line tools, generating documentation screenshots, creating tutorials, and building demos. By using ttyd with xterm.js running in headless Chrome, it provides consistent, realistic terminal rendering across different platforms without the quirks of terminal-specific emulators.

## Installation

### Prerequisites

- **Go 1.21+**: Required to build the tool from source
- **ttyd**: Terminal sharing daemon (install via package manager)
- **Chrome or Chromium**: Required for headless browser automation

### Install ttyd

**macOS:**

```bash
brew install ttyd
```

**Ubuntu/Debian:**

```bash
apt-get install ttyd
```

**Fedora/RHEL:**

```bash
dnf install ttyd
```

For other systems, see [ttyd releases](https://github.com/tsl0922/ttyd/releases).

### Install Chrome/Chromium

Chrome or Chromium is typically pre-installed on most systems. If not:

**macOS:**

```bash
brew install --cask google-chrome
```

**Ubuntu/Debian:**

```bash
apt-get install chromium-browser
```

The tool will automatically detect Chrome in common locations. You can also set the `CHROME_BIN` environment variable to specify a custom path:

```bash
export CHROME_BIN=/usr/bin/chromium
```

### Install tui-capture

```bash
go install github.com/your-org/tui-capture@latest  # TODO: Update after repo move
```

Or clone and build manually:

```bash
git clone https://github.com/your-org/tui-capture.git
cd tui-capture
go build -o tui-capture ./cmd
```

## Quick Start

Capture a simple bash session with a few keypresses:

```bash
tui-capture \
  --command "bash" \
  --keypresses "echo,hello,world,Enter" \
  --delays "100ms,100ms,100ms" \
  --output-dir "./demo"
```

**Expected output:**

```
Capture completed successfully
```

This will create 4+ screenshots in `./demo/` showing:

1. Initial empty bash prompt
2. After typing "echo"
3. After typing "hello"
4. After typing "world"
5. After pressing Enter (showing output "hello world")

## Usage

```bash
tui-capture --command <cmd> --keypresses <keys> --delays <delays> [flags]
```

### Required Flags

| Flag           | Description                                                                                                                                        | Example                                                              |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| `--command`    | The TUI command to execute. Supports commands with arguments.                                                                                      | `--command "fzf"` or `--command "bash -c 'echo test'"`               |
| `--keypresses` | Comma-separated sequence of key inputs. See [Supported Keys](#supported-keys).                                                                     | `--keypresses "h,e,l,l,o,Enter"` or `--keypresses "Down,Down,Enter"` |
| `--delays`     | Comma-separated delays between keypresses. Format: `100ms`, `1s`. Must have exactly one fewer delay than keypresses (delays occur _between_ keys). | `--delays "100ms,100ms,100ms,100ms"` (4 delays for 5 keys)           |

### Optional Flags

| Flag                    | Default         | Description                                                                                                                       |
| ----------------------- | --------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `--output-dir`          | `./screenshots` | Directory to save PNG screenshots. Created automatically if missing. Files named `screenshot_001.png`, `screenshot_002.png`, etc. |
| `--screenshot-interval` | `500ms`         | Capture a screenshot every N milliseconds during execution, plus final screenshot when complete.                                  |
| `--ttyd-port`           | `7681`          | Port for ttyd server to listen on. Change if port 7681 is already in use.                                                         |
| `--timeout`             | `60s`           | Maximum time to wait for entire operation before forcefully terminating.                                                          |
| `--verbose`             | `false`         | Enable verbose logging to stderr, showing detailed progress and debug information.                                                |

### Flag Examples

**Basic command with defaults:**

```bash
tui-capture --command "ls -la" --keypresses "" --delays ""
```

**Custom output directory and screenshot timing:**

```bash
tui-capture \
  --command "vim" \
  --keypresses "i,h,e,l,l,o,Escape,:,q,Enter" \
  --delays "200ms,50ms,50ms,50ms,50ms,100ms,100ms,100ms,100ms" \
  --output-dir "./vim_demo" \
  --screenshot-interval 300ms
```

**Non-default port (if 7681 is in use):**

```bash
tui-capture \
  --command "bash" \
  --keypresses "uptime,Enter" \
  --delays "100ms" \
  --ttyd-port 8080
```

**Verbose mode for debugging:**

```bash
tui-capture \
  --command "htop" \
  --keypresses "q" \
  --delays "500ms" \
  --verbose
```

## Examples

### Example 1: Bash Command Automation

Automate typing a command in bash and capturing the result:

```bash
tui-capture \
  --command "bash" \
  --keypresses "l,s, ,-,l,a,Enter" \
  --delays "50ms,50ms,50ms,50ms,50ms,50ms,50ms" \
  --output-dir "./bash_demo" \
  --screenshot-interval 200ms
```

This types `ls -la` in bash and captures the directory listing output.

### Example 2: fzf Interactive Navigation

Capture navigation through an fzf list:

```bash
tui-capture \
  --command "seq 1 100 | fzf" \
  --keypresses "Down,Down,Down,Enter" \
  --delays "200ms,200ms,200ms" \
  --output-dir "./fzf_demo" \
  --screenshot-interval 300ms
```

This creates a numbered list with `seq`, opens it in fzf, navigates down 3 items, and selects. Screenshots show the cursor movement.

### Example 3: Application Help Screen Capture

Capture a help screen with no interaction:

```bash
tui-capture \
  --command "git --help" \
  --keypresses "q" \
  --delays "1s" \
  --output-dir "./help_screen"
```

Captures the git help output. The `q` keypress exits the pager after 1 second.

### Example 4: Multi-step Interactive Workflow

Demonstrate a complex interactive session:

```bash
tui-capture \
  --command "python3" \
  --keypresses "p,r,i,n,t,(,h,e,l,l,o,),Enter,Ctrl+D" \
  --delays "50ms,50ms,50ms,50ms,50ms,50ms,50ms,50ms,50ms,50ms,50ms,100ms,500ms" \
  --output-dir "./python_demo" \
  --screenshot-interval 100ms \
  --verbose
```

This starts Python REPL, types `print(hello)`, executes it, and exits with Ctrl+D.

## Supported Keys

The `--keypresses` flag accepts comma-separated key names. Keys are case-insensitive for special keys.

### Alphanumeric Characters

- Any single letter: `a` through `z`, `A` through `Z`
- Any single digit: `0` through `9`
- Symbols: `-`, `=`, `,`, `.`, `/`, etc. (single character keys)

### Special Keys

| Key Name    | Description           | CDP Code                    |
| ----------- | --------------------- | --------------------------- |
| `Enter`     | Return/Enter key      | `Enter`                     |
| `Escape`    | Escape key            | `Escape`                    |
| `Tab`       | Tab key               | `Tab`                       |
| `Up`        | Up arrow              | `ArrowUp`                   |
| `Down`      | Down arrow            | `ArrowDown`                 |
| `Left`      | Left arrow            | `ArrowLeft`                 |
| `Right`     | Right arrow           | `ArrowRight`                |
| `Home`      | Home key              | `Home`                      |
| `End`       | End key               | `End`                       |
| `Backspace` | Backspace key         | `Backspace`                 |
| `Delete`    | Delete key            | `Delete`                    |
| `Ctrl+C`    | Control+C (interrupt) | `c` (with Control modifier) |
| `Ctrl+D`    | Control+D (EOF)       | `d` (with Control modifier) |

### Key Sequence Examples

Type "hello world" and press Enter:

```bash
--keypresses "h,e,l,l,o,Space,w,o,r,l,d,Enter"
```

Navigate a menu and select:

```bash
--keypresses "Down,Down,Down,Enter"
```

Exit a process with Ctrl+C:

```bash
--keypresses "Ctrl+C"
```

Exit a REPL with Ctrl+D:

```bash
--keypresses "Ctrl+D"
```

## Troubleshooting

### "ttyd binary not found"

**Error:** `ttyd binary not found. Install ttyd and ensure it's in PATH`

**Solution:** Install ttyd:

- macOS: `brew install ttyd`
- Linux: `apt-get install ttyd` or `dnf install ttyd`

Verify installation: `which ttyd` should return a path.

### "Headless Chrome not found"

**Error:** `Headless Chrome not found. Install Chromium or use CHROME_BIN env var`

**Solution:**

1. Install Chrome or Chromium
2. Set `CHROME_BIN` environment variable to the Chrome executable path:
   ```bash
   export CHROME_BIN=/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome
   # or
   export CHROME_BIN=/usr/bin/chromium
   ```

### "Port already in use"

**Error:** `ttyd failed to start: listen tcp :7681: bind: address already in use`

**Solution:** The default port 7681 is in use. Specify a different port:

```bash
tui-capture --command "bash" --keypresses "Enter" --delays "100ms" --ttyd-port 8080
```

### Screenshots are blank or terminal content not visible

**Possible causes and solutions:**

1. **Timing too fast**: Increase `--screenshot-interval` to give the terminal more time to render:

   ```bash
   --screenshot-interval 1000ms
   ```

2. **Command exits immediately**: Ensure the command stays running long enough. Use `bash` for interactive sessions or add delays before first screenshot.

3. **xterm.js not loaded**: Try increasing the default delay before first screenshot by adding an initial empty keypress with delay.

4. **Check verbose output**: Run with `--verbose` to see detailed progress and identify where the issue occurs.

### "delays length must be equal to keypresses length - 1"

**Error:** Validation error about mismatched delays and keypresses

**Solution:** You must have exactly one fewer delay than keypresses. For N keypresses, provide N-1 delays:

- 1 keypress: 0 delays (`--keypresses "q" --delays ""`)
- 2 keypresses: 1 delay (`--keypresses "a,b" --delays "100ms"`)
- 5 keypresses: 4 delays (`--keypresses "a,b,c,d,e" --delays "100ms,100ms,100ms,100ms"`)

### Timeout errors

**Error:** `context deadline exceeded` or timeout-related errors

**Solution:** Increase the `--timeout` value:

```bash
tui-capture --command "slow-app" --keypresses "Enter" --delays "100ms" --timeout 120s
```

### Orphaned processes (ttyd or Chrome still running)

**Issue:** After interruption, `ttyd` or `chrome` processes remain.

**Solution:** The tool catches SIGINT (Ctrl+C) and attempts cleanup. If processes remain:

```bash
# macOS/Linux
pkill -f ttyd
pkill -f "Google Chrome"

# Or more specifically
ps aux | grep ttyd
kill <pid>
```

### Empty or corrupted PNG files

**Issue:** Screenshot files are 0 bytes or cannot be opened.

**Solution:**

1. Check that the output directory is writable: `ls -la <output-dir>`
2. Ensure sufficient disk space
3. Run with `--verbose` to see if screenshot capture is failing silently
4. Try a simpler command first to isolate the issue

### "invalid key" errors

**Error:** `invalid key: "somekey"`

**Solution:** Only supported keys are allowed. Check the [Supported Keys](#supported-keys) section. Common mistakes:

- Using `Return` instead of `Enter`
- Using `Space` (not supported; use a literal space: ` ` or `, ,` in keypresses)
- Case sensitivity: use `Enter` not `ENTER`
- Modifier keys: Only `Ctrl+C` and `Ctrl+D` are supported currently
