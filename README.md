
# Freshman Adaptation
> [!WARNING]
> deprecated ФАЙЛ, могжет содержать ошибки

Система адаптации первокурсников.

## Требования

Перед запуском должны быть установлены:

* Go 1.24+
* PostgreSQL 15+
* Git

## Клонирование проекта

```bash
git clone <repository-url>
cd Group1
```

## Настройка PostgreSQL

Создать пользователя PostgreSQL или использовать существующего.

Пример:

```sql
CREATE USER postgres WITH PASSWORD 'pass';
ALTER USER postgres CREATEDB;
```

## Переменные окружения

### Windows PowerShell

```powershell
$env:DB_USER="postgres"
$env:DB_PASS="pass"
$env:DB_NAME="freshman_adaptation"
```

### Linux/macOS

```bash
export DB_USER=postgres
export DB_PASS=pass
export DB_NAME=freshman_adaptation
```

## Установка зависимостей

```bash
go mod tidy
```

## Сборка приложения миграций

Из корня проекта:

```bash
go build -o migrate-app ./cmd/migrate
```

Windows:

```powershell
go build -o migrate-app.exe ./cmd/migrate
```

## Запуск миграций

Linux/macOS:

```bash
./migrate-app
```

Windows:

```powershell
.\migrate-app.exe
```

После успешного запуска будет:

```text
База данных freshman_adaptation создана
Миграции успешно применены
```

или

```text
База данных freshman_adaptation уже существует
Миграции успешно применены
```

## Структура базы данных

### groups

| Поле | Тип                   |
| ---- | --------------------- |
| id   | BIGSERIAL PRIMARY KEY |
| name | TEXT NOT NULL         |

### students

| Поле       | Тип                          |
| ---------- | ---------------------------- |
| id         | BIGSERIAL PRIMARY KEY        |
| first_name | TEXT NOT NULL                |
| last_name  | TEXT NOT NULL                |
| group_id   | BIGINT REFERENCES groups(id) |
| created_at | TIMESTAMPTZ                  |

### tasks

| Поле       | Тип                            |
| ---------- | ------------------------------ |
| id         | BIGSERIAL PRIMARY KEY          |
| student_id | BIGINT REFERENCES students(id) |
| title      | TEXT NOT NULL                  |
| status     | todo | in_progress | done      |
| created_at | TIMESTAMPTZ                    |

## Создание новой миграции

Создать файлы:

```text
migrations/
├── 000002_name.up.sql
└── 000002_name.down.sql
```

Пример:

```sql
-- 000002_add_email.up.sql
ALTER TABLE students ADD COLUMN email TEXT;
```

```sql
-- 000002_add_email.down.sql
ALTER TABLE students DROP COLUMN email;
```

После создания миграции снова выполнить:

```powershell
.\migrate-app.exe
```

или

```bash
./migrate-app
```

Новая миграция будет применена автоматически.

## Проверка таблиц

Подключиться к PostgreSQL:

```bash
psql -U postgres -d freshman_adaptation
```

Список таблиц:

```sql
\dt
```

Ожидаемый результат:

```text
groups
students
tasks
schema_migrations
```
