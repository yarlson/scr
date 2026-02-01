package script

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Action
		wantErr string
	}{
		{
			name:  "simple type",
			input: "Type 'hello'",
			want:  []Action{{Kind: ActionType, Text: "hello", Speed: 50 * time.Millisecond}},
		},
		{
			name:  "type with double quotes",
			input: `Type "hello"`,
			want:  []Action{{Kind: ActionType, Text: "hello", Speed: 50 * time.Millisecond}},
		},
		{
			name:  "type with speed",
			input: "Type@30ms 'hello'",
			want:  []Action{{Kind: ActionType, Text: "hello", Speed: 30 * time.Millisecond}},
		},
		{
			name:  "type with speed in seconds",
			input: "Type@1s 'hello'",
			want:  []Action{{Kind: ActionType, Text: "hello", Speed: 1 * time.Second}},
		},
		{
			name:  "type with spaces in text",
			input: "Type 'ls -la'",
			want:  []Action{{Kind: ActionType, Text: "ls -la", Speed: 50 * time.Millisecond}},
		},
		{
			name:  "sleep with ms",
			input: "Sleep 500ms",
			want:  []Action{{Kind: ActionSleep, Duration: 500 * time.Millisecond}},
		},
		{
			name:  "sleep with seconds",
			input: "Sleep 2s",
			want:  []Action{{Kind: ActionSleep, Duration: 2 * time.Second}},
		},
		{
			name:  "key simple - Enter",
			input: "Enter",
			want:  []Action{{Kind: ActionKey, Key: "Enter", Repeat: 1}},
		},
		{
			name:  "key simple - Tab (lowercase)",
			input: "tab",
			want:  []Action{{Kind: ActionKey, Key: "tab", Repeat: 1}},
		},
		{
			name:  "key simple - Escape",
			input: "Escape",
			want:  []Action{{Kind: ActionKey, Key: "Escape", Repeat: 1}},
		},
		{
			name:  "key simple - Space",
			input: "Space",
			want:  []Action{{Kind: ActionKey, Key: "Space", Repeat: 1}},
		},
		{
			name:  "key simple - Backspace",
			input: "Backspace",
			want:  []Action{{Kind: ActionKey, Key: "Backspace", Repeat: 1}},
		},
		{
			name:  "key simple - Delete",
			input: "Delete",
			want:  []Action{{Kind: ActionKey, Key: "Delete", Repeat: 1}},
		},
		{
			name:  "key simple - Up",
			input: "Up",
			want:  []Action{{Kind: ActionKey, Key: "Up", Repeat: 1}},
		},
		{
			name:  "key simple - Down",
			input: "Down",
			want:  []Action{{Kind: ActionKey, Key: "Down", Repeat: 1}},
		},
		{
			name:  "key simple - Left",
			input: "Left",
			want:  []Action{{Kind: ActionKey, Key: "Left", Repeat: 1}},
		},
		{
			name:  "key simple - Right",
			input: "Right",
			want:  []Action{{Kind: ActionKey, Key: "Right", Repeat: 1}},
		},
		{
			name:  "key simple - Home",
			input: "Home",
			want:  []Action{{Kind: ActionKey, Key: "Home", Repeat: 1}},
		},
		{
			name:  "key simple - End",
			input: "End",
			want:  []Action{{Kind: ActionKey, Key: "End", Repeat: 1}},
		},
		{
			name:  "key simple - PageUp",
			input: "PageUp",
			want:  []Action{{Kind: ActionKey, Key: "PageUp", Repeat: 1}},
		},
		{
			name:  "key simple - PageDown",
			input: "PageDown",
			want:  []Action{{Kind: ActionKey, Key: "PageDown", Repeat: 1}},
		},
		{
			name:  "key repeat",
			input: "Down 3",
			want:  []Action{{Kind: ActionKey, Key: "Down", Repeat: 3}},
		},
		{
			name:  "key repeat large",
			input: "Enter 10",
			want:  []Action{{Kind: ActionKey, Key: "Enter", Repeat: 10}},
		},
		{
			name:  "key delay",
			input: "Enter@200ms",
			want:  []Action{{Kind: ActionKey, Key: "Enter", Delay: 200 * time.Millisecond, Repeat: 1}},
		},
		{
			name:  "key delay in seconds",
			input: "Tab@1s",
			want:  []Action{{Kind: ActionKey, Key: "Tab", Delay: 1 * time.Second, Repeat: 1}},
		},
		{
			name:  "key with delay and repeat",
			input: "Down@500ms 3",
			want:  []Action{{Kind: ActionKey, Key: "Down", Delay: 500 * time.Millisecond, Repeat: 3}},
		},
		{
			name:  "ctrl combo - Ctrl+C",
			input: "Ctrl+C",
			want:  []Action{{Kind: ActionCtrl, Key: "c"}},
		},
		{
			name:  "ctrl combo - Ctrl+D",
			input: "Ctrl+D",
			want:  []Action{{Kind: ActionCtrl, Key: "d"}},
		},
		{
			name:  "ctrl combo - Ctrl+L (lowercase)",
			input: "ctrl+l",
			want:  []Action{{Kind: ActionCtrl, Key: "l"}},
		},
		{
			name:  "ctrl combo - Ctrl+Z",
			input: "Ctrl+Z",
			want:  []Action{{Kind: ActionCtrl, Key: "z"}},
		},
		{
			name:  "complex script",
			input: "Sleep 1s Type 'ls -la' Enter Sleep 500ms",
			want: []Action{
				{Kind: ActionSleep, Duration: time.Second},
				{Kind: ActionType, Text: "ls -la", Speed: 50 * time.Millisecond},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
				{Kind: ActionSleep, Duration: 500 * time.Millisecond},
			},
		},
		{
			name:  "type echo and enter",
			input: "Type 'echo hello' Enter",
			want: []Action{
				{Kind: ActionType, Text: "echo hello", Speed: 50 * time.Millisecond},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
			},
		},
		{
			name:  "fzf navigation",
			input: "Down 5 Enter",
			want: []Action{
				{Kind: ActionKey, Key: "Down", Repeat: 5},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
			},
		},
		{
			name:  "cat with Ctrl+D",
			input: "Type 'hello' Enter Sleep 500ms Ctrl+D",
			want: []Action{
				{Kind: ActionType, Text: "hello", Speed: 50 * time.Millisecond},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
				{Kind: ActionSleep, Duration: 500 * time.Millisecond},
				{Kind: ActionCtrl, Key: "d"},
			},
		},
		{
			name:    "missing quote - type without quotes",
			input:   "Type hello",
			wantErr: "expected quoted string",
		},
		{
			name:    "unknown key - Foo",
			input:   "Foo",
			wantErr: "unknown key",
		},
		{
			name:    "unknown key - random",
			input:   "Type 'test' Random",
			wantErr: "unknown key",
		},
		{
			name:    "invalid duration - missing unit",
			input:   "Sleep 500",
			wantErr: "invalid duration",
		},
		{
			name:    "invalid duration - bad format",
			input:   "Sleep 500x",
			wantErr: "invalid duration",
		},
		{
			name:    "expected duration after at",
			input:   "Type@ 'hello'",
			wantErr: "expected duration after @",
		},
		{
			name:    "expected duration after sleep",
			input:   "Sleep",
			wantErr: "expected duration",
		},
		{
			name:  "empty input",
			input: "",
			want:  []Action{},
		},
		{
			name:  "whitespace only",
			input: "   \t\n  ",
			want:  []Action{},
		},
		{
			name:  "type with fast speed",
			input: "Type@20ms 'fast typing test'",
			want:  []Action{{Kind: ActionType, Text: "fast typing test", Speed: 20 * time.Millisecond}},
		},
		{
			name:  "enter with delay",
			input: "Enter@500ms",
			want:  []Action{{Kind: ActionKey, Key: "Enter", Delay: 500 * time.Millisecond, Repeat: 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  string
		position int
	}{
		{
			name:     "unknown key position",
			input:    "Foo",
			wantErr:  "unknown key",
			position: 0,
		},
		{
			name:     "unknown key in sequence",
			input:    "Type 'test' Foo",
			wantErr:  "unknown key",
			position: 12,
		},
		{
			name:     "missing quote position",
			input:    "Type hello",
			wantErr:  "expected quoted string",
			position: 5,
		},
		{
			name:     "invalid duration position",
			input:    "Sleep 500",
			wantErr:  "invalid duration",
			position: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok, "expected ParseError type")

			assert.Contains(t, parseErr.Error(), tt.wantErr)
			assert.Equal(t, tt.position, parseErr.Position)
		})
	}
}

