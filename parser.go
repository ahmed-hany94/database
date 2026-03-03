package main

import (
	"fmt"
	"strconv"
)

// Column types
const (
	INT_TYPE = iota
	FLOAT_TYPE
	BOOL_TYPE
	TEMPORAL_TYPE
	STRING_TYPE
	BINARY_TYPE
)

type Column struct {
	Name        string
	Length      int
	FixedSize   int
	MaxSize     int
	NeedsLength bool
	Type        int
}

type Value struct {
	ColumnName string
	Raw        string
	TokenType  int
}

// Token types
const (
	UNDEF_STMNT = iota
	CREATE_STMNT
	INSERT_STMNT
)

type Statement struct {
	TableName string
	Type      int
	Columns   []Column
	Rows      [][]Value
}

func NewStatement() Statement {
	return Statement{
		TableName: "",
		Type:      UNDEF_STMNT,
		Columns:   []Column{},
		Rows:      [][]Value{},
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

	return Token{"", Eof, -1}
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
	case CreateTok:
		return p.parseCreateTable()
	case InsertTok:
		return p.parseInsertStatement()
	default:
		return nil, fmt.Errorf("Unexpected token at start: '%s'", firstTok.Content)
	}
}

func (p *Parser) parseCreateTable() (*Statement, error) {
	statement := NewStatement()

	if _, err := p.expect(CreateTok); err != nil {
		return nil, err
	}

	if _, err := p.expect(TableTok); err != nil {
		return nil, err
	}

	tableNameToken, err := p.expect(Identifier)
	if err != nil {
		return nil, err
	}

	statement.TableName = tableNameToken.Content

	if p.peek().Type == Semicolon {
		return &statement, nil
	}

	_, err = p.expect(LeftParen)
	if err != nil {
		return nil, err
	}

	// id INT, name VARCHAR(50);
	for p.peek().Type != RightParen {
		colName, err := p.expect(Identifier)
		if err != nil {
			return nil, err
		}

		colType, err := p.expect(Identifier)
		if err != nil {
			return nil, err
		}

		var length int = 0
		if p.peek().Type == LeftParen {
			p.expect(LeftParen)
			lenTok, err := p.expect(Number)
			if err != nil {
				return nil, err
			}
			p.expect(RightParen)

			length, err = strconv.Atoi(lenTok.Content)
			if err != nil {
				return nil, err
			}
		}

		if p.peek().Type == Comma {
			_, err := p.expect(Comma)
			if err != nil {
				return nil, err
			}
		}

		statement.Columns = append(statement.Columns, Column{
			Name:        colName.Content,
			Type:        TypeMap[colType.Content].Type,
			Length:      length,
			FixedSize:   TypeMap[colType.Content].FixedSize,
			MaxSize:     TypeMap[colType.Content].MaxSize,
			NeedsLength: TypeMap[colType.Content].NeedsLength,
		})
	}

	_, err = p.expect(RightParen)
	if err != nil {
		return nil, err
	}

	statement.Type = CREATE_STMNT
	return &statement, nil
}

func (p *Parser) parseInsertStatement() (*Statement, error) {
	statement := NewStatement()

	if _, err := p.expect(InsertTok); err != nil {
		return nil, err
	}

	if _, err := p.expect(IntoTok); err != nil {
		return nil, err
	}

	tableNameToken, err := p.expect(Identifier)
	if err != nil {
		return nil, err
	}

	statement.TableName = tableNameToken.Content

	if _, err = p.expect(LeftParen); err != nil {
		return nil, err
	}

	var colNames []string
	for p.peek().Type != RightParen {
		colName, err := p.expect(Identifier)
		if err != nil {
			return nil, err
		}

		colNames = append(colNames, colName.Content)

		if p.peek().Type == Comma {
			p.advance()
		}
	}

	if _, err = p.expect(RightParen); err != nil {
		return nil, err
	}

	if _, err := p.expect(ValuesTok); err != nil {
		return nil, err
	}

	if _, err = p.expect(LeftParen); err != nil {
		return nil, err
	}

	var row []Value
	i := 0
	for p.peek().Type != RightParen {
		if i >= len(colNames) {
			return nil, fmt.Errorf("more values than columns provided")
		}

		tok := p.advance()
		row = append(row, Value{
			ColumnName: colNames[i],
			Raw:        tok.Content,
			TokenType:  tok.Type,
		})
		i++

		if p.peek().Type == Comma {
			p.advance()
		}
	}

	if len(row) < len(colNames) {
		return nil, fmt.Errorf("fewer values than columns provided")
	}

	if _, err = p.expect(RightParen); err != nil {
		return nil, err
	}

	statement.Rows = append(statement.Rows, row)
	statement.Type = INSERT_STMNT

	return &statement, nil
}

