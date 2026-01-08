package main

import "fmt"

func main() {
	sql := "CREATE TABLE users (id INT, name VARCHAR(50));"
	tokens := tokenize(sql)

	for _, tok := range tokens {
		log.Printf("%-15s %-12s at position %d\n",
			tok.Content,
			tokenTypeName(tok.Type),
			tok.Position)
	}
}
