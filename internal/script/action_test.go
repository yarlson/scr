package script

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestActionKind_Constants(t *testing.T) {
	// Verify iota-based constants
	assert.Equal(t, ActionKind(0), ActionType)
	assert.Equal(t, ActionKind(1), ActionSleep)
	assert.Equal(t, ActionKind(2), ActionKey)
	assert.Equal(t, ActionKind(3), ActionCtrl)
}

func TestAction_ZeroValues(t *testing.T) {
	// Verify zero values make sense
	action := Action{}
	assert.Equal(t, ActionKind(0), action.Kind)
	assert.Equal(t, "", action.Text)
	assert.Equal(t, "", action.Key)
	assert.Equal(t, time.Duration(0), action.Duration)
	assert.Equal(t, time.Duration(0), action.Speed)
	assert.Equal(t, time.Duration(0), action.Delay)
	assert.Equal(t, 0, action.Repeat) // Zero value is 0, caller should default to 1
}

func TestAction_FullValues(t *testing.T) {
	// Test that all fields can be set correctly
	action := Action{
		Kind:     ActionType,
		Text:     "hello world",
		Key:      "enter",
		Duration: 5 * time.Second,
		Speed:    100 * time.Millisecond,
		Delay:    500 * time.Millisecond,
		Repeat:   3,
	}

	assert.Equal(t, ActionType, action.Kind)
	assert.Equal(t, "hello world", action.Text)
	assert.Equal(t, "enter", action.Key)
	assert.Equal(t, 5*time.Second, action.Duration)
	assert.Equal(t, 100*time.Millisecond, action.Speed)
	assert.Equal(t, 500*time.Millisecond, action.Delay)
	assert.Equal(t, 3, action.Repeat)
}
