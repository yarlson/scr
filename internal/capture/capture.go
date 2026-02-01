package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"

	"github.com/yarlson/scr/internal/config"
	inputpkg "github.com/yarlson/scr/internal/input"
	"github.com/yarlson/scr/internal/script"
)

// Capturer orchestrates the TUI interaction workflow, connecting ttyd,
// Chrome DevTools Protocol, and screenshot management.
type Capturer struct {
	config          *config.Config
	ttyd            *TTydServer
	screenshotCount int
	mu              sync.Mutex
}

// NewCapturer creates and returns a new Capturer with the provided config.
// It initializes ttyd with cfg.Command and cfg.TTydPort.
// It does NOT start ttyd yet (that happens in Run()).
// It does NOT validate config (caller has already done so).
func NewCapturer(cfg *config.Config) *Capturer {
	if cfg == nil {
		panic("NewCapturer: config must not be nil")
	}
	return &Capturer{
		config:          cfg,
		ttyd:            NewTTydServer(cfg.Command, cfg.TTydPort),
		screenshotCount: 0,
	}
}

// Validate checks that the Capturer configuration is valid.
// It checks that config is not nil.
func (c *Capturer) Validate() error {
	if c.config == nil {
		return fmt.Errorf("config must not be nil")
	}
	return nil
}

// Run orchestrates the TUI capture workflow:
// 1. Creates output directory
// 2. Starts ttyd process
// 3. Launches Chrome browser
// 4. Navigates to ttyd URL
// 5. Captures initial screenshot
// 6. Sends keypresses with configured delays
// 7. Captures screenshots at specified intervals
// 8. Captures final screenshot
// All cleanup defers execute even on error.
func (c *Capturer) Run(ctx context.Context) error {
	// Create output directory
	if err := os.MkdirAll(c.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("output directory: %w", err)
	}

	// Start ttyd process
	if err := c.ttyd.Start(ctx); err != nil {
		return fmt.Errorf("start ttyd: %w", err)
	}
	defer c.ttyd.Stop()

	// Launch Chrome browser
	browserCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	// chromedp.Cancel() explicitly terminates the Chrome process,
	// distinct from context cancel which only closes the connection
	defer chromedp.Cancel(browserCtx)

	// Navigate to ttyd URL
	if err := chromedp.Run(browserCtx, chromedp.Navigate(c.ttyd.URL())); err != nil {
		return fmt.Errorf("navigate to ttyd: %w", err)
	}

	// Set viewport size for consistent screenshots
	if err := chromedp.Run(browserCtx, chromedp.EmulateViewport(1280, 720)); err != nil {
		return fmt.Errorf("set viewport: %w", err)
	}

	// Wait for xterm terminal to be ready
	if err := chromedp.Run(browserCtx,
		chromedp.WaitVisible(".xterm-screen", chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("wait for terminal: %w", err)
	}

	// Capture initial screenshot at t=0
	if c.config.Verbose {
		fmt.Fprintf(os.Stderr, "Capturing initial screenshot\n")
	}
	if err := c.captureScreenshot(browserCtx, c.getScreenshotFilename()); err != nil {
		return fmt.Errorf("initial screenshot: %w", err)
	}

	// Start interval-based screenshot capture
	var intervalStopChan chan struct{}
	var wg sync.WaitGroup
	if c.config.ScreenshotInterval > 0 {
		intervalStopChan = make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.captureIntervalScreenshots(browserCtx, intervalStopChan)
		}()
	}

	// Execute actions directly
	if err := c.executeActions(ctx, browserCtx, intervalStopChan, &wg); err != nil {
		return err
	}

	// Stop interval-based screenshots
	if intervalStopChan != nil {
		close(intervalStopChan)
		wg.Wait()
	}

	// Small delay to ensure final state is rendered
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// continue
	}

	// Capture final screenshot
	if c.config.Verbose {
		fmt.Fprintf(os.Stderr, "Capturing final screenshot\n")
	}
	if err := c.captureScreenshot(browserCtx, c.getScreenshotFilename()); err != nil {
		return fmt.Errorf("final screenshot: %w", err)
	}

	return nil
}

