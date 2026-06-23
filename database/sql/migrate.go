package sql

import (
	"database/sql"
	"fmt"
	"freshman-adaptation/internal/config"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Migrate(cfg *config.Config) error {
	user := cfg.DB_USER
	pass := cfg.DB_PASS
	dbname := cfg.DB_NAME
	if user == "" || pass == "" || dbname == "" {
		log.Fatal("Не заданы переменные окружения DB_USER, DB_PASS или DB_NAME")
	}

	// Подключаемся к стандартной БД postgres для создания нашей БД
	adminDSN := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=postgres sslmode=disable", user, pass)
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		log.Fatalf("Ошибка подключения к postgres: %v", err)
		return err
	}
	defer adminDB.Close()

	// Проверяем, существует ли БД с именем dbname
	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		log.Fatalf("Ошибка проверки существования БД: %v", err)
		return err
	}

	if !exists {
		// Создаём БД
		_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			log.Fatalf("Не удалось создать БД: %v", err)
		}
		log.Printf("База данных %s создана", dbname)
	} else {
		log.Printf("База данных %s уже существует", dbname)
	}

	// Теперь подключаемся к нашей БД для миграций
	dsn := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", user, pass, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к %s: %v", dbname, err)
		return err
	}
	defer db.Close()

	// Инициализируем драйвер миграций
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Ошибка инициализации драйвера миграций: %v", err)
		return err
	}

	// Путь к папке с миграциями (относительно корня проекта)
	migrationsPath := "file://database/migrations"

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, dbname, driver)
	if err != nil {
		log.Fatalf("Ошибка создания объекта миграции: %v", err)
		return err
	}

	// Применяем все доступные up-миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка выполнения миграции: %v", err)
		return err
	}

	log.Println("Миграции успешно применены")
	return nil
}
