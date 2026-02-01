package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// Config holds the configuration for the TUI screen capture application.
type Config struct {
	Command            string
	Keypresses         []string
	Delays             []time.Duration
	OutputDir          string
	ScreenshotInterval time.Duration
	TTydPort           int
	Timeout            time.Duration
	Verbose            bool
}

// ParseConfig extracts configuration from Cobra command flags.
func ParseConfig(cmd *cobra.Command) (*Config, error) {
	if cmd == nil {
		return nil, fmt.Errorf("command must not be nil")
	}

	command, err := cmd.Flags().GetString("command")
	if err != nil {
		return nil, fmt.Errorf("get command flag: %w", err)
	}

	keypresses, err := cmd.Flags().GetStringSlice("keypresses")
	if err != nil {
		return nil, fmt.Errorf("get keypresses flag: %w", err)
	}

	delays, err := cmd.Flags().GetDurationSlice("delays")
	if err != nil {
		return nil, fmt.Errorf("get delays flag: %w", err)
	}

	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		return nil, fmt.Errorf("get output-dir flag: %w", err)
	}

	screenshotInterval, err := cmd.Flags().GetDuration("screenshot-interval")
	if err != nil {
		return nil, fmt.Errorf("get screenshot-interval flag: %w", err)
	}

	ttydPort, err := cmd.Flags().GetInt("ttyd-port")
	if err != nil {
		return nil, fmt.Errorf("get ttyd-port flag: %w", err)
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return nil, fmt.Errorf("get timeout flag: %w", err)
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, fmt.Errorf("get verbose flag: %w", err)
	}

	return &Config{
		Command:            command,
		Keypresses:         keypresses,
		Delays:             delays,
		OutputDir:          outputDir,
		ScreenshotInterval: screenshotInterval,
		TTydPort:           ttydPort,
		Timeout:            timeout,
		Verbose:            verbose,
	}, nil
}

// Validate checks that all configuration fields are valid.
func (c *Config) Validate() error {
	if c.Command == "" {
		return fmt.Errorf("command must be non-empty")
	}

	if c.OutputDir == "" {
		return fmt.Errorf("output-dir must be non-empty")
	}

	if c.TTydPort < 1 || c.TTydPort > 65535 {
		return fmt.Errorf("ttyd-port must be between 1 and 65535")
	}

	if c.ScreenshotInterval <= 0 {
		return fmt.Errorf("screenshot-interval must be > 0")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be > 0")
	}

	if len(c.Keypresses) == 0 {
		return fmt.Errorf("keypresses must not be empty")
	}

	if len(c.Delays) != len(c.Keypresses)-1 {
		return fmt.Errorf("delays length must be equal to keypresses length - 1")
	}

	return nil
}
