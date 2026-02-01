package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/yarlson/scr/internal/capture"
	"github.com/yarlson/scr/internal/config"
	"github.com/yarlson/scr/internal/input"
)

// NewRootCommand creates and returns the root Cobra command for tui-capture.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui-capture",
		Short: "Capture TUI interactions and generate screenshots",
		RunE:  runCommand,
	}

	// Define required flags
	cmd.Flags().String("command", "", "Command to execute (required)")
	cmd.Flags().String("keypresses", "", "Comma-separated keypresses to send (required)")
	cmd.Flags().String("delays", "", "Comma-separated delays between keypresses (required)")

	// Define optional flags with defaults
	cmd.Flags().String("output-dir", "./screenshots", "Directory to save screenshots")
	cmd.Flags().Duration("screenshot-interval", 500*time.Millisecond, "Interval between screenshots")
	cmd.Flags().Int("ttyd-port", 7681, "Port for ttyd server")
	cmd.Flags().Duration("timeout", 60*time.Second, "Timeout for the entire operation")
	cmd.Flags().Bool("verbose", false, "Enable verbose logging")

	// Mark required flags
	if err := cmd.MarkFlagRequired("command"); err != nil {
		panic(fmt.Sprintf("failed to mark command flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("keypresses"); err != nil {
		panic(fmt.Sprintf("failed to mark keypresses flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("delays"); err != nil {
		panic(fmt.Sprintf("failed to mark delays flag as required: %v", err))
	}

	return cmd
}

// runCommand is the RunE function that handles flag parsing and validation.
func runCommand(cmd *cobra.Command, args []string) error {
	// Get flag values
	command, err := cmd.Flags().GetString("command")
	if err != nil {
		return fmt.Errorf("get command flag: %w", err)
	}

	keypressStr, err := cmd.Flags().GetString("keypresses")
	if err != nil {
		return fmt.Errorf("get keypresses flag: %w", err)
	}

	delaysStr, err := cmd.Flags().GetString("delays")
	if err != nil {
		return fmt.Errorf("get delays flag: %w", err)
	}

	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		return fmt.Errorf("get output-dir flag: %w", err)
	}

	screenshotInterval, err := cmd.Flags().GetDuration("screenshot-interval")
	if err != nil {
		return fmt.Errorf("get screenshot-interval flag: %w", err)
	}

	ttydPort, err := cmd.Flags().GetInt("ttyd-port")
	if err != nil {
		return fmt.Errorf("get ttyd-port flag: %w", err)
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return fmt.Errorf("get timeout flag: %w", err)
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("get verbose flag: %w", err)
	}

	// Parse keypresses string and validate keys
	parsedKeys, err := input.ParseKeypresses(keypressStr)
	if err != nil {
		return fmt.Errorf("parse keypresses: %w", err)
	}

	// Parse delays string
	var delays []time.Duration
	if delaysStr == "" {
		// For single keypress (N=1), no delays are needed (N-1=0)
		// For multiple keypresses, delays is required
		if len(parsedKeys) > 1 {
			return fmt.Errorf("delays flag is required for multiple keypresses (comma-separated list, e.g. '100ms,200ms')")
		}
		delays = []time.Duration{}
	} else {
		delayStrs := strings.Split(delaysStr, ",")
		delays = make([]time.Duration, len(delayStrs))
		for i, ds := range delayStrs {
			d, err := time.ParseDuration(strings.TrimSpace(ds))
			if err != nil {
				return fmt.Errorf("parse delay %q: %w", ds, err)
			}
			delays[i] = d
		}
	}

	// Create config
	cfg := &config.Config{
		Command:            command,
		Keypresses:         parsedKeys,
		Delays:             delays,
		OutputDir:          outputDir,
		ScreenshotInterval: screenshotInterval,
		TTydPort:           ttydPort,
		Timeout:            timeout,
		Verbose:            verbose,
	}

	// Validate entire config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// Log success if verbose
	if cfg.Verbose {
		logger := log.New(os.Stderr, "", log.LstdFlags)
		logger.Printf("Configuration validated successfully")
		logger.Printf("Command: %s", cfg.Command)
		logger.Printf("Keypresses: %v", cfg.Keypresses)
		logger.Printf("Output Directory: %s", cfg.OutputDir)
	}

	// Create capturer and execute capture workflow
	capturer := capture.NewCapturer(cfg)

	// Apply timeout from config
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Set up signal handling for graceful shutdown on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Run capture in a goroutine to allow signal handling
	runErr := make(chan error, 1)
	go func() {
		runErr <- capturer.Run(ctx)
	}()

	// Wait for either capture completion or signal
	select {
	case err := <-runErr:
		if err != nil {
			return fmt.Errorf("capture execution: %w", err)
		}
	case sig := <-sigChan:
		// Cancel context on signal to trigger cleanup
		cancel()
		// Wait for capture to finish cleanup
		if err := <-runErr; err != nil {
			log.Printf("shutdown error: %v", err)
		}
		fmt.Fprintf(os.Stderr, "\nReceived signal %s, shutting down gracefully...\n", sig)
		return fmt.Errorf("interrupted by signal: %s", sig)
	}

	// Print success message
	fmt.Printf("Capture completed successfully\n")

	return nil
}

// main is the entry point for the CLI application.
func main() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
