# Деплой на сервер через GitHub Actions

Деплой выполняется через **Docker Compose**: приложение, PostgreSQL и MinIO разворачиваются в контейнерах.

## Что нужно подготовить

### 1. На сервере

- **Docker и Docker Compose** — установлены и пользователь деплоя в группе `docker`
- **SSH-доступ** — пользователь с правами на деплой
- **Файл `.env`** — создать вручную в `DEPLOY_PATH` перед первым деплоем

Минимальный `.env` на сервере:

```env
JWT_SECRET=your-super-secret-key-min-32-chars
```

Остальные переменные (DB_*, MINIO_*) заданы в docker-compose и переопределяются при необходимости.

### 2. В GitHub

Добавь **Secrets** в репозитории: Settings → Secrets and variables → Actions:

| Secret | Описание |
|--------|----------|
| `DEPLOY_HOST` | IP или hostname сервера |
| `DEPLOY_USER` | SSH-пользователь |
| `DEPLOY_SSH_KEY` | Приватный SSH-ключ |
| `DEPLOY_PATH` | Путь на сервере (например `/home/user1/testik`) |

### 3. SSH-ключ

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/deploy_key -N ""
ssh-copy-id -i ~/.ssh/deploy_key.pub deploy@YOUR_SERVER
```

Содержимое `~/.ssh/deploy_key` → GitHub Secret `DEPLOY_SSH_KEY`.

---

## Что делает workflow

1. Копирует проект на сервер (rsync, без `.env`)
2. Запускает `docker compose up -d --build` — сборка образов и запуск app, db, minio

---

## Стек в docker-compose

| Сервис | Порт | Описание |
|--------|------|----------|
| app | 8080 | API (Go) |
| db | 5432 | PostgreSQL 16 |
| minio | 9000, 9001 | MinIO (API + консоль) |

---

## Локальный запуск

```bash
cp .env.example .env
# Заполни JWT_SECRET в .env
docker compose up -d --build
```

API: http://localhost:8080  
MinIO Console: http://localhost:9001 (minioadmin:minioadmin)

---

## Проверка

После `git push origin main`:

1. GitHub Actions → вкладка Actions
2. Workflow "Deploy"
3. Проверь логи на ошибки
