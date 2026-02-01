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
	"github.com/yarlson/scr/internal/script"
)

// NewRootCommand creates and returns the root Cobra command for scr.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scr [flags] COMMAND [SCRIPT]",
		Short: "Capture TUI interactions and generate screenshots",
		Long: `scr captures screenshots of terminal interactions.

Usage:
  scr [flags] COMMAND
  scr [flags] COMMAND SCRIPT

Examples:
  # Static capture - just run a command
  scr bash

  # With tape script
  scr bash "Type 'echo hello' Enter"`,
		Args: cobra.RangeArgs(0, 2),
		RunE: runCommand,
	}

	// New short flags
	cmd.Flags().StringP("out", "o", "./screenshots", "Directory to save screenshots")
	cmd.Flags().DurationP("interval", "i", 500*time.Millisecond, "Interval between screenshots")
	cmd.Flags().DurationP("timeout", "t", 60*time.Second, "Timeout for the entire operation")
	cmd.Flags().IntP("port", "p", 7681, "Port for ttyd server")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")

	// Hidden deprecated flags (for backward compatibility)
	cmd.Flags().String("command", "", "Command to execute (deprecated: use positional arg)")
	cmd.Flags().String("keypresses", "", "Comma-separated keypresses (deprecated: use SCRIPT arg)")
	cmd.Flags().String("delays", "", "Comma-separated delays (deprecated: use SCRIPT arg)")
	cmd.Flags().String("output-dir", "", "Directory to save screenshots (deprecated: use -o)")
	cmd.Flags().Duration("screenshot-interval", 0, "Interval between screenshots (deprecated: use -i)")
	cmd.Flags().Int("ttyd-port", 0, "Port for ttyd server (deprecated: use -p)")

	// Mark deprecated flags as hidden and deprecated
	_ = cmd.Flags().MarkHidden("command")
	_ = cmd.Flags().MarkHidden("keypresses")
	_ = cmd.Flags().MarkHidden("delays")
	_ = cmd.Flags().MarkHidden("output-dir")
	_ = cmd.Flags().MarkHidden("screenshot-interval")
	_ = cmd.Flags().MarkHidden("ttyd-port")

	_ = cmd.Flags().MarkDeprecated("command", "use the COMMAND positional argument instead")
	_ = cmd.Flags().MarkDeprecated("keypresses", "use the SCRIPT positional argument instead")
	_ = cmd.Flags().MarkDeprecated("delays", "use the SCRIPT positional argument instead")
	_ = cmd.Flags().MarkDeprecated("output-dir", "use -o or --out instead")
	_ = cmd.Flags().MarkDeprecated("screenshot-interval", "use -i or --interval instead")
	_ = cmd.Flags().MarkDeprecated("ttyd-port", "use -p or --port instead")

	return cmd
}

// runCommand is the RunE function that handles flag parsing and validation.
func runCommand(cmd *cobra.Command, args []string) error {
	// Check for deprecated flag usage
	deprecatedFlagsUsed := cmd.Flags().Changed("command") || cmd.Flags().Changed("keypresses") || cmd.Flags().Changed("delays")

	// Determine command source
	var command string
	if len(args) > 0 {
		command = args[0]
	}

	// Check for mixing positional args with deprecated flags
	if len(args) > 0 && deprecatedFlagsUsed {
		return fmt.Errorf("cannot use both positional arguments and deprecated flags: use either 'scr COMMAND [SCRIPT]' or deprecated flags, not both")
	}

	// Get optional script from args
	var scriptStr string
	if len(args) > 1 {
		scriptStr = args[1]
	}

	// Handle deprecated flag mode
	if deprecatedFlagsUsed {
		return runWithDeprecatedFlags(cmd)
	}

	// Handle new positional arg mode
	return runWithPositionalArgs(cmd, command, scriptStr)
}

// runWithPositionalArgs handles the new positional argument interface.
func runWithPositionalArgs(cmd *cobra.Command, command, scriptStr string) error {
	if command == "" {
		return fmt.Errorf("COMMAND is required (e.g., 'scr bash' or 'scr bash \"Type ...\"')")
	}

	// Parse new short flags
	outputDir, err := cmd.Flags().GetString("out")
	if err != nil {
		return fmt.Errorf("get out flag: %w", err)
	}

	screenshotInterval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return fmt.Errorf("get interval flag: %w", err)
	}

	ttydPort, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("get port flag: %w", err)
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return fmt.Errorf("get timeout flag: %w", err)
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("get verbose flag: %w", err)
	}

	// Parse script if provided
	var actions []script.Action
	if scriptStr != "" {
		parsedActions, err := script.Parse(scriptStr)
		if err != nil {
			return fmt.Errorf("parse script: %w", err)
		}
		actions = parsedActions
	}

	// Create config - pass actions directly to capture engine
	cfg := &config.Config{
		Command:            command,
		OutputDir:          outputDir,
		ScreenshotInterval: screenshotInterval,
		TTydPort:           ttydPort,
		Timeout:            timeout,
		Verbose:            verbose,
		Actions:            actions,
		Script:             scriptStr,
	}

	// Validate config (skip keypresses/delays validation if script was used)
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// Log success if verbose
	if cfg.Verbose {
		logger := log.New(os.Stderr, "", log.LstdFlags)
		logger.Printf("Configuration validated successfully")
		logger.Printf("Command: %s", cfg.Command)
		if cfg.Script != "" {
			logger.Printf("Script: %s", cfg.Script)
			logger.Printf("Actions: %d", len(cfg.Actions))
		} else {
			logger.Printf("Keypresses: %v", cfg.Keypresses)
		}
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

// runWithDeprecatedFlags handles the old flag-based interface for backward compatibility.
func runWithDeprecatedFlags(cmd *cobra.Command) error {
	// Get deprecated flag values
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

	// Get new flags (with defaults)
	outputDir, err := cmd.Flags().GetString("out")
	if err != nil {
		return fmt.Errorf("get out flag: %w", err)
	}

	screenshotInterval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return fmt.Errorf("get interval flag: %w", err)
	}

	ttydPort, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("get port flag: %w", err)
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

	// Print deprecation warning
	fmt.Fprintf(os.Stderr, "Warning: Using deprecated flags. Please migrate to: scr [flags] COMMAND [SCRIPT]\n")

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
