package main

import (
	"freshman-adaptation/database/sql"
	"freshman-adaptation/internal/config"
	"log"
	//"freshman-adaptation/internal/models"
)

func main() {
	cfg := config.Load()
	if err := sql.Migrate(cfg); err != nil {
		log.Fatal("Миграция не удалась:", err)
	}
}
