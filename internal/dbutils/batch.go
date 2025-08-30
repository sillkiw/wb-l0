package dbutils

import (
	"fmt"
	"strings"
)

// BuildBatchInsert формирует SQL-запрос для batch insert
func BuildBatchInsert(table string, columns []string, rows [][]interface{}) (string, []interface{}) {
	if len(rows) == 0 {
		return "", nil
	}

	query := fmt.Sprintf("INSERT INTO %s(%s) VALUES ",
		table,
		strings.Join(columns, ", "),
	)

	args := []interface{}{}
	placeholders := []string{}
	argIdx := 1

	for _, row := range rows {
		values := []string{}
		for range row {
			values = append(values, fmt.Sprintf("$%d", argIdx))
			argIdx++
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(values, ",")))
		args = append(args, row...)
	}

	query += strings.Join(placeholders, ", ")
	query += " ON CONFLICT DO NOTHING"

	return query, args
}