func TestParseComplexScripts(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Action
	}{
		{
			name:  "full demo script",
			input: "Sleep 1s Type 'ls -la' Enter Sleep 500ms Type 'echo done' Enter",
			want: []Action{
				{Kind: ActionSleep, Duration: 1 * time.Second},
				{Kind: ActionType, Text: "ls -la", Speed: 50 * time.Millisecond},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
				{Kind: ActionSleep, Duration: 500 * time.Millisecond},
				{Kind: ActionType, Text: "echo done", Speed: 50 * time.Millisecond},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
			},
		},
		{
			name:  "vim-like navigation",
			input: "Down Down Down Enter",
			want: []Action{
				{Kind: ActionKey, Key: "Down", Repeat: 1},
				{Kind: ActionKey, Key: "Down", Repeat: 1},
				{Kind: ActionKey, Key: "Down", Repeat: 1},
				{Kind: ActionKey, Key: "Enter", Repeat: 1},
			},
		},
		{
			name:  "multiple sleeps",
			input: "Sleep 100ms Sleep 200ms Sleep 300ms",
			want: []Action{
				{Kind: ActionSleep, Duration: 100 * time.Millisecond},
				{Kind: ActionSleep, Duration: 200 * time.Millisecond},
				{Kind: ActionSleep, Duration: 300 * time.Millisecond},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLexer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []token
	}{
		{
			name:  "simple tokens",
			input: "Type 'hello'",
			want: []token{
				{kind: tokenIdent, literal: "Type"},
				{kind: tokenString, literal: "hello"},
				{kind: tokenEOF},
			},
		},
		{
			name:  "duration token",
			input: "Sleep 500ms",
			want: []token{
				{kind: tokenIdent, literal: "Sleep"},
				{kind: tokenDuration, literal: "500ms"},
				{kind: tokenEOF},
			},
		},
		{
			name:  "at token",
			input: "Type@30ms",
			want: []token{
				{kind: tokenIdent, literal: "Type"},
				{kind: tokenAt, literal: "@"},
				{kind: tokenDuration, literal: "30ms"},
				{kind: tokenEOF},
			},
		},
		{
			name:  "ctrl token",
			input: "Ctrl+C",
			want: []token{
				{kind: tokenIdent, literal: "Ctrl+C"},
				{kind: tokenEOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newLexer(tt.input)
			var got []token

			for {
				tok := l.nextToken()
				got = append(got, tok)
				if tok.kind == tokenEOF {
					break
				}
			}

			// Compare tokens without positions for simplicity
			require.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i].kind, got[i].kind, "token %d kind mismatch", i)
				assert.Equal(t, tt.want[i].literal, got[i].literal, "token %d literal mismatch", i)
			}
		})
	}
}