// sendKeypress sends a keypress to the browser using CDP Input.dispatchKeyEvent.
// It handles both regular keys and special keys (including Ctrl combinations).
func (c *Capturer) sendKeypress(ctx context.Context, key string) error {
	// Check if this is a Ctrl key combination
	if c.isCtrlKey(key) {
		return c.sendCtrlKeypress(ctx, key)
	}

	keyCode, err := inputpkg.KeyToKeyCode(key)
	if err != nil {
		return fmt.Errorf("lookup key code for %q: %w", key, err)
	}

	// For regular keys, use chromedp.KeyEvent
	return chromedp.Run(ctx, chromedp.KeyEvent(keyCode))
}

// sendCtrlKeypress sends a Ctrl+key combination using chromedp.KeyEvent with modifiers.
func (c *Capturer) sendCtrlKeypress(ctx context.Context, key string) error {
	lowerKey := strings.ToLower(key)
	if !strings.HasPrefix(lowerKey, "ctrl+") || len(lowerKey) != 6 {
		return fmt.Errorf("invalid Ctrl key format: %s", key)
	}
	keyChar := lowerKey[5:] // Extract the character after "ctrl+"

	// Use chromedp.KeyEvent with Ctrl modifier
	return chromedp.Run(ctx, chromedp.KeyEvent(keyChar, chromedp.KeyModifiers(input.ModifierCtrl)))
}

// isCtrlKey checks if a key is a Ctrl key combination.
func (c *Capturer) isCtrlKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return strings.HasPrefix(lowerKey, "ctrl+") && len(lowerKey) == 6
}

// executeActions executes the configured actions in sequence.
// It handles ActionType, ActionSleep, ActionKey, and ActionCtrl.
// All blocking operations respect ctx.Done() for graceful shutdown.
func (c *Capturer) executeActions(ctx, browserCtx context.Context, intervalStopChan chan struct{}, wg *sync.WaitGroup) error {
	// Determine which action set to use
	actions := c.config.Actions
	useActions := len(actions) > 0

	if !useActions {
		// Fall back to legacy keypresses/delays for backward compatibility
		return c.executeKeypresses(ctx, browserCtx, intervalStopChan, wg)
	}

	for i, action := range actions {
		if err := c.executeSingleAction(ctx, browserCtx, action, i, intervalStopChan, wg); err != nil {
			return err
		}
	}

	return nil
}

// executeSingleAction executes a single action based on its kind.
func (c *Capturer) executeSingleAction(ctx, browserCtx context.Context, action script.Action, index int, intervalStopChan chan struct{}, wg *sync.WaitGroup) error {
	switch action.Kind {
	case script.ActionType:
		return c.executeTypeAction(ctx, browserCtx, action, index, intervalStopChan, wg)
	case script.ActionSleep:
		return c.executeSleepAction(ctx, action, index)
	case script.ActionKey:
		return c.executeKeyAction(ctx, browserCtx, action, index, intervalStopChan, wg)
	case script.ActionCtrl:
		return c.executeCtrlAction(ctx, browserCtx, action, index, intervalStopChan, wg)
	default:
		return fmt.Errorf("unknown action kind: %v", action.Kind)
	}
}

// executeTypeAction executes a type action by sending each character with per-char delay.
func (c *Capturer) executeTypeAction(ctx, browserCtx context.Context, action script.Action, index int, intervalStopChan chan struct{}, wg *sync.WaitGroup) error {
	for _, char := range action.Text {
		// Check for context cancellation before each character
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if c.config.Verbose {
			fmt.Fprintf(os.Stderr, "Sending character: %s\n", string(char))
		}

		if err := c.sendKeypress(browserCtx, string(char)); err != nil {
			if intervalStopChan != nil {
				close(intervalStopChan)
				wg.Wait()
			}
			return fmt.Errorf("send character %q: %w", char, err)
		}

		// Sleep for per-character speed
		if action.Speed > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(action.Speed):
				// continue
			}
		}
	}

	// Apply post-action delay if specified
	if action.Delay > 0 {
		if c.config.Verbose {
			fmt.Fprintf(os.Stderr, "Waiting %v after type action %d\n", action.Delay, index)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(action.Delay):
			// continue
		}
	}

	return nil
}

// executeSleepAction executes a sleep action with context-aware cancellation.
func (c *Capturer) executeSleepAction(ctx context.Context, action script.Action, index int) error {
	if c.config.Verbose {
		fmt.Fprintf(os.Stderr, "Sleeping for %v (action %d)\n", action.Duration, index)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(action.Duration):
		// continue
	}

	return nil
}

