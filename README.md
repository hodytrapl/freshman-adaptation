# Адаптация студентов — REST API (MVP)

## Запуск под Windows 10

### 1. Установите Go (версия 1.21+)
Скачайте с [golang.org](https://golang.org/dl/).

### 2. Установите PostgreSQL
Скачайте и установите [PostgreSQL](https://www.postgresql.org/download/windows/).  
Запомните пароль суперпользователя (например, `pass`).

### 3. Установите зависимости
В корне проекта выполните:
```bash
go mod tidy
```
### 4. Настройте переменные окружения
Задайте переменные в командной строке (Windows):

```cmd
set DB_USER=postgres
set DB_PASS=your_password
set DB_NAME=freshman_adaptation
set HTTP_PORT=8080
```

Если используете PowerShell, то:

```powershell
$env:DB_USER="postgres"
$env:DB_PASS="your_password"
$env:DB_NAME="freshman_adaptation"
$env:HTTP_PORT="8080"
```

### 5. Примените миграции
При первом запуске программа автоматически создаст базу данных и накатит миграции.
Просто запустите сервер.

### 6. Запустите сервер
```bash
go run cmd/app/main.go  
```
Сервер будет доступен по адресу http://localhost:8080.

### 7. Тестирование
Используйте файл ```requests.http``` (или cURL) для проверки всех эндпоинтов.
> [!IMPORTANT]
> находится в папке docs

# 8. Тестирование транзакции (откат)
Чтобы убедиться, что транзакция работает, можно временно модифицировать код в ```CreateDefaultTasks``` (например, заменить ```title``` на несуществующую колонку) и убедиться, что при ошибке студент не создаётся.