# timeManager Backend

Backend-сервис для работы с командой над проектом. Позволяет состоять в организациях и командах, отмечаться на рабочем месте, устанавливать статусы (отпуск/больничный), общаться в чате проекта и тет-а-тет, работать с доской задач.

## Стек

- Go 1.26 — чистая архитектура
- HTTP: gin
- БД: PostgreSQL pgx + hand‑written queries
- Конфигурация: Viper (env + YAML)
- Логирование: zerolog
- Контейнеризация: Docker + docker-compose

## Основные возможности

- Аутентификация и авторизация (JWT).
- CRUD по сущностям: организации, команды, пользователи, статусы, задачи, чаты.
- Валидация входных данных (validator.v10).
- Стандартизированный формат ошибок (JSON API).
- Чат: проектный чат + личные переписки (WebSocket)
- Доска задач: Kanban-style с drag-n-drop (frontend API)
- Статусы: на рабочем месте, отпуск, больничный


## Локальный запуск

### Вариант 1: Docker Compose (рекомендуется)

```bash
cp .env.example .env
# Заполни JWT_SECRET в .env (минимум 32 символа)
docker compose up -d --build
```

Запускает API (порт 8080), PostgreSQL (5432) и MinIO (9000, консоль 9001).

### Вариант 2: Только БД + Go

```bash
cp .env.example .env
# Заполни JWT_SECRET, MINIO_* в .env
docker compose up -d db minio
go run ./cmd/api
```

- Файл `.env` загружается автоматически при запуске (godotenv).
- Миграции выполняются автоматически при запуске сервера.

## Команды

```bash
# Запуск сервера
go run ./cmd/api

# Тесты
go test ./...

# Линтер
golangci-lint run ./...

# Генерация Swagger (после изменения API)
make swagger
```

## Swagger

После запуска сервера документация API доступна по адресу: http://localhost:8080/swagger/index.html

## API для фронтенда

Спецификация API для интеграции с фронтендом: [docs/API_SPEC.md](docs/API_SPEC.md)