// executeKeyAction executes a key action with optional delay and repeat count.
func (c *Capturer) executeKeyAction(ctx, browserCtx context.Context, action script.Action, index int, intervalStopChan chan struct{}, wg *sync.WaitGroup) error {
	// Determine repeat count (defaults to 1)
	repeat := action.Repeat
	if repeat <= 0 {
		repeat = 1
	}

	for i := 0; i < repeat; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if c.config.Verbose {
			fmt.Fprintf(os.Stderr, "Sending keypress: %s (repeat %d/%d)\n", action.Key, i+1, repeat)
		}

		if err := c.sendKeypress(browserCtx, action.Key); err != nil {
			if intervalStopChan != nil {
				close(intervalStopChan)
				wg.Wait()
			}
			return fmt.Errorf("send key %q (repeat %d): %w", action.Key, i+1, err)
		}
	}

	// Apply post-action delay if specified
	if action.Delay > 0 {
		if c.config.Verbose {
			fmt.Fprintf(os.Stderr, "Waiting %v after key action %d\n", action.Delay, index)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(action.Delay):
			// continue
		}
	}

	return nil
}

// executeCtrlAction executes a control key combination action.
func (c *Capturer) executeCtrlAction(ctx, browserCtx context.Context, action script.Action, index int, intervalStopChan chan struct{}, wg *sync.WaitGroup) error {
	if c.config.Verbose {
		fmt.Fprintf(os.Stderr, "Sending Ctrl+%s (action %d)\n", action.Key, index)
	}

	if err := c.sendCtrlKeypress(browserCtx, "ctrl+"+action.Key); err != nil {
		if intervalStopChan != nil {
			close(intervalStopChan)
			wg.Wait()
		}
		return fmt.Errorf("send Ctrl+%s: %w", action.Key, err)
	}

	return nil
}

// executeKeypresses executes the legacy keypresses/delays configuration.
func (c *Capturer) executeKeypresses(ctx, browserCtx context.Context, intervalStopChan chan struct{}, wg *sync.WaitGroup) error {
	for i, key := range c.config.Keypresses {
		// Wait for delay before sending key (except for first key)
		if i > 0 && i-1 < len(c.config.Delays) {
			delay := c.config.Delays[i-1]
			if c.config.Verbose {
				fmt.Fprintf(os.Stderr, "Waiting %v before sending keypress %d\n", delay, i)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// continue
			}
		}

		if c.config.Verbose {
			fmt.Fprintf(os.Stderr, "Sending keypress: %s\n", key)
		}

		if err := c.sendKeypress(browserCtx, key); err != nil {
			if intervalStopChan != nil {
				close(intervalStopChan)
				wg.Wait()
			}
			return fmt.Errorf("send keypress %d (%s): %w", i, key, err)
		}
	}

	return nil
}

// getScreenshotFilename returns the filename for the next screenshot
// with sequential naming (screenshot_001.png, etc.).
func (c *Capturer) getScreenshotFilename() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.screenshotCount++
	return filepath.Join(c.config.OutputDir, fmt.Sprintf("screenshot_%03d.png", c.screenshotCount))
}

// captureIntervalScreenshots captures screenshots at the configured interval
// until the stop channel is closed.
func (c *Capturer) captureIntervalScreenshots(ctx context.Context, stopChan chan struct{}) {
	if c.config.ScreenshotInterval <= 0 {
		return
	}

	ticker := time.NewTicker(c.config.ScreenshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			if c.config.Verbose {
				fmt.Fprintf(os.Stderr, "Capturing interval screenshot %d\n", c.screenshotCount+1)
			}
			// We need to create a new context for each screenshot since the
			// parent context might be cancelled
			if err := c.captureScreenshot(ctx, c.getScreenshotFilename()); err != nil {
				// Log error but don't stop - interval screenshots are best effort
				if c.config.Verbose {
					fmt.Fprintf(os.Stderr, "Failed to capture interval screenshot: %v\n", err)
				}
			}
		}
	}
}

// captureScreenshot captures the terminal and saves it as PNG.
// Returns an error if the capture or save fails.
func (c *Capturer) captureScreenshot(ctx context.Context, filename string) error {
	var buf []byte

	// Capture the terminal container element
	err := chromedp.Run(ctx,
		chromedp.Screenshot("#terminal-container", &buf, chromedp.NodeVisible, chromedp.ByID),
	)
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}

	// Write the captured screenshot data to file
	err = os.WriteFile(filename, buf, 0o644)
	if err != nil {
		return fmt.Errorf("write screenshot: %w", err)
	}

	return nil
}
