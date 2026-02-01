package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKeypresses_ValidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "simple sequence",
			input:   "h,e,l,l,o,Enter",
			want:    []string{"h", "e", "l", "l", "o", "Enter"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeypresses(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseKeypresses_WithWhitespace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "whitespace around keys",
			input:   "h, e, l, l, o",
			want:    []string{"h", "e", "l", "l", "o"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeypresses(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseKeypresses_EmptyString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseKeypresses(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidKey_AlphanumericAndSpecial(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "single letter",
			key:   "a",
			valid: true,
		},
		{
			name:  "single digit",
			key:   "5",
			valid: true,
		},
		{
			name:  "Enter key",
			key:   "Enter",
			valid: true,
		},
		{
			name:  "Escape key",
			key:   "Escape",
			valid: true,
		},
		{
			name:  "Tab key",
			key:   "Tab",
			valid: true,
		},
		{
			name:  "Up arrow",
			key:   "Up",
			valid: true,
		},
		{
			name:  "Down arrow",
			key:   "Down",
			valid: true,
		},
		{
			name:  "Left arrow",
			key:   "Left",
			valid: true,
		},
		{
			name:  "Right arrow",
			key:   "Right",
			valid: true,
		},
		{
			name:  "Home key",
			key:   "Home",
			valid: true,
		},
		{
			name:  "End key",
			key:   "End",
			valid: true,
		},
		{
			name:  "Backspace key",
			key:   "Backspace",
			valid: true,
		},
		{
			name:  "Delete key",
			key:   "Delete",
			valid: true,
		},
		{
			name:  "Ctrl+C",
			key:   "Ctrl+C",
			valid: true,
		},
		{
			name:  "Ctrl+D",
			key:   "Ctrl+D",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidKey(tt.key)
			assert.Equal(t, tt.valid, got)
		})
	}
}

func TestIsValidKey_InvalidKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "invalid key",
			key:   "NotAKey",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidKey(tt.key)
			assert.Equal(t, tt.valid, got)
		})
	}
}

func TestIsValidKey_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "lowercase enter",
			key:   "enter",
			valid: true,
		},
		{
			name:  "uppercase ENTER",
			key:   "ENTER",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidKey(tt.key)
			assert.Equal(t, tt.valid, got)
		})
	}
}

func TestKeyToKeyCode_SingleChars(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "single letter a",
			key:     "a",
			want:    "a",
			wantErr: false,
		},
		{
			name:    "single digit 5",
			key:     "5",
			want:    "5",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyToKeyCode(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestKeyToKeyCode_SpecialKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "Enter key",
			key:     "Enter",
			want:    "Enter",
			wantErr: false,
		},
		{
			name:    "Escape key",
			key:     "Escape",
			want:    "Escape",
			wantErr: false,
		},
		{
			name:    "Tab key",
			key:     "Tab",
			want:    "Tab",
			wantErr: false,
		},
		{
			name:    "Up arrow",
			key:     "Up",
			want:    "ArrowUp",
			wantErr: false,
		},
		{
			name:    "Down arrow",
			key:     "Down",
			want:    "ArrowDown",
			wantErr: false,
		},
		{
			name:    "Left arrow",
			key:     "Left",
			want:    "ArrowLeft",
			wantErr: false,
		},
		{
			name:    "Right arrow",
			key:     "Right",
			want:    "ArrowRight",
			wantErr: false,
		},
		{
			name:    "Home key",
			key:     "Home",
			want:    "Home",
			wantErr: false,
		},
		{
			name:    "End key",
			key:     "End",
			want:    "End",
			wantErr: false,
		},
		{
			name:    "Backspace key",
			key:     "Backspace",
			want:    "Backspace",
			wantErr: false,
		},
		{
			name:    "Delete key",
			key:     "Delete",
			want:    "Delete",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyToKeyCode(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestKeyToKeyCode_CtrlKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "Ctrl+C",
			key:     "Ctrl+C",
			want:    "c",
			wantErr: false,
		},
		{
			name:    "Ctrl+D",
			key:     "Ctrl+D",
			want:    "d",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyToKeyCode(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestKeyToKeyCode_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "lowercase enter",
			key:     "enter",
			want:    "Enter",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyToKeyCode(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestKeyToKeyCode_InvalidKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid key",
			key:     "NotAKey",
			wantErr: true,
			errMsg:  "", // any error is fine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := KeyToKeyCode(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
