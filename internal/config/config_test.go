package config

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a", "b", "c"},
		Delays:             []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           8080,
		Timeout:            30 * time.Second,
		Verbose:            false,
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_EmptyCommand(t *testing.T) {
	cfg := &Config{
		Command:            "",
		Keypresses:         []string{"a"},
		Delays:             []time.Duration{},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           8080,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command must be non-empty")
}

func TestValidate_EmptyOutputDir(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a"},
		Delays:             []time.Duration{},
		OutputDir:          "",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           8080,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output-dir must be non-empty")
}

func TestValidate_InvalidTTydPort_TooLow(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a"},
		Delays:             []time.Duration{},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           0,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ttyd-port must be between 1 and 65535")
}

func TestValidate_InvalidTTydPort_TooHigh(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a"},
		Delays:             []time.Duration{},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           65536,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ttyd-port must be between 1 and 65535")
}

func TestValidate_InvalidScreenshotInterval(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a"},
		Delays:             []time.Duration{},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 0,
		TTydPort:           8080,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "screenshot-interval must be > 0")
}

func TestValidate_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a"},
		Delays:             []time.Duration{},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           8080,
		Timeout:            0,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be > 0")
}

func TestValidate_EmptyKeypresses(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{},
		Delays:             []time.Duration{},
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           8080,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keypresses must not be empty")
}

func TestValidate_InvalidDelaysLength(t *testing.T) {
	cfg := &Config{
		Command:            "echo hello",
		Keypresses:         []string{"a", "b", "c"},
		Delays:             []time.Duration{100 * time.Millisecond}, // Should have 2 delays
		OutputDir:          "/tmp/output",
		ScreenshotInterval: 500 * time.Millisecond,
		TTydPort:           8080,
		Timeout:            30 * time.Second,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delays length must be equal to keypresses length - 1")
}

func TestParseConfig_Success(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("command", "echo hello", "")
	cmd.Flags().StringSlice("keypresses", []string{"a", "b"}, "")
	cmd.Flags().DurationSlice("delays", []time.Duration{100 * time.Millisecond}, "")
	cmd.Flags().String("out", "/tmp/output", "")
	cmd.Flags().Duration("interval", 500*time.Millisecond, "")
	cmd.Flags().Int("port", 8080, "")
	cmd.Flags().Duration("timeout", 30*time.Second, "")
	cmd.Flags().Bool("verbose", false, "")

	cfg, err := ParseConfig(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "echo hello", cfg.Command)
	assert.Equal(t, []string{"a", "b"}, cfg.Keypresses)
	assert.Equal(t, []time.Duration{100 * time.Millisecond}, cfg.Delays)
	assert.Equal(t, "/tmp/output", cfg.OutputDir)
	assert.Equal(t, 500*time.Millisecond, cfg.ScreenshotInterval)
	assert.Equal(t, 8080, cfg.TTydPort)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, false, cfg.Verbose)
}

func TestParseConfig_NilCommand(t *testing.T) {
	cfg, err := ParseConfig(nil)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "command must not be nil")
}
