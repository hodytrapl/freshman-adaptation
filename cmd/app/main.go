package main

import (
	"fmt"
	"freshman-adaptation/database/sql"
	"freshman-adaptation/internal/api"
	"freshman-adaptation/internal/config"
	"freshman-adaptation/internal/middleware"
	"freshman-adaptation/internal/repository"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	if err := sql.Migrate(cfg); err != nil {
		log.Fatal("Миграция не удалась:", err)
	}

	db, err := sql.Open(cfg)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}
	defer db.Close()

	repo := repository.New(db)
	handler := api.NewHandler(repo)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: middleware.RequestLogger(middleware.Recover(mux)),
	}

	log.Printf("Сервер запущен на http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
