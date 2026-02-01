package script

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// tokenKind represents the type of a token.
type tokenKind int

const (
	tokenEOF      tokenKind = iota
	tokenIdent              // Type, Sleep, Enter, Down, Ctrl
	tokenString             // 'quoted' or "quoted"
	tokenNumber             // 123
	tokenDuration           // 500ms, 2s
	tokenAt                 // @
	tokenPlus               // +
)

// token represents a lexical token with its kind, literal value, and position.
type token struct {
	kind     tokenKind
	literal  string
	position int // byte position in input
}

// lexer scans the input script and produces tokens.
type lexer struct {
	input    string
	position int  // current position in input (points to current char)
	readPos  int  // current reading position (after current char)
	ch       byte // current char under examination
}

// newLexer creates a new lexer for the given input string.
func newLexer(input string) *lexer {
	l := &lexer{input: input}
	l.readChar()
	return l
}

// readChar advances the lexer to the next character.
func (l *lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // ASCII code for NUL, signifies EOF
	} else {
		l.ch = l.input[l.readPos]
	}
	l.position = l.readPos
	l.readPos++
}

// peekChar returns the next character without advancing.
func (l *lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// skipWhitespace skips spaces, tabs, newlines, and carriage returns.
func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	l.skipWhitespace()

	pos := l.position

	switch l.ch {
	case '@':
		l.readChar()
		return token{kind: tokenAt, literal: "@", position: pos}
	case '+':
		l.readChar()
		return token{kind: tokenPlus, literal: "+", position: pos}
	case '\'':
		return l.readString('\'')
	case '"':
		return l.readString('"')
	case 0:
		return token{kind: tokenEOF, literal: "", position: pos}
	default:
		if isDigit(l.ch) {
			return l.readNumberOrDuration()
		}
		if isLetter(l.ch) {
			return l.readIdent()
		}
		// Unknown character - skip it and return next token
		l.readChar()
		return l.nextToken()
	}
}

// readString reads a quoted string (single or double quotes).
func (l *lexer) readString(quote byte) token {
	pos := l.position
	l.readChar() // consume opening quote

	var sb strings.Builder
	for l.ch != quote && l.ch != 0 {
		sb.WriteByte(l.ch)
		l.readChar()
	}

	l.readChar() // consume closing quote
	return token{kind: tokenString, literal: sb.String(), position: pos}
}

// readNumberOrDuration reads a number, which may be followed by 'ms' or 's' to form a duration.
func (l *lexer) readNumberOrDuration() token {
	pos := l.position
	var sb strings.Builder

	for isDigit(l.ch) {
		sb.WriteByte(l.ch)
		l.readChar()
	}

	numStr := sb.String()

	// Check for duration suffix
	if l.ch == 'm' && l.peekChar() == 's' {
		sb.WriteByte(l.ch)
		l.readChar()
		sb.WriteByte(l.ch)
		l.readChar()
		return token{kind: tokenDuration, literal: sb.String(), position: pos}
	}

	if l.ch == 's' {
		sb.WriteByte(l.ch)
		l.readChar()
		return token{kind: tokenDuration, literal: sb.String(), position: pos}
	}

	// Just a number
	return token{kind: tokenNumber, literal: numStr, position: pos}
}

// readIdent reads an identifier (sequence of letters and digits, case-insensitive).
func (l *lexer) readIdent() token {
	pos := l.position
	var sb strings.Builder

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '+' {
		sb.WriteByte(l.ch)
		l.readChar()
	}

	return token{kind: tokenIdent, literal: sb.String(), position: pos}
}

// isLetter checks if a byte is a letter (or underscore).
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// isDigit checks if a byte is a digit.
func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

// parser parses tokens into Actions.
type parser struct {
	l         *lexer
	curToken  token
	peekToken token
}

