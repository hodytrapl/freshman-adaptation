package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const defaultUser = "postgres"
const defaultPass = "pass"
const defaultTestDB = "freshman_adaptation_test"

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Open возвращает подключение к тестовой БД или пропускает тест, если PostgreSQL недоступен.
func Open(t *testing.T) *sql.DB {
	t.Helper()

	user := envOrDefault("DB_USER", defaultUser)
	pass := envOrDefault("DB_PASS", defaultPass)
	dbName := envOrDefault("TEST_DB_NAME", defaultTestDB)

	if err := ensureDatabase(user, pass, dbName); err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable",
		user,
		pass,
		dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skipf("postgres unavailable: %v", err)
	}

	if err := applySchema(ctx, db); err != nil {
		db.Close()
		t.Fatalf("apply schema: %v", err)
	}

	if err := resetTables(ctx, db); err != nil {
		db.Close()
		t.Fatalf("reset tables: %v", err)
	}

	t.Cleanup(func() {
		_ = resetTables(context.Background(), db)
		_ = db.Close()
	})

	return db
}

func ensureDatabase(user, pass, dbName string) error {
	adminDSN := fmt.Sprintf(
		"host=localhost port=5432 user=%s password=%s dbname=postgres sslmode=disable",
		user,
		pass,
	)

	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer adminDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := adminDB.PingContext(ctx); err != nil {
		return err
	}

	var exists bool
	if err := adminDB.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName,
	).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return nil
	}

	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName)); err != nil {
		return err
	}

	return nil
}

func applySchema(ctx context.Context, db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS groups (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS students (
    id BIGSERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    group_id BIGINT NOT NULL REFERENCES groups(id),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo'
        CHECK (status IN ('todo', 'in_progress', 'done')),
    created_at TIMESTAMPTZ DEFAULT now()
);`

	_, err := db.ExecContext(ctx, schema)
	return err
}

func resetTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "TRUNCATE tasks, students, groups RESTART IDENTITY CASCADE")
	return err
}
