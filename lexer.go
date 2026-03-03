package main

import (
	"strings"
)

// Token types
const (
	CreateTok = iota
	TableTok
	InsertTok
	IntoTok
	ValuesTok
	Identifier
	Number
	StringLiteral
	LeftParen
	RightParen
	Comma
	Semicolon
	Unknown
	Eof
)

// Lexer states
const (
	stateStart = iota
	stateInIdentifier
	stateInNumber
	stateInString
)

type Token struct {
	Content  string
	Type     int
	Position int
}

// tokenize converts a SQL Command into tokens
func tokenize(cmd string) []Token {
	tokens := make([]Token, 0, 32)
	var buf strings.Builder
	buf.Grow(32)

	state := stateStart
	start := 0
	runes := []rune(cmd)

	for i := 0; i < len(runes); i++ {
		c := runes[i]

		switch state {
		case stateStart:
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
				state = stateInIdentifier
				start = i
				buf.WriteRune(c)
			} else if c >= '0' && c <= '9' {
				state = stateInNumber
				start = i
				buf.WriteRune(c)
			} else if c == '\'' || c == '"' {
				state = stateInString
				start = i
			} else if c == '(' {
				tokens = append(tokens, Token{"(", LeftParen, i})
			} else if c == ')' {
				tokens = append(tokens, Token{")", RightParen, i})
			} else if c == ',' {
				tokens = append(tokens, Token{",", Comma, i})
			} else if c == ';' {
				tokens = append(tokens, Token{";", Semicolon, i})
			} else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				// Skip whitespace
			} else {
				tokens = append(tokens, Token{string(c), Unknown, i})
			}

		case stateInIdentifier:
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_' {
				buf.WriteRune(c)
			} else {
				content := buf.String()
				tokType := keywordOrIdentifier(content)
				tokens = append(tokens, Token{content, tokType, start})
				buf.Reset()
				state = stateStart
				i--
			}

		case stateInNumber:
			if c >= '0' && c <= '9' {
				buf.WriteRune(c)
			} else {
				tokens = append(tokens, Token{buf.String(), Number, start})
				buf.Reset()
				state = stateStart
				i--
			}

		case stateInString:
			if c == '\'' || c == '"' {
				tokens = append(tokens, Token{buf.String(), StringLiteral, start})
				buf.Reset()
				state = stateStart
			} else {
				buf.WriteRune(c)
			}
		}
	}

	// Handle remaining token
	if buf.Len() > 0 {
		content := buf.String()
		switch state {
		case stateInIdentifier:
			tokens = append(tokens, Token{content, keywordOrIdentifier(content), start})
		case stateInNumber:
			tokens = append(tokens, Token{content, Number, start})
		case stateInString:
			tokens = append(tokens, Token{content, StringLiteral, start})
		}
	}

	return tokens
}

// keywordOrIdentifier returns the token type for a keyword, or Identifier
func keywordOrIdentifier(word string) int {
	switch strings.ToUpper(word) {
	case "CREATE":
		return CreateTok
	case "TABLE":
		return TableTok
	case "INSERT":
		return InsertTok
	case "INTO":
		return IntoTok
	case "VALUES":
		return ValuesTok
	default:
		return Identifier
	}
}

// tokenTypeName returns a human-readable name for a token type
func tokenTypeName(t int) string {
	switch t {
	case CreateTok:
		return "CREATE"
	case TableTok:
		return "TABLE"
	case InsertTok:
		return "INSERT"
	case IntoTok:
		return "INTO"
	case ValuesTok:
		return "VALUES"
	case Identifier:
		return "Identifier"
	case Number:
		return "Number"
	case StringLiteral:
		return "STRING"
	case LeftParen:
		return "("
	case RightParen:
		return ")"
	case Comma:
		return ","
	case Semicolon:
		return ";"
	case Unknown:
		return "Unknown"
	case Eof:
		return "Eof"
	default:
		return "Unknown"
	}
}
