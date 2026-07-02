package sql

import (
	"database/sql"
	"fmt"
	"freshman-adaptation/internal/config"

	_ "github.com/lib/pq"
)

func Open(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB_USER,
		cfg.DB_PASS,
		cfg.DB_NAME,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
