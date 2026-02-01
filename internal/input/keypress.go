package input

import (
	"fmt"
	"strings"
)

var specialKeyCodes = map[string]string{
	"enter":     "Enter",
	"escape":    "Escape",
	"tab":       "Tab",
	"space":     "Space",
	"up":        "ArrowUp",
	"down":      "ArrowDown",
	"left":      "ArrowLeft",
	"right":     "ArrowRight",
	"home":      "Home",
	"end":       "End",
	"pageup":    "PageUp",
	"pagedown":  "PageDown",
	"backspace": "Backspace",
	"delete":    "Delete",
	"ctrl+c":    "c",
	"ctrl+d":    "d",
}

// isSinglePrintableASCII reports whether key is exactly one printable ASCII character.
// This intentionally excludes non-ASCII and control characters.
func isSinglePrintableASCII(key string) bool {
	if len(key) != 1 {
		return false
	}
	ch := key[0]
	return ch >= 0x20 && ch <= 0x7e
}

// ParseKeypresses splits a comma-separated keypress string into individual keys.
// It trims whitespace from each key and returns the ordered slice.
// Returns error if input is an empty string.
func ParseKeypresses(keypressStr string) ([]string, error) {
	if keypressStr == "" {
		return nil, fmt.Errorf("keypress string must not be empty")
	}

	parts := strings.Split(keypressStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		result = append(result, trimmed)
	}

	// Validate that all parsed keys are actually valid
	for _, key := range result {
		if !IsValidKey(key) {
			return nil, fmt.Errorf("invalid key: %q", key)
		}
	}

	return result, nil
}

// IsValidKey checks if a key is valid (either alphanumeric or a recognized special key).
// Check is case-insensitive for special keys.
func IsValidKey(key string) bool {
	if isSinglePrintableASCII(key) {
		return true
	}
	_, ok := specialKeyCodes[strings.ToLower(key)]
	return ok
}

// KeyToKeyCode maps a key name to its Chrome DevTools Protocol key code.
// Single character keys are returned as-is.
// Special keys are mapped to their CDP codes.
// Ctrl+C and Ctrl+D return "c" and "d" respectively.
// Returns error if the key is not recognized.
func KeyToKeyCode(key string) (string, error) {
	// Single character keys are sent directly, except space which is a named key in CDP.
	if isSinglePrintableASCII(key) {
		if key == " " {
			return "Space", nil
		}
		return key, nil
	}

	lowerKey := strings.ToLower(key)
	if code, exists := specialKeyCodes[lowerKey]; exists {
		return code, nil
	}

	return "", fmt.Errorf("key %q is not recognized", key)
}
