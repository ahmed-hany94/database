package main

import (
	"fmt"
	"log"
)

type Table struct {
	Schema []Column
	Rows   [][]Value
}

type Database struct {
	tables map[string]Table
}

func NewDatabase() Database {
	return Database{
		tables: make(map[string]Table),
	}
}

func (d *Database) execute(statement *Statement) error {
	tableName := statement.TableName
	switch statement.Type {
	case CREATE_STMNT:
		schema := make([]Column, len(statement.Columns))
		copy(schema, statement.Columns)
		d.tables[tableName] = Table{Schema: schema, Rows: [][]Value{}}

	case INSERT_STMNT:
		table, ok := d.tables[tableName]
		if !ok {
			return fmt.Errorf("table '%s' does not exist", tableName)
		}
		for _, row := range statement.Rows {
			table.Rows = append(table.Rows, row)
		}
		d.tables[tableName] = table
	}
	return nil
}

/**
* Journey
*   Basic
*     INSERT Statement
*     SELECT Statement
*     Row Storage
*     Type Validation
*     Basic Error Handling
*   Production
*     Persistence
*       Write rows to disk (binary format, fixed-width for fixed types, length-prefixed for variable)
*       A page/block system (typically 4KB or 16KB pages, like real engines use)
*       A write-ahead log (WAL) so crashes don't corrupt data
*     Indexing
*       B-tree index on primary keys so you're not scanning every row for every query
*       Secondary indexes
*     Query Engine
*       WHERE clause parsing and evaluation
*       JOIN support
*       ORDER BY, GROUP BY, LIMIT
*       A query planner that picks between a full scan vs index scan
*     Transactions
*       ACID guarantees — atomicity, isolation between concurrent readers/writers
*       MVCC (multi-version concurrency control) or locking
*     Concurrency
*       Multiple clients connecting simultaneously
*       A network layer (MySQL wire protocol, or your own)
**/

func main() {
	sql := []string{
		"CREATE TABLE users (id INT, name VARCHAR(50));",
		"INSERT INTO users (id, name) VALUES (1, 'ahmed');",
	}

	database := NewDatabase()

	for _, query := range sql {
		tokens := tokenize(query)

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

		if err := database.execute(statement); err != nil {
			log.Fatalf("Execute error: %v\n", err)
			return
		}

		log.Println("============================================")
	}

	// verify
	for tableName, table := range database.tables {
		log.Printf("Table: %s\n", tableName)
		for _, col := range table.Schema {
			log.Printf("  Column: %-15s Type: %d  FixedSize: %d  MaxSize: %d  Length: %d\n",
				col.Name, col.Type, col.FixedSize, col.MaxSize, col.Length)
		}
		for i, row := range table.Rows {
			log.Printf("  Row %d:\n", i)
			for _, val := range row {
				log.Printf("    %-15s = %s\n", val.ColumnName, val.Raw)
			}
		}
	}

}
