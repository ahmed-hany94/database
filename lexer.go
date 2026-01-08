package main

import (
	"strings"
)

// Token types
const (
	CREATE_TOK = iota
	TABLE_TOK
	INSERT_TOK
	INTO_TOK
	VALUES_TOK
	IDENTIFIER
	NUMBER
	STRING_LITERAL
	LEFT_PAREN
	RIGHT_PAREN
	COMMA
	SEMICOLON
	UNKNOWN
	EOF
)

// Lexer states
const (
	STATE_START = iota
	STATE_IN_IDENTIFIER
	STATE_IN_NUMBER
	STATE_IN_STRING
)

type Token struct {
	Content  string
	Type     int
	Position int
}

// tokenize converts a SQL command into tokens
func tokenize(cmd string) []Token {
	tokens := make([]Token, 0, 32)
	var buf strings.Builder
	buf.Grow(32)

	state := STATE_START
	start := 0
	runes := []rune(cmd)

	for i := 0; i < len(runes); i++ {
		c := runes[i]

		switch state {
		case STATE_START:
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
				state = STATE_IN_IDENTIFIER
				start = i
				buf.WriteRune(c)
			} else if c >= '0' && c <= '9' {
				state = STATE_IN_NUMBER
				start = i
				buf.WriteRune(c)
			} else if c == '\'' || c == '"' {
				state = STATE_IN_STRING
				start = i
			} else if c == '(' {
				tokens = append(tokens, Token{"(", LEFT_PAREN, i})
			} else if c == ')' {
				tokens = append(tokens, Token{")", RIGHT_PAREN, i})
			} else if c == ',' {
				tokens = append(tokens, Token{",", COMMA, i})
			} else if c == ';' {
				tokens = append(tokens, Token{";", SEMICOLON, i})
			} else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				// Skip whitespace
			} else {
				tokens = append(tokens, Token{string(c), UNKNOWN, i})
			}

		case STATE_IN_IDENTIFIER:
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_' {
				buf.WriteRune(c)
			} else {
				content := buf.String()
				tokType := keywordOrIdentifier(content)
				tokens = append(tokens, Token{content, tokType, start})
				buf.Reset()
				state = STATE_START
				i--
			}

		case STATE_IN_NUMBER:
			if c >= '0' && c <= '9' {
				buf.WriteRune(c)
			} else {
				tokens = append(tokens, Token{buf.String(), NUMBER, start})
				buf.Reset()
				state = STATE_START
				i--
			}

		case STATE_IN_STRING:
			if c == '\'' || c == '"' {
				tokens = append(tokens, Token{buf.String(), STRING_LITERAL, start})
				buf.Reset()
				state = STATE_START
			} else {
				buf.WriteRune(c)
			}
		}
	}

	// Handle remaining token
	if buf.Len() > 0 {
		content := buf.String()
		switch state {
		case STATE_IN_IDENTIFIER:
			tokens = append(tokens, Token{content, keywordOrIdentifier(content), start})
		case STATE_IN_NUMBER:
			tokens = append(tokens, Token{content, NUMBER, start})
		case STATE_IN_STRING:
			tokens = append(tokens, Token{content, STRING_LITERAL, start})
		}
	}

	return tokens
}

// keywordOrIdentifier returns the token type for a keyword, or IDENTIFIER
func keywordOrIdentifier(word string) int {
	switch strings.ToUpper(word) {
	case "CREATE":
		return CREATE_TOK
	case "TABLE":
		return TABLE_TOK
	case "INSERT":
		return INSERT_TOK
	case "INTO":
		return INTO_TOK
	case "VALUES":
		return VALUES_TOK
	default:
		return IDENTIFIER
	}
}

// tokenTypeName returns a human-readable name for a token type
func tokenTypeName(t int) string {
	switch t {
	case CREATE_TOK:
		return "CREATE"
	case TABLE_TOK:
		return "TABLE"
	case INSERT_TOK:
		return "INSERT"
	case INTO_TOK:
		return "INTO"
	case VALUES_TOK:
		return "VALUES"
	case IDENTIFIER:
		return "IDENTIFIER"
	case NUMBER:
		return "NUMBER"
	case STRING_LITERAL:
		return "STRING"
	case LEFT_PAREN:
		return "("
	case RIGHT_PAREN:
		return ")"
	case COMMA:
		return ","
	case SEMICOLON:
		return ";"
	case UNKNOWN:
		return "UNKNOWN"
	case EOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}
