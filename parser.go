package main

import (
	"fmt"
	"strconv"
)

// Column types
const (
	IntType = iota
	FloatType
	BoolType
	TemporalType
	StringType
	BinaryType
)

type Column struct {
	NeedsLength bool
	Length      int
	FixedSize   int
	MaxSize     int
	Type        int
	Name        string
}

type Value struct {
	TokenType  int
	ColumnName string
	Raw        string
	Typed      any
}

// Token types
const (
	UndefStmnt = iota
	CreateStmnt
	InsertStmnt
	SelectStmnt
)

type Statement struct {
	SelectAll       bool
	Type            int
	TableName       string
	SelectedColumns []string
	Columns         []Column
	Rows            [][]Value
}

func NewStatement() Statement {
	return Statement{
		TableName:       "",
		Type:            UndefStmnt,
		Columns:         []Column{},
		Rows:            [][]Value{},
		SelectAll:       false,
		SelectedColumns: []string{},
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

	return Token{Eof, -1, ""}
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
	case SelectTok:
		return p.parseSelectStatement()
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
		statement.Type = CreateStmnt
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

	statement.Type = CreateStmnt
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

	for {
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

		if p.peek().Type != Comma {
			break
		}
		p.advance()
	}
	statement.Type = InsertStmnt

	return &statement, nil
}

func (p *Parser) parseSelectStatement() (*Statement, error) {
	statement := NewStatement()

	if _, err := p.expect(SelectTok); err != nil {
		return nil, err
	}

	if p.peek().Type == StarSelector {
		p.advance()
		statement.SelectAll = true
	} else {
		for p.peek().Type != FromTok {
			col, err := p.expect(Identifier)
			if err != nil {
				return nil, err
			}
			statement.SelectedColumns = append(statement.SelectedColumns, col.Content)
			if p.peek().Type == Comma {
				p.advance()
			}
		}
	}

	if _, err := p.expect(FromTok); err != nil {
		return nil, err
	}

	tableNameToken, err := p.expect(Identifier)
	if err != nil {
		return nil, err
	}
	statement.TableName = tableNameToken.Content
	statement.Type = SelectStmnt

	return &statement, nil
}

var TypeMap = map[string]Column{
	"TINYINT":    {Name: "TINYINT", Type: IntType, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"SMALLINT":   {Name: "SMALLINT", Type: IntType, FixedSize: 2, MaxSize: 2, NeedsLength: false},
	"MEDIUMINT":  {Name: "MEDIUMINT", Type: IntType, FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"INT":        {Name: "INT", Type: IntType, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"INTEGER":    {Name: "INTEGER", Type: IntType, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"BIGINT":     {Name: "BIGINT", Type: IntType, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"FLOAT":      {Name: "FLOAT", Type: FloatType, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"DOUBLE":     {Name: "DOUBLE", Type: FloatType, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"REAL":       {Name: "REAL", Type: FloatType, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"BOOL":       {Name: "BOOL", Type: BoolType, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"BOOLEAN":    {Name: "BOOLEAN", Type: BoolType, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"DATE":       {Name: "DATE", Type: TemporalType, FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"TIME":       {Name: "TIME", Type: TemporalType, FixedSize: 3, MaxSize: 3, NeedsLength: false},
	"DATETIME":   {Name: "DATETIME", Type: TemporalType, FixedSize: 8, MaxSize: 8, NeedsLength: false},
	"TIMESTAMP":  {Name: "TIMESTAMP", Type: TemporalType, FixedSize: 4, MaxSize: 4, NeedsLength: false},
	"YEAR":       {Name: "YEAR", Type: TemporalType, FixedSize: 1, MaxSize: 1, NeedsLength: false},
	"CHAR":       {Name: "CHAR", Type: StringType, FixedSize: 0, MaxSize: 255, NeedsLength: true},
	"VARCHAR":    {Name: "VARCHAR", Type: StringType, FixedSize: 0, MaxSize: 65535, NeedsLength: true},
	"BINARY":     {Name: "BINARY", Type: BinaryType, FixedSize: 0, MaxSize: 255, NeedsLength: true},
	"VARBINARY":  {Name: "VARBINARY", Type: BinaryType, FixedSize: 0, MaxSize: 65535, NeedsLength: true},
	"TEXT":       {Name: "TEXT", Type: StringType, FixedSize: 0, MaxSize: 65535, NeedsLength: false},
	"MEDIUMTEXT": {Name: "MEDIUMTEXT", Type: StringType, FixedSize: 0, MaxSize: 16777215, NeedsLength: false},
	"LONGTEXT":   {Name: "LONGTEXT", Type: StringType, FixedSize: 0, MaxSize: 4294967295, NeedsLength: false},
	"BLOB":       {Name: "BLOB", Type: BinaryType, FixedSize: 0, MaxSize: 65535, NeedsLength: false},
	"MEDIUMBLOB": {Name: "MEDIUMBLOB", Type: BinaryType, FixedSize: 0, MaxSize: 16777215, NeedsLength: false},
	"LONGBLOB":   {Name: "LONGBLOB", Type: BinaryType, FixedSize: 0, MaxSize: 4294967295, NeedsLength: false},
}
