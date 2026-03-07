package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

func (d *Database) execute(statement *Statement) ([][]Value, error) {
	tableName := statement.TableName
	switch statement.Type {
	case CreateStmnt:
		schema := make([]Column, len(statement.Columns))
		copy(schema, statement.Columns)
		d.tables[tableName] = Table{Schema: schema, Rows: [][]Value{}}
		return nil, nil

	case InsertStmnt:
		table, ok := d.tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table '%s' does not exist", tableName)
		}
		for _, row := range statement.Rows {
			var coercedRow []Value
			for _, val := range row {
				var col *Column
				for i := range table.Schema {
					if table.Schema[i].Name == val.ColumnName {
						col = &table.Schema[i]
						break
					}
				}
				if col == nil {
					return nil, fmt.Errorf("column '%s' does not exist in table '%s'", val.ColumnName, tableName)
				}

				typed, err := coerce(val, *col)
				if err != nil {
					return nil, err
				}

				coercedRow = append(coercedRow, Value{
					ColumnName: val.ColumnName,
					Raw:        val.Raw,
					TokenType:  val.TokenType,
					Typed:      typed,
				})
			}
			table.Rows = append(table.Rows, coercedRow)
		}
		d.tables[tableName] = table
		return nil, nil

	case SelectStmnt:
		table, ok := d.tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table '%s' does not exist", tableName)
		}

		if statement.SelectAll {
			return table.Rows, nil
		}

		var result [][]Value
		for _, row := range table.Rows {
			var filteredRow []Value
			for _, val := range row {
				for _, colName := range statement.SelectedColumns {
					if val.ColumnName == colName {
						filteredRow = append(filteredRow, val)
					}
				}
			}
			result = append(result, filteredRow)
		}
		return result, nil
	}

	return nil, nil
}

func coerce(val Value, col Column) (any, error) {
	switch col.Type {
	case IntType:
		n, err := strconv.ParseInt(val.Raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("column '%s': expected integer, got '%s'", col.Name, val.Raw)
		}
		return n, nil
	case FloatType:
		f, err := strconv.ParseFloat(val.Raw, 64)
		if err != nil {
			return nil, fmt.Errorf("column '%s': expected float, got '%s'", col.Name, val.Raw)
		}
		return f, nil
	case BoolType:
		switch strings.ToUpper(val.Raw) {
		case "TRUE", "1":
			return true, nil
		case "FALSE", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("column '%s': expected boolean, got '%s'", col.Name, val.Raw)
		}

	case StringType:
		if val.TokenType != StringLiteral {
			return nil, fmt.Errorf("column '%s': expected string, got '%s'", col.Name, val.Raw)
		}
		if col.NeedsLength && len(val.Raw) > col.Length {
			return nil, fmt.Errorf("column '%s': value exceeds max length %d", col.Name, col.Length)
		}
		return val.Raw, nil

	case BinaryType:
		return []byte(val.Raw), nil

	case TemporalType:
		t, err := time.Parse(time.DateTime, val.Raw)
		if err != nil {
			return nil, fmt.Errorf("column '%s': expected datetime, got '%s'", col.Name, val.Raw)
		}
		return t, nil
	default:
		return nil, fmt.Errorf("column '%s': unknown type", col.Name)
	}
}
