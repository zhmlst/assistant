package migrations

import (
	"embed"
	"database/sql"
	"github.com/pressly/goose/v3"
	"fmt"
)

//go:embed *.sql
var fs embed.FS

func Up(db *sql.DB, dialect string) error {
	goose.SetBaseFS(fs)

	if err := goose.SetDialect(dialect); err !=nil {
		return fmt.Errorf("set dialect %s: %w", dialect, err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}

	return nil
}
