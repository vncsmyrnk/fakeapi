package sqlite

import (
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations.sql
var initialMigrationsStmt string

// InitDB applies the initial migrations to the database.
func InitDB(db *sqlx.DB) error {
	_, err := db.Exec(initialMigrationsStmt)
	if err != nil {
		return fmt.Errorf("apply initial migrations: %w", err)
	}
	return nil
}
