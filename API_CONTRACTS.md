# API Contracts — Задачи (Tasks)

Контракты JSON для эндпоинтов задач. Подробная спецификация — в `docs/API_SPEC.md`.

## Структура объекта задачи (Task)

### Response (GET /tasks/:taskId, GET /projects/:projectId/tasks, POST, PATCH)

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "key": "APP-1",
  "title": "string",
  "description": "string",
  "status": "TODO",
  "type": "TASK",
  "priority": "MEDIUM",
  "assignee_id": "uuid | null",
  "reporter_id": "uuid",
  "author_id": "uuid",
  "due_date": "2025-03-15T12:00:00Z | null",
  "tags": ["string"],
  "result_url": "string | null",
  "order": 0,
  "created_at": "2025-03-10T00:00:00Z",
  "updated_at": "2025-03-10T00:00:00Z"
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| id | string (UUID) | ID задачи |
| project_id | string (UUID) | ID проекта |
| key | string | Уникальный ключ (PROJECTKEY-N) |
| title | string | Заголовок (summary) |
| description | string | Описание |
| status | string | Ключ статуса из project_statuses (канбан-колонка) |
| type | string | BUG, TASK, STORY |
| priority | string | LOW, MEDIUM, HIGH, CRITICAL |
| assignee_id | string \| null | UUID исполнителя |
| reporter_id | string | UUID автора (reporter) |
| author_id | string | Алиас reporter_id (Jira-совместимость) |
| due_date | string \| null | ISO 8601 (RFC3339) |
| tags | string[] | Теги/метки |
| result_url | string \| null | Ссылка на результат |
| order | number | Позиция в колонке |
| created_at | string | ISO 8601 |
| updated_at | string | ISO 8601 |

### GET /projects/:projectId/tasks — фильтры (query)

| Параметр | Описание |
|----------|----------|
| status | Статус (ключ колонки) |
| assignee_id | UUID исполнителя |
| title | Подстрока заголовка (ILIKE, без учёта регистра) |
| type | BUG, TASK, STORY |
| due_from | RFC3339 — срок «от» (только задачи с `due_date`) |
| due_to | RFC3339 — срок «до» (только задачи с `due_date`) |

Те же параметры поддерживает **GET /projects/:projectId/board** (задачи в колонках доски фильтруются так же).

### Допустимые значения

| Поле | Значения |
|------|----------|
| status | TODO, IN_PROGRESS, IN_REVIEW, DONE (и кастомные из project_statuses) |
| type | BUG, TASK, STORY |
| priority | LOW, MEDIUM, HIGH, CRITICAL |

### Формат due_date

ISO 8601 / RFC3339: `2025-03-15T12:00:00Z` или `2025-03-15T15:00:00+03:00`.

---

## POST /projects/:projectId/tasks — Создать задачу

**Request:**
```json
{
  "title": "Реализовать API",
  "description": "Описание",
  "type": "TASK",
  "priority": "MEDIUM",
  "assignee_id": "uuid",
  "due_date": "2025-03-15T12:00:00Z",
  "tags": ["backend", "api"],
  "result_url": "https://example.com/result"
}
```

- `title` — обязательно.
- `author_id` / `reporter_id` — берётся из JWT (текущий пользователь), в запросе не передаётся.
- Остальные поля опциональны.

---

## PATCH /tasks/:taskId — Обновить задачу

**Request (все поля опциональны):**
```json
{
  "title": "string",
  "description": "string",
  "status": "IN_PROGRESS",
  "type": "BUG",
  "priority": "HIGH",
  "assignee_id": "uuid",
  "due_date": "2025-03-20T00:00:00Z",
  "tags": ["urgent"],
  "result_url": "https://example.com/patch",
  "order": 1
}
```

- Пустая строка `assignee_id` — сброс исполнителя.
- Пустой массив `tags` — очистка тегов.
