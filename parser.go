package main

import (
	"fmt"
	"strconv"
)

// Column types
const (
	INT_TYPE = iota
	VARCHAR_TYPE
)

type Column struct {
	Name        string
	Length      int
	FixedSize   int
	MaxSize     int
	NeedsLength bool
	// Type        int
}

// Token types
const (
	CREATE_STMNT = iota
)

type Statement struct {
	TableName string
	Values    []Column
}

func makeStatement() Statement {
	return Statement{
		TableName: "",
		Values:    []Column{},
	}
}

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

func (p *Parser) peek() Token {
	if p.current < len(p.tokens) {
		return p.tokens[p.current]
	}

	return Token{"", EOF, -1}
}

func (p *Parser) advance() Token {
	token := p.peek()
	if p.current < len(p.tokens) {
		p.current++
	}

	return token
}

func (p *Parser) expect(tokenType int) (Token, error) {
	token := p.peek()
	if token.Type != tokenType {
		return token, fmt.Errorf("expected token type %d, got %d ('%s') at position %d",
			tokenType, token.Type, token.Content, token.Position)
	} else {
		return p.advance(), nil
	}
}

func (p *Parser) parse() (*Statement, error) {
	firstTok := p.peek()

	switch firstTok.Type {
	case CREATE_TOK:
		return p.parseCreateTable()
	default:
		return nil, fmt.Errorf("Unexpected token at start: '%s'", firstTok.Content)
	}
}

func (p *Parser) parseCreateTable() (*Statement, error) {
	statement := makeStatement()

	if _, err := p.expect(CREATE_TOK); err != nil {
		return nil, err
	}

	if _, err := p.expect(TABLE_TOK); err != nil {
		return nil, err
	}

	tableNameToken, err := p.expect(IDENTIFIER)
	if err != nil {
		return nil, err
	}

	statement.TableName = tableNameToken.Content

	if p.peek().Type == SEMICOLON {
		return &statement, nil
	}

	_, err = p.expect(LEFT_PAREN)
	if err != nil {
		return nil, err
	}

	// id INT, name VARCHAR(50);
	for p.peek().Type != RIGHT_PAREN {
		colName, err := p.expect(IDENTIFIER)
		if err != nil {
			return nil, err
		}

		colType, err := p.expect(IDENTIFIER)
		if err != nil {
			return nil, err
		}

		var length int = 0
		if p.peek().Type == LEFT_PAREN {
			p.expect(LEFT_PAREN)
			lenTok, err := p.expect(NUMBER)
			if err != nil {
				return nil, err
			}
			p.expect(RIGHT_PAREN)

			length, err = strconv.Atoi(lenTok.Content)
			if err != nil {
				return nil, err
			}
		}

		if p.peek().Type == COMMA {
			_, err := p.expect(COMMA)
			if err != nil {
				return nil, err
			}
		}

		statement.Values = append(statement.Values, Column{
			Name: colName.Content,
			// Type:        colType.Type,
			Length:      length,
			FixedSize:   TypeMap[colType.Content].FixedSize,
			MaxSize:     TypeMap[colType.Content].MaxSize,
			NeedsLength: TypeMap[colType.Content].NeedsLength,
		})
	}

	_, err = p.expect(RIGHT_PAREN)
	if err != nil {
		return nil, err
	}

	return &statement, nil
}

var TypeMap = map[string]Column{
	"TINYINT":    {Name: "TINYINT", FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"SMALLINT":   {Name: "SMALLINT", FixedSize: 2, MaxSize: 2, NeedsLength: false},
	"MEDIUMINT":  {Name: "MEDIUMINT", FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"INT":        {Name: "INT", FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"INTEGER":    {Name: "INTEGER", FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"BIGINT":     {Name: "BIGINT", FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"FLOAT":      {Name: "FLOAT", FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"DOUBLE":     {Name: "DOUBLE", FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"REAL":       {Name: "REAL", FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"BOOL":       {Name: "BOOL", FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"BOOLEAN":    {Name: "BOOLEAN", FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"DATE":       {Name: "DATE", FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"TIME":       {Name: "TIME", FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"DATETIME":   {Name: "DATETIME", FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"TIMESTAMP":  {Name: "TIMESTAMP", FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"YEAR":       {Name: "YEAR", FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"CHAR":       {Name: "CHAR", FixedSize: 0, MaxSize: 255, NeedsLength: true},
	"VARCHAR":    {Name: "VARCHAR", FixedSize: 0, MaxSize: 65535, NeedsLength: true},
	"BINARY":     {Name: "BINARY", FixedSize: 0, MaxSize: 255, NeedsLength: true},
	"VARBINARY":  {Name: "VARBINARY", FixedSize: 0, MaxSize: 65535, NeedsLength: true},
	"TEXT":       {Name: "TEXT", FixedSize: 0, MaxSize: 65535, NeedsLength: false},
	"MEDIUMTEXT": {Name: "MEDIUMTEXT", FixedSize: 0, MaxSize: 16777215, NeedsLength: false},
	"LONGTEXT":   {Name: "LONGTEXT", FixedSize: 0, MaxSize: 4294967295, NeedsLength: false},
	"BLOB":       {Name: "BLOB", FixedSize: 0, MaxSize: 65535, NeedsLength: false},
	"MEDIUMBLOB": {Name: "MEDIUMBLOB", FixedSize: 0, MaxSize: 16777215, NeedsLength: false},
	"LONGBLOB":   {Name: "LONGBLOB", FixedSize: 0, MaxSize: 4294967295, NeedsLength: false},
}
