package script

import "time"

// ActionKind represents the type of action in a tape script.
type ActionKind int

const (
	// ActionType types text into the terminal.
	ActionType ActionKind = iota
	// ActionSleep pauses for a specified duration.
	ActionSleep
	// ActionKey presses a special key.
	ActionKey
	// ActionCtrl presses a control key combination.
	ActionCtrl
)

// Action represents a single action in a tape script.
type Action struct {
	// Kind is the type of action (Type, Sleep, Key, Ctrl).
	Kind ActionKind
	// Text is the text to type (for ActionType).
	Text string
	// Key is the key name (for ActionKey and ActionCtrl).
	Key string
	// Duration is the sleep duration (for ActionSleep).
	Duration time.Duration
	// Speed is the typing speed as a per-character delay (for ActionType).
	Speed time.Duration
	// Delay is the delay after typing this action (for ActionType, ActionKey, ActionCtrl).
	Delay time.Duration
	// Repeat is the number of times to repeat the key press (for ActionKey and ActionCtrl).
	// Defaults to 1.
	Repeat int
}
