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

	// Send keypresses with configured delays
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
			// Stop interval goroutine before returning error
			if intervalStopChan != nil {
				close(intervalStopChan)
				wg.Wait()
			}
			return fmt.Errorf("send keypress %d (%s): %w", i, key, err)
		}
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
	var keyChar string
	switch lowerKey {
	case "ctrl+c":
		keyChar = "c"
	case "ctrl+d":
		keyChar = "d"
	default:
		return fmt.Errorf("unsupported Ctrl key: %s", key)
	}

	// Use chromedp.KeyEvent with Ctrl modifier
	return chromedp.Run(ctx, chromedp.KeyEvent(keyChar, chromedp.KeyModifiers(input.ModifierCtrl)))
}

// isCtrlKey checks if a key is a Ctrl key combination.
func (c *Capturer) isCtrlKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return lowerKey == "ctrl+c" || lowerKey == "ctrl+d"
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

// captureScreenshot uses chromedp to capture a PNG screenshot and saves it
// to the specified filename.
// Returns an error if the capture or save fails.
func (c *Capturer) captureScreenshot(ctx context.Context, filename string) error {
	// Capture screenshot using chromedp to byte slice
	var buf []byte
	err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf))
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}

	// Write the captured screenshot data to file
	err = os.WriteFile(filename, buf, 0o644)
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}

	return nil
}
