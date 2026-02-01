package input

import (
	"fmt"
	"strings"
)

// specialKeys maps recognized special key names to a boolean indicator.
var specialKeys = map[string]bool{
	"enter":     true,
	"escape":    true,
	"tab":       true,
	"up":        true,
	"down":      true,
	"left":      true,
	"right":     true,
	"home":      true,
	"end":       true,
	"backspace": true,
	"delete":    true,
	"ctrl+c":    true,
	"ctrl+d":    true,
}

// isAlphanumeric checks if a key is a single alphanumeric character.
func isAlphanumeric(key string) bool {
	if len(key) != 1 {
		return false
	}
	char := key[0]
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
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
	if isAlphanumeric(key) {
		return true
	}

	return specialKeys[strings.ToLower(key)]
}

// KeyToKeyCode maps a key name to its Chrome DevTools Protocol key code.
// Single character keys are returned as-is.
// Special keys are mapped to their CDP codes.
// Ctrl+C and Ctrl+D return "c" and "d" respectively.
// Returns error if the key is not recognized.
func KeyToKeyCode(key string) (string, error) {
	if isAlphanumeric(key) {
		return key, nil
	}

	// Special keys mapping (case-insensitive)
	specialKeyMap := map[string]string{
		"enter":     "Enter",
		"escape":    "Escape",
		"tab":       "Tab",
		"up":        "ArrowUp",
		"down":      "ArrowDown",
		"left":      "ArrowLeft",
		"right":     "ArrowRight",
		"home":      "Home",
		"end":       "End",
		"backspace": "Backspace",
		"delete":    "Delete",
		"ctrl+c":    "c",
		"ctrl+d":    "d",
	}

	lowerKey := strings.ToLower(key)
	if code, exists := specialKeyMap[lowerKey]; exists {
		return code, nil
	}

	return "", fmt.Errorf("key %q is not recognized", key)
}