var TypeMap = map[string]Column{
	"TINYINT":    {Name: "TINYINT", Type: INT_TYPE, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"SMALLINT":   {Name: "SMALLINT", Type: INT_TYPE, FixedSize: 2, MaxSize: 2, NeedsLength: false},
	"MEDIUMINT":  {Name: "MEDIUMINT", Type: INT_TYPE, FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"INT":        {Name: "INT", Type: INT_TYPE, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"INTEGER":    {Name: "INTEGER", Type: INT_TYPE, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"BIGINT":     {Name: "BIGINT", Type: INT_TYPE, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"FLOAT":      {Name: "FLOAT", Type: FLOAT_TYPE, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"DOUBLE":     {Name: "DOUBLE", Type: FLOAT_TYPE, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"REAL":       {Name: "REAL", Type: FLOAT_TYPE, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"BOOL":       {Name: "BOOL", Type: BOOL_TYPE, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"BOOLEAN":    {Name: "BOOLEAN", Type: BOOL_TYPE, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"DATE":       {Name: "DATE", Type: TEMPORAL_TYPE, FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"TIME":       {Name: "TIME", Type: TEMPORAL_TYPE, FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"DATETIME":   {Name: "DATETIME", Type: TEMPORAL_TYPE, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"TIMESTAMP":  {Name: "TIMESTAMP", Type: TEMPORAL_TYPE, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"YEAR":       {Name: "YEAR", Type: TEMPORAL_TYPE, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"CHAR":       {Name: "CHAR", Type: STRING_TYPE, FixedSize: 0, MaxSize: 255, NeedsLength: true},
	"VARCHAR":    {Name: "VARCHAR", Type: STRING_TYPE, FixedSize: 0, MaxSize: 65535, NeedsLength: true},
	"BINARY":     {Name: "BINARY", Type: BINARY_TYPE, FixedSize: 0, MaxSize: 255, NeedsLength: true},
	"VARBINARY":  {Name: "VARBINARY", Type: BINARY_TYPE, FixedSize: 0, MaxSize: 65535, NeedsLength: true},
	"TEXT":       {Name: "TEXT", Type: STRING_TYPE, FixedSize: 0, MaxSize: 65535, NeedsLength: false},
	"MEDIUMTEXT": {Name: "MEDIUMTEXT", Type: STRING_TYPE, FixedSize: 0, MaxSize: 16777215, NeedsLength: false},
	"LONGTEXT":   {Name: "LONGTEXT", Type: STRING_TYPE, FixedSize: 0, MaxSize: 4294967295, NeedsLength: false},
	"BLOB":       {Name: "BLOB", Type: BINARY_TYPE, FixedSize: 0, MaxSize: 65535, NeedsLength: false},
	"MEDIUMBLOB": {Name: "MEDIUMBLOB", Type: BINARY_TYPE, FixedSize: 0, MaxSize: 16777215, NeedsLength: false},
	"LONGBLOB":   {Name: "LONGBLOB", Type: BINARY_TYPE, FixedSize: 0, MaxSize: 4294967295, NeedsLength: false},
}