// newParser creates a new parser for the given lexer.
func newParser(l *lexer) *parser {
	p := &parser{l: l}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances the parser to the next token.
func (p *parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.nextToken()
}

// ParseError represents a parsing error with position information.
type ParseError struct {
	Position int
	Message  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at position %d: %s", e.Position, e.Message)
}

// validKeys contains all recognized special key names (case-insensitive).
var validKeys = map[string]bool{
	"enter":     true,
	"tab":       true,
	"escape":    true,
	"space":     true,
	"backspace": true,
	"delete":    true,
	"up":        true,
	"down":      true,
	"left":      true,
	"right":     true,
	"home":      true,
	"end":       true,
	"pageup":    true,
	"pagedown":  true,
}

// isValidKey checks if a key name is valid.
func isValidKey(key string) bool {
	lowerKey := strings.ToLower(key)
	if validKeys[lowerKey] {
		return true
	}
	// Check for Ctrl combinations
	if strings.HasPrefix(lowerKey, "ctrl+") && len(lowerKey) == 6 {
		return true
	}
	return false
}

// parseDuration parses a duration string (e.g., "500ms", "2s").
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// Parse converts a tape script string into a slice of Actions.
// Returns error with position info on parse failure.
func Parse(script string) ([]Action, error) {
	l := newLexer(script)
	p := newParser(l)

	actions := []Action{}

	for p.curToken.kind != tokenEOF {
		action, err := p.parseAction()
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// parseAction parses a single action from the current token.
func (p *parser) parseAction() (Action, error) {
	if p.curToken.kind != tokenIdent {
		return Action{}, &ParseError{
			Position: p.curToken.position,
			Message:  fmt.Sprintf("expected command or key, got %s", p.curToken.literal),
		}
	}

	ident := strings.ToLower(p.curToken.literal)

	// Check for Ctrl+ combinations first
	if strings.HasPrefix(ident, "ctrl+") {
		return p.parseCtrlAction()
	}

	// Check for Type command
	if ident == "type" {
		return p.parseTypeAction()
	}

	// Check for Sleep command
	if ident == "sleep" {
		return p.parseSleepAction()
	}

	// Otherwise, treat as a key press
	return p.parseKeyAction()
}

// parseTypeAction parses a Type command with optional speed modifier.
func (p *parser) parseTypeAction() (Action, error) {
	action := Action{Kind: ActionType, Speed: 50 * time.Millisecond}

	p.nextToken() // consume 'Type'

	// Check for @speed modifier
	if p.curToken.kind == tokenAt {
		p.nextToken() // consume '@'
		if p.curToken.kind != tokenDuration && p.curToken.kind != tokenNumber {
			return Action{}, &ParseError{
				Position: p.curToken.position,
				Message:  "expected duration after @",
			}
		}

		duration, err := parseDuration(p.curToken.literal)
		if err != nil {
			return Action{}, &ParseError{
				Position: p.curToken.position,
				Message:  fmt.Sprintf("invalid duration %q; use '500ms' or '2s'", p.curToken.literal),
			}
		}
		action.Speed = duration
		p.nextToken() // consume duration
	}

	// Expect quoted string
	if p.curToken.kind != tokenString {
		return Action{}, &ParseError{
			Position: p.curToken.position,
			Message:  "expected quoted string after Type",
		}
	}

	action.Text = p.curToken.literal
	p.nextToken() // consume string

	return action, nil
}

// parseSleepAction parses a Sleep command with duration.
func (p *parser) parseSleepAction() (Action, error) {
	action := Action{Kind: ActionSleep}

	p.nextToken() // consume 'Sleep'

	// Expect duration
	if p.curToken.kind != tokenDuration && p.curToken.kind != tokenNumber {
		return Action{}, &ParseError{
			Position: p.curToken.position,
			Message:  "expected duration after Sleep",
		}
	}

	duration, err := parseDuration(p.curToken.literal)
	if err != nil {
		return Action{}, &ParseError{
			Position: p.curToken.position,
			Message:  fmt.Sprintf("invalid duration %q; use '500ms' or '2s'", p.curToken.literal),
		}
	}
	action.Duration = duration
	p.nextToken() // consume duration

	return action, nil
}

// parseKeyAction parses a key press command with optional repeat count or delay modifier.
func (p *parser) parseKeyAction() (Action, error) {
	keyName := p.curToken.literal
	pos := p.curToken.position

	if !isValidKey(keyName) {
		return Action{}, &ParseError{
			Position: pos,
			Message:  fmt.Sprintf("unknown key %q; valid keys: Enter, Tab, Escape, Space, Backspace, Delete, Up, Down, Left, Right, Home, End, PageUp, PageDown, Ctrl+C, Ctrl+D, Ctrl+L, Ctrl+Z", keyName),
		}
	}

	action := Action{
		Kind:   ActionKey,
		Key:    keyName,
		Repeat: 1,
	}

	p.nextToken() // consume key name

	// Check for @delay modifier
	if p.curToken.kind == tokenAt {
		p.nextToken() // consume '@'
		if p.curToken.kind != tokenDuration && p.curToken.kind != tokenNumber {
			return Action{}, &ParseError{
				Position: p.curToken.position,
				Message:  "expected duration after @",
			}
		}

		duration, err := parseDuration(p.curToken.literal)
		if err != nil {
			return Action{}, &ParseError{
				Position: p.curToken.position,
				Message:  fmt.Sprintf("invalid duration %q; use '500ms' or '2s'", p.curToken.literal),
			}
		}
		action.Delay = duration
		p.nextToken() // consume duration
	}

	// Check for repeat count
	if p.curToken.kind == tokenNumber {
		repeat, err := strconv.Atoi(p.curToken.literal)
		if err != nil {
			return Action{}, &ParseError{
				Position: p.curToken.position,
				Message:  fmt.Sprintf("invalid repeat count %q", p.curToken.literal),
			}
		}
		action.Repeat = repeat
		p.nextToken() // consume number
	}

	return action, nil
}

// parseCtrlAction parses a Ctrl+ combination.
func (p *parser) parseCtrlAction() (Action, error) {
	keyName := p.curToken.literal

	action := Action{
		Kind: ActionCtrl,
		Key:  strings.ToLower(keyName[5:]), // Extract key after "Ctrl+"
	}

	p.nextToken() // consume Ctrl+key

	return action, nil
}
