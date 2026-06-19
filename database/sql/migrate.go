package cli

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func Migrate() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Нет .env файла, используем системные переменные")
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")
	if user == "" || pass == "" || dbname == "" {
		log.Fatal("Не заданы переменные окружения DB_USER, DB_PASS или DB_NAME")
	}

	// Подключаемся к стандартной БД postgres для создания нашей БД
	adminDSN := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=postgres sslmode=disable", user, pass)
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		log.Fatalf("Ошибка подключения к postgres: %v", err)
	}
	defer adminDB.Close()

	// Проверяем, существует ли БД с именем dbname
	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		log.Fatalf("Ошибка проверки существования БД: %v", err)
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
	}
	defer db.Close()

	// Инициализируем драйвер миграций
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Ошибка инициализации драйвера миграций: %v", err)
	}

	// Путь к папке с миграциями (относительно корня проекта)
	migrationsPath := "file://database/migrations"

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, dbname, driver)
	if err != nil {
		log.Fatalf("Ошибка создания объекта миграции: %v", err)
	}

	// Применяем все доступные up-миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка выполнения миграции: %v", err)
	}

	log.Println("Миграции успешно применены")
}
