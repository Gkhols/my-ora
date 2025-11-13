package main

import (
	"fmt"

	"github.com/ghiant/my-ora/sqlrewrite/mysql/orm"
)

func main() {
	// Example DSN (adjust to your own Oracle connection)
	dsn := "user=user password=password connectString=host:port/pdb"

	// Open connection using your custom GORM dialect
	db, err := orm.Initialize(dsn)
	if err != nil {
		return
	}

	// Just to confirm connection
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("failed to get sql.DB: %v", err)
		return
	}

	if err = sqlDB.Ping(); err != nil {
		fmt.Println("database not reachable: %v", err)
		return
	}
	fmt.Println("✅ Connected to Oracle using MySQL-compatible syntax!")

	// Example query (MySQL-like)
	var results []map[string]interface{}
	tx := orm.RawWithRewriter(db, "SELECT * FROM USERS LIMIT ? OFFSET ?", 5, 0).Scan(&results)
	if tx.Error != nil {
		fmt.Println("query error: %v", tx.Error)
		return
	}
	var results2 []map[string]interface{}
	tx = db.Debug().Table("USERS").Limit(0).Offset(5).Find(&results2)
	if tx.Error != nil {
		fmt.Println("query error: %v", tx.Error)
		return
	}

	// Print results
	fmt.Printf("✅ Query 1 OK, rows: %d\n", len(results))
	fmt.Printf("✅ Query 2 OK, rows: %d\n", len(results2))
}
