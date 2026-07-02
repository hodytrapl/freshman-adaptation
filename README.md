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

Если используете PowerShell:

```powershell
$env:DB_USER="postgres"
$env:DB_PASS="your_password"
$env:DB_NAME="freshman_adaptation"
$env:HTTP_PORT="8080"
```

Если переменные не заданы, используются значения по умолчанию из `internal/config/config.go`:
- `DB_USER=postgres`
- `DB_PASS=pass`
- `DB_NAME=freshman_adaptation`
- `HTTP_PORT=8080`

### 5. Примените миграции
При первом запуске программа автоматически создаст базу данных и накатит миграции.  
Просто запустите сервер.

### 6. Запустите сервер

```bash
go run cmd/app/main.go
```

Сервер будет доступен по адресу http://localhost:8080.

Проверка жизнеспособности:

```bash
curl http://localhost:8080/health
```

### 7. Тестирование
Используйте файл `docs/requests.http` (REST Client в VS Code / Cursor) или cURL из `docs/API.md` для проверки всех эндпоинтов.

### 8. Тестирование транзакции (откат)
Чтобы убедиться, что транзакция работает:

1. Временно измените функцию `createDefaultTasks` в `internal/repository/repository.go` — замените колонку `title` на несуществующую (например, `titl`).
2. Перезапустите сервер.
3. Выполните `POST /students` с существующим `group_id`.
4. Убедитесь, что ответ — `500`, а студент **не** появился в `GET /students?group_id=1`.
5. Верните код в исходное состояние и перезапустите сервер.

Подробные cURL-примеры и сценарии ошибок — в `docs/requests.http` и `docs/API.md`.

## Структура проекта

```
cmd/app/main.go              — точка входа (миграции + HTTP-сервер)
database/migrations/         — SQL-миграции
database/sql/                — подключение к БД и migrate
internal/api/                — HTTP-хендлеры и DTO запросов
internal/apperrors/          — sentinel-ошибки и AppError
internal/config/             — конфигурация из env
internal/middleware/         — JSON-ответы, логирование, recover
internal/models/             — модели данных
internal/repository/         — работа с PostgreSQL (в т.ч. транзакции)
docs/API.md                  — документация для фронтенда
docs/requests.http           — сценарии запросов
```

## Обработка ошибок

Все ошибки API возвращаются в едином формате:

```json
{
  "error": "сообщение"
}
```

Sentinel-ошибки (`ErrNotFound`, `ErrValidation`, `ErrInternal`) преобразуются в HTTP-статусы `404`, `400`, `500` соответственно.

## Тесты

```bash
go test ./... -v
```

- **Unit-тесты** (apperrors, middleware, API с sqlmock) — работают без PostgreSQL.
- **Интеграционные тесты** (`TestAPIFlowIntegration`, `internal/repository/*`) — требуют запущенный PostgreSQL.

Переменные для интеграционных тестов (необязательно):

```powershell
$env:DB_USER="postgres"
$env:DB_PASS="your_password"
$env:TEST_DB_NAME="freshman_adaptation_test"
```
