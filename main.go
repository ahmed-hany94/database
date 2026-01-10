package main

import (
	"log"
)

type Database struct {
	statement []Statement
}

func (d *Database) execute(statement *Statement) {}

func main() {
	sql := "CREATE TABLE users (id INT, name VARCHAR(50));"
	tokens := tokenize(sql)

	for _, tok := range tokens {
		log.Printf("%-15s %-12s at position %d\n",
			tok.Content,
			tokenTypeName(tok.Type),
			tok.Position)
	}

	parser := NewParser(tokens)
	statement, err := parser.parse()
	if err != nil {
		log.Fatalf("Parse error: %v\n", err)
		return
	}

	database := &Database{}
	database.execute(statement)
}
