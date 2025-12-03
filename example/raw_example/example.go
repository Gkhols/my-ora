package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Gkhols/my-ora/sqlrewrite/mysql/raw"
)

func main() {
	// Register the rewriting driver (only one line)
	raw.Register()

	// Connect to Oracle
	db, err := sql.Open("my-ora", `user/pass@host:port/db`)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Example queries
	examples := []string{
		"SELECT * FROM users LIMIT ? OFFSET ?",
		"SELECT * FROM users OFFSET ? LIMIT ?",
	}

	for _, q := range examples {
		rows, err := db.Query(q, 5, 0)
		if err != nil {
			fmt.Println("QUERY ERR: ", err)
		}

		columns, err := rows.Columns()
		if err != nil {
			log.Fatal(err)
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				log.Fatal(err)
			}

			rowMap := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					rowMap[col] = string(b)
				} else {
					rowMap[col] = val
				}
			}

			m, _ := json.Marshal(rowMap)

			fmt.Println(string(m))
		}
	}
}
