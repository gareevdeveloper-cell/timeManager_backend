# Архитектура backend

## Стек

- HTTP: gin
- БД: PostgreSQL, pgx
- JWT: golang-jwt/jwt
- Пароли: bcrypt (golang.org/x/crypto)

## Общая структура

Репозиторий организован по принципам чистой / гексагональной архитектуры. Используется **feature-based** подход: каждая фича — отдельный модуль в `internal/<feature>`.

- `cmd/api` — точка входа (загрузка конфига, создание App, Run).
- `internal/app` — инициализация приложения (БД, миграции, репозитории, сервисы, роутер, запуск сервера).
- `internal/migrate` — миграции БД (выполняются автоматически при старте).
- `internal/config` — конфигурация приложения (env).
- `pkg` — переиспользуемые утилиты: `pkg/response` (JSON-ответы API), `pkg/env` (переменные окружения).
- `internal/domain` — общие доменные модели (User и др.).
- `internal/http` — HTTP‑слой: `router.go` (сборка роутера), `auth_routes.go` (маршруты auth) и др. по типам.
- `internal/http/middleware` — общие middleware (auth, logging, recovery).
- `internal/<feature>` — feature‑модули:
  - `handler.go` — HTTP‑обработчики (парсинг, вызов сервиса, маппинг ответов).
  - `service.go` — бизнес‑логика (use‑cases).
  - `repository.go` — интерфейс доступа к данным.
  - `repository_postgres.go` — реализация для PostgreSQL.
  - `dto.go` — DTO для запросов/ответов.
  - `errors.go` — доменные ошибки фичи.

Пример: `internal/auth` — модуль аутентификации (регистрация, логин, JWT).

Правило зависимостей:

- `domain` не зависит ни от чего.
- `internal/<feature>/service` зависит только от `domain` + интерфейса репозитория.
- `internal/<feature>/repository` зависит от pgx и т.п., но не от `http`.
- `internal/<feature>/handler` зависит от сервиса и domain, не от repository напрямую.
- `internal/http/middleware` — общий слой, может зависеть от интерфейсов (например, TokenProvider).

## Добавление новой фичи

1. Создать `internal/<feature>` (например, `internal/organizations`).
2. Добавить `handler.go`, `service.go`, `repository.go`, `repository_postgres.go`, `dto.go`, `errors.go`.
3. Доменные модели — в `internal/domain` или в `internal/<feature>`, если сущность локальна для фичи.
4. Зарегистрировать роуты в `internal/http/<feature>_routes.go` и DI в `internal/app/app.go`.

## API

- REST API, формат JSON.
- Документация: Swagger UI (`/swagger/index.html`). При изменении API — обновить аннотации и выполнить `make swagger`.
- Версионирование: `/api/v1/...`.
- Ответы:
  - успешный ответ: `{ "data": ..., "meta": ... }`
  - ошибка: `{ "error": { "code": "string", "message": "string", "details": ... } }`
- Конвенции HTTP‑кодов:
  - 200/201 — успех
  - 400 — ошибка валидации/формата
  - 401 — неавторизован
  - 403 — нет доступа
  - 404 — не найдено
  - 500 — внутренняя ошибка

## Ошибки и логирование

- Все ошибки из слоёв `service` и `repository` оборачиваются с контекстом.
- HTTP‑слой мапит ошибки на HTTP‑коды и возвращает стандартизованный JSON.
- Логи:
  - `INFO` — бизнес‑события (создание сущностей и т.п.).
  - `ERROR` — ошибки с полным stack/контекстом.

## Аутентификация

- Регистрация: `POST /api/v1/auth/register` (email, password)
- Логин: `POST /api/v1/auth/login` → access_token (JWT)
- OAuth: `GET /api/v1/auth/{google,yandex,github}/redirect` → редирект на провайдера; callback возвращает JWT
- Защищённые эндпоинты: заголовок `Authorization: Bearer <token>`
- Middleware `internal/http/middleware.Auth` проверяет JWT и кладёт `user_id` в gin.Context (`c.Get("user_id")`)
- Подробнее: `docs/AUTH_ARCHITECTURE.md`

## Тестирование

- Unit‑тесты для:
  - доменных моделей (`internal/domain`)
  - сервисов (`internal/<feature>/service`) с моками репозиториев.
- Интеграционные тесты HTTP (по возможности).
- Запуск: `go test ./...`.

Агент при изменении архитектуры обязан следовать этим правилам и не ломать слои.
