package config

import "os"

type Config struct {
	Port string
	DB_USER string
	DB_PASS string
	DB_NAME string
}

func Load() *Config {
	/*если пользователь не ввёл переменные по умолчанию, то заносим это в код
	манал я это всё вводить для проверки каждый раз*/
	cfg := &Config{
		Port: "8080", // Порт по умолчанию
		DB_USER: "postgres", //по умолчанию
		DB_PASS: "pass", //по умолчанию
		DB_NAME :"freshman_adaptation", //по умолчанию
	}

	if port := os.Getenv("HTTP_PORT"); port != "" {
		cfg.Port = port
	}
	if db_user := os.Getenv("DB_USER"); db_user != "" {
		cfg.DB_USER = db_user
	}
	if db_pass := os.Getenv("DB_PASS"); db_pass != "" {
		cfg.DB_PASS = db_pass
	}
	if db_name := os.Getenv("DB_NAME"); db_name != "" {
		cfg.DB_NAME = db_name
	}

	return cfg
}
