package orm

import (
	"time"

	"github.com/oracle-samples/gorm-oracle/oracle"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Initialize – register driver, connect, attach callbacks
func Initialize(dsn string) (*gorm.DB, error) {
	// Open connection using your custom GORM dialect
	db, err := gorm.Open(oracle.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		PrepareStmt:        true,
		PrepareStmtMaxSize: 100,
		PrepareStmtTTL:     time.Hour,
	})
	if err != nil {
		return db, err
	}
	return db, err
}
