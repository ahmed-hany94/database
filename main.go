package main

import (
	"log"
)

/**
* Journey
*   Basic
*     [X] CREATE Statement
*     [X] INSERT Statement
*     [X] SELECT Statement
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
		"CREATE TABLE users (id INT, name VARCHAR(50), is_admin BOOL);",
		"INSERT INTO users (id, name, is_admin) VALUES (1, 'ahmed', TRUE), (2, 'amr', FALSE);",
		"SELECT * FROM users;",
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

		rows, err := database.execute(statement)
		if err != nil {
			log.Fatalf("Execute error: %v\n", err)
			return
		}

		if rows != nil {
			log.Println("Results:")
			for i, row := range rows {
				log.Printf("  Row %d:\n", i)
				for _, val := range row {
					log.Printf("    %-15s = %-15s (typed: %v)\n", val.ColumnName, val.Raw, val.Typed)
				}
			}
		}

		log.Println("============================================")
	}
}
