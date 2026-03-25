# API Specification

Спецификация REST API для фронтенд-приложения timeManager. Все эндпоинты возвращают JSON.

**Base URL:** `http://localhost:8080/api/v1`

**Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## Аутентификация

Защищённые эндпоинты требуют заголовок:

```
Authorization: Bearer <access_token>
```

Токен получается при логине (`POST /auth/login`) или через OAuth (Google, Yandex, GitHub).

---

## Формат ответов

### Успешный ответ (200, 201)

```json
{
  "data": { ... }
}
```

### Ошибка (4xx, 5xx)

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

### HTTP-коды

| Код | Описание |
|-----|----------|
| 200 | OK |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request — ошибка валидации |
| 401 | Unauthorized — не авторизован |
| 403 | Forbidden — нет доступа |
| 404 | Not Found |
| 409 | Conflict — конфликт (например, дубликат) |
| 500 | Internal Server Error |

---

## Эндпоинты

### Files (публичные)

#### GET /files/{path} — Получить файл из хранилища

Прокси для картинок из MinIO. Принимает частичный путь к файлу и возвращает содержимое.

**Примеры путей:**
- `users/{user_id}/avatar.jpg` — аватар пользователя
- `organizations/{org_id}/avatar.jpg` — аватар организации
- `teams/{team_id}/avatar.jpg` — аватар команды

**Пример:** `GET /api/v1/files/users/550e8400-e29b-41d4-a716-446655440000/avatar.jpg`

**Response 200:** бинарное содержимое файла с заголовком Content-Type (image/jpeg, image/png и т.д.)

**Ошибки:** 400 (validation_error), 403 (forbidden — путь не разрешён), 404 (not_found)

---

### Auth (публичные)

#### POST /auth/register — Регистрация

Создаёт нового пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| email | string | да | Валидный email |
| password | string | да | Минимум 8 символов |

**Response 201:**
```json
{
  "data": {
    "user_id": "uuid",
    "email": "user@example.com"
  }
}
```

**Ошибки:** 400 (validation_error), 409 (user_exists)

---

#### POST /auth/login — Вход

Проверяет учётные данные и возвращает JWT access token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

**Ошибки:** 400 (validation_error), 401 (invalid_credentials)

---

#### OAuth — вход через Google, Yandex, GitHub

Поддерживаются провайдеры: `google`, `yandex`, `github`.

**Шаг 1.** Редирект на провайдера:
```
GET /auth/{provider}/redirect
```
Пример: `GET /auth/google/redirect` — редирект на страницу авторизации Google.

**Шаг 2.** Callback (обрабатывается автоматически):
```
GET /auth/{provider}/callback?code=...&state=...
```
После успешной авторизации — редирект на `OAUTH_FRONTEND_REDIRECT` с параметрами:
- `token` — JWT access token
- `expires_in` — время жизни токена в секундах

При ошибке — редирект с параметром `error`:
- `invalid_provider` — неверный провайдер
- `missing_code` — отсутствует code
- `invalid_state` — неверный state (CSRF)
- `email_already_registered` — email уже зарегистрирован с паролем
- `user_inactive` — пользователь неактивен
- `oauth_failed` — ошибка OAuth

**Переменные окружения:**
- `OAUTH_REDIRECT_BASE` — базовый URL API (для callback)
- `OAUTH_FRONTEND_REDIRECT` — URL фронтенда для редиректа с токеном
- `OAUTH_GOOGLE_CLIENT_ID`, `OAUTH_GOOGLE_CLIENT_SECRET`
- `OAUTH_YANDEX_CLIENT_ID`, `OAUTH_YANDEX_CLIENT_SECRET`
- `OAUTH_GITHUB_CLIENT_ID`, `OAUTH_GITHUB_CLIENT_SECRET`

---

### Auth (защищённые)

#### GET /users/me — Текущий пользователь

Возвращает данные авторизованного пользователя.

**Headers:** `Authorization: Bearer <token>`

**Response 200:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "firstname": "string",
    "lastname": "string",
    "birthday": "1990-05-15",
    "about": "О себе",
    "position": "Backend Developer",
    "skills": ["Go", "PostgreSQL", "Docker"],
    "role": "user",
    "status": "active",
    "work_status": "working",
    "current_task_id": "uuid",
    "avatar_url": "/api/v1/files/users/uuid/avatar.jpg",
    "created_at": "2025-03-09T00:00:00Z"
  }
}
```

`current_task_id` — UUID задачи «в работе», если задана; иначе поле можно не возвращать.

**work_status:** working | resting | lunch | vacation | sick_leave | business_trip

**Ошибки:** 401 (unauthorized), 404 (not_found)

---

#### PUT /users/me/work-status — Обновить рабочий статус

Обновляет рабочий статус сотрудника. Запись добавляется в историю.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "work_status": "lunch"
}
```

| work_status | Описание |
|-------------|----------|
| working | работает |
| resting | отдыхает |
| lunch | обед |
| vacation | отпуск |
| sick_leave | больничный |
| business_trip | командировка |

**Response 200:** данные пользователя с обновлённым work_status.

**Ошибки:** 400 (validation_error), 401 (unauthorized), 500 (internal_error)

---

#### GET /users/me/work-status/history — История изменений статуса

Возвращает историю изменений рабочего статуса текущего пользователя.

**Headers:** `Authorization: Bearer <token>`

**Query:** `limit` (опционально, по умолчанию 50, макс 100)

**Response 200:**
```json
{
  "data": {
    "history": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "work_status": "lunch",
        "changed_at": "2025-03-09T12:00:00Z",
        "changed_by": "uuid"
      }
    ]
  }
}
```

---

#### PATCH /users/me — Обновить профиль

Обновляет поля профиля: firstname, lastname, birthday, about (о себе), position (должность), skills (скиллы). Все поля опциональны — при отсутствии сохраняется текущее значение. Скиллы создаются в БД при первом добавлении.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "firstname": "Иван",
  "lastname": "Иванов",
  "birthday": "1990-05-15",
  "about": "О себе",
  "position": "Backend Developer",
  "skills": ["Go", "PostgreSQL", "Docker"]
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| firstname | string | нет | Имя |
| lastname | string | нет | Фамилия |
| birthday | string | нет | Дата рождения: `YYYY-MM-DD` или RFC3339; пустая строка `""` — сбросить дату |
| about | string | нет | О себе |
| position | string | нет | Должность |
| skills | string[] | нет | Список скиллов (создаются в БД при отсутствии) |

**Response 200:** данные пользователя с обновлённым профилем.

**Ошибки:** 400 (validation_error), 401 (unauthorized), 500 (internal_error)

---

#### GET /users/me/tasks — Мои задачи (исполнитель)

Возвращает задачи, где текущий пользователь назначен исполнителем (`assignee_id`). У каждой записи поле **`in_work`**: `true` только у той задачи, которая выбрана как текущая «в работе» (не больше одной; совпадает с `current_task_id` в профиле).

**Response 200:**
```json
{
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "project_id": "uuid",
        "key": "APP-1",
        "title": "...",
        "in_work": true,
        "...": "остальные поля задачи как в проекте"
      }
    ]
  }
}
```

**Ошибки:** 401 (unauthorized), 500 (internal_error)

---

#### PUT /users/me/current-task — Текущая задача в работе

Устанавливает одну задачу как «текущую в работе». Задача должна существовать и быть **назначена на пользователя** (исполнитель). Одновременно может быть только одна такая задача; повторный вызов с другим `task_id` заменяет выбор.

**Request:**
```json
{ "task_id": "550e8400-e29b-41d4-a716-446655440000" }
```

Сброс (ничего не «в работе»):
```json
{ "task_id": null }
```
или `"task_id": ""`.

| Поле | Описание |
|------|----------|
| task_id | UUID задачи; `null` или пустая строка — сброс |

**Response 200:** данные пользователя (как GET /users/me), с обновлённым `current_task_id`.

**Ошибки:** 400 (validation_error), 401 (unauthorized), 403 (forbidden — задача не назначена на вас), 404 (not_found — задача не найдена), 500 (internal_error)

---

#### PUT /users/me/avatar — Установить аватарку текущего пользователя

**Request:** multipart/form-data с полем `avatar` (jpeg/png/webp/gif, max 5MB).

**Response 200:** данные пользователя с обновлённым avatar_url.

**Ошибки:** 400 (validation_error), 401 (unauthorized)

---

### Organizations (защищённые)

Все эндпоинты организаций требуют `Authorization: Bearer <token>`.

#### GET /organizations — Список моих организаций

Возвращает организации, в которых текущий пользователь является участником.

**Response 200:**
```json
{
  "data": {
    "organizations": [
      {
        "id": "uuid",
        "name": "My Company",
        "slug": "my-company",
        "owner_id": "uuid",
        "status": "active",
        "avatar_url": "/api/v1/files/organizations/uuid/avatar.jpg",
        "created_at": "2025-03-09T00:00:00Z",
        "updated_at": "2025-03-09T00:00:00Z"
      }
    ]
  }
}
```

**Ошибки:** 401 (unauthorized)

---

#### GET /organizations/:id — Получить организацию по ID

Возвращает организацию. Доступ только для членов организации.

**Response 200:** данные организации (как в списке выше).

**Ошибки:** 401 (unauthorized), 404 (not_found)

---

#### GET /organizations/:id/members — Список участников организации

Возвращает всех пользователей, входящих в организацию. Доступ только для членов организации.

У каждого участника: **`work_status`** (working | resting | lunch | vacation | sick_leave | business_trip), **`current_task_id`** (UUID или `null`) и **`current_task`** — объект `{ "id", "title", "project_id" }` или `null` — текущая задача «в работе» (для канбана/канваса без отдельного `GET /tasks/:id`).

**Response 200:**
```json
{
  "data": {
    "members": [
      {
        "id": "uuid",
        "email": "user@example.com",
        "firstname": "string",
        "lastname": "string",
        "middlename": "string",
        "role": "administrator",
        "work_status": "working",
        "avatar_url": "/api/v1/files/users/uuid/avatar.jpg",
        "current_task_id": "uuid",
        "current_task": {
          "id": "uuid",
          "title": "Задача в работе",
          "project_id": "uuid"
        }
      }
    ]
  }
}
```

**role:** administrator | participant | user

**work_status:** working | resting | lunch | vacation | sick_leave | business_trip

**Ошибки:** 401 (unauthorized), 404 (not_found)

---

#### POST /organizations — Создать организацию

Поддерживает JSON или multipart/form-data (name + опционально avatar).

**Request (JSON):**
```json
{
  "name": "My Company"
}
```

**Request (multipart):** поля `name` (обязательно), `avatar` (опционально, файл jpeg/png/webp/gif, max 5MB).

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| name | string | да | 1–255 символов |
| avatar | file | нет | Изображение (jpeg/png/webp/gif, max 5MB) |

**Response 201:**
```json
{
  "data": {
    "id": "uuid",
    "name": "My Company",
    "slug": "my-company",
    "owner_id": "uuid",
    "status": "active",
    "avatar_url": "/api/v1/files/organizations/uuid/avatar.jpg",
    "created_at": "2025-03-09T00:00:00Z",
    "updated_at": "2025-03-09T00:00:00Z"
  }
}
```

---

#### PATCH /organizations/:id — Обновить организацию

Поддерживает JSON или multipart (name + опционально avatar). Slug остаётся неизменным.

**Request (JSON):**
```json
{
  "name": "My Company Updated"
}
```

**Request (multipart):** поля `name` (обязательно), `avatar` (опционально).

**Response 200:** как у Create, с обновлёнными данными.

**Ошибки:** 403 (archived), 404 (not_found)

---

#### PUT /organizations/:id/avatar — Установить аватарку

Загружает аватарку в MinIO. Только член организации может изменить аватарку.

**Request:** multipart/form-data с полем `avatar` (файл jpeg/png/webp/gif, max 5MB).

**Response 200:** данные организации с обновлённым avatar_url.

**Ошибки:** 400 (validation_error), 403 (archived), 404 (not_found)

---

#### POST /organizations/:id/archive — Архивировать организацию

**Response 200:** данные организации со статусом `archived`.

---

#### POST /organizations/:id/members — Добавить участника

Пользователь может быть только в одной организации. Роль по умолчанию — participant.

**Request:**
```json
{
  "user_id": "uuid",
  "role": "participant"
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| user_id | string (UUID) | да | ID пользователя |
| role | string | нет | administrator, participant, user (по умолчанию participant) |

**Response 204:** No Content

**Ошибки:** 403 (archived), 404 (not_found), 409 (user_already_in_org)

---

#### DELETE /organizations/:id/members/:user_id — Удалить участника

**Response 204:** No Content

**Ошибки:** 403 (archived), 404 (not_found)

---

### Teams (защищённые)

Все эндпоинты команд требуют `Authorization: Bearer <token>`.

#### POST /teams — Создать команду

Создатель должен быть членом организации. Поддерживает JSON или multipart (name, description, organization_id + опционально avatar).

**Request (JSON):**
```json
{
  "name": "Backend Team",
  "description": "Команда разработки бэкенда",
  "organization_id": "uuid"
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| name | string | да | 1–255 символов |
| description | string | нет | Описание команды |
| organization_id | string | да | UUID организации |
| avatar | file | нет | Изображение (jpeg/png/webp/gif, max 5MB) |

**Response 201:**
```json
{
  "data": {
    "id": "uuid",
    "name": "Backend Team",
    "description": "Команда разработки бэкенда",
    "organization_id": "uuid",
    "creator_id": "uuid",
    "avatar_url": "/api/v1/files/teams/uuid/avatar.jpg",
    "created_at": "2025-03-09T00:00:00Z"
  }
}
```

**Ошибки:** 403 (forbidden — не в организации), 404 (not_found)

---

#### GET /teams/:id — Получить команду

**Response 200:** данные команды (как в Create, с avatar_url при наличии).

---

#### PATCH /teams/:id — Обновить команду

Поддерживает JSON или multipart (name, description + опционально avatar).

**Request (JSON):**
```json
{
  "name": "Backend Team Updated",
  "description": "Обновлённое описание"
}
```

**Response 200:** данные команды.

---

#### PUT /teams/:id/avatar — Установить аватарку команды

Только член организации может изменить аватарку.

**Request:** multipart/form-data с полем `avatar`.

**Response 200:** данные команды с обновлённым avatar_url.

**Ошибки:** 400 (validation_error), 404 (not_found)

---

#### DELETE /teams/:id — Удалить команду

**Response 204:** No Content

---

#### POST /teams/:id/members — Добавить участника в команду

Добавлять можно только пользователей, состоящих в организации команды. Роль по умолчанию — participant.

**Request:**
```json
{
  "user_id": "uuid",
  "role": "participant"
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| user_id | string (UUID) | да | ID пользователя |
| role | string | нет | administrator, participant, user (по умолчанию participant) |

**Response 204:** No Content

**Ошибки:** 403 (forbidden — пользователь не в организации), 404 (not_found), 409 (user_already_in_team)

---

#### DELETE /teams/:id/members/:user_id — Удалить участника из команды

**Response 204:** No Content

**Ошибки:** 404 (not_found)

---

#### GET /teams/:id/projects — Проекты команды

Возвращает проекты команды. Только для членов команды.

**Response 200:**
```json
{
  "data": {
    "projects": [
      {
        "id": "uuid",
        "key": "APP",
        "name": "My Application",
        "description": "...",
        "team_id": "uuid",
        "owner_id": "uuid",
        "created_at": "2025-03-10T00:00:00Z",
        "updated_at": "2025-03-10T00:00:00Z"
      }
    ]
  }
}
```

**Ошибки:** 403 (forbidden), 404 (not_found)

---

#### GET /teams/:id/members — Список участников команды

У каждого участника: **`work_status`**, **`current_task_id`** и **`current_task`** `{ id, title, project_id }` — как у списка участников организации.

**Response 200:**
```json
{
  "data": {
    "members": [
      {
        "id": "uuid",
        "email": "user@example.com",
        "firstname": "string",
        "lastname": "string",
        "middlename": "string",
        "role": "administrator",
        "work_status": "working",
        "avatar_url": "/api/v1/files/users/uuid/avatar.jpg",
        "current_task_id": "uuid",
        "current_task": {
          "id": "uuid",
          "title": "Задача в работе",
          "project_id": "uuid"
        }
      }
    ]
  }
}
```

**role:** administrator | participant | user

**work_status:** working | resting | lunch | vacation | sick_leave | business_trip

---

#### GET /organizations/:id/teams — Команды организации

**Response 200:**
```json
{
  "data": {
    "teams": [
      {
        "id": "uuid",
        "name": "string",
        "description": "string",
        "organization_id": "uuid",
        "creator_id": "uuid",
        "avatar_url": "/api/v1/files/teams/uuid/avatar.jpg",
        "created_at": "2025-03-09T00:00:00Z"
      }
    ]
  }
}
```

---

## Projects (защищённые)

Все эндпоинты проектов и задач требуют `Authorization: Bearer <token>`.
Доступ имеют только владельцы проектов.

#### POST /projects — Создать проект

Проект может быть создан с привязкой к команде (team_id) или без — личный проект.

**Request:**
```json
{
  "team_id": "550e8400-e29b-41d4-a716-446655440000",
  "key": "APP",
  "name": "My Application",
  "description": "Описание проекта"
}
```

Без team_id — личный проект:
```json
{
  "key": "APP",
  "name": "My Application",
  "description": "Описание проекта"
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| team_id | string (UUID) | нет | ID команды (если указан — создатель должен быть членом команды) |
| key | string | да | Уникальный короткий код (1–50 символов, A-Z0-9, например APP) |
| name | string | да | 1–255 символов |
| description | string | нет | Описание проекта |

**Response 201:**
```json
{
  "data": {
    "id": "uuid",
    "key": "APP",
    "name": "My Application",
    "description": "Описание проекта",
    "team_id": "uuid",
    "owner_id": "uuid",
    "created_at": "2025-03-10T00:00:00Z",
    "updated_at": "2025-03-10T00:00:00Z"
  }
}
```

**Ошибки:** 400 (validation_error), 403 (forbidden — пользователь не в команде), 404 (not_found — команда не найдена), 409 (key_exists)

---

#### GET /projects — Список проектов

Возвращает проекты, доступные пользователю (владелец или член команды).

**Response 200:**
```json
{
  "data": {
    "projects": [
      {
        "id": "uuid",
        "key": "APP",
        "name": "My Application",
        "description": "...",
        "team_id": "uuid",
        "owner_id": "uuid",
        "created_at": "2025-03-10T00:00:00Z",
        "updated_at": "2025-03-10T00:00:00Z"
      }
    ]
  }
}
```

---

#### GET /projects/:projectId — Получить проект

**Response 200:** данные проекта.

**Ошибки:** 403 (forbidden), 404 (not_found)

---

#### GET /projects/:projectId/members — Участники проекта

Возвращает участников проекта с ролями. У каждого: **`current_task_id`** и **`current_task`** `{ id, title, project_id }`. Доступ: владелец или член команды.

**Response 200:**
```json
{
  "data": {
    "members": [
      {
        "id": "uuid",
        "email": "user@example.com",
        "firstname": "string",
        "lastname": "string",
        "middlename": "string",
        "role": "administrator",
        "avatar_url": "/api/v1/files/users/uuid/avatar.jpg",
        "current_task_id": "uuid",
        "current_task": {
          "id": "uuid",
          "title": "Задача в работе",
          "project_id": "uuid"
        }
      }
    ]
  }
}
```

**role:** administrator | participant | user

**Ошибки:** 403 (forbidden), 404 (not_found)

---

#### GET /projects/:projectId/statuses — Статусы (колонки) проекта

Возвращает динамические статусы проекта. По ним строится канбан-доска. При создании проекта добавляются дефолтные: TODO, IN_PROGRESS, IN_REVIEW, DONE.

**Response 200:**
```json
{
  "data": {
    "statuses": [
      {
        "id": "uuid",
        "project_id": "uuid",
        "key": "TODO",
        "title": "To Do",
        "order": 0
      }
    ]
  }
}
```

**Ошибки:** 403 (forbidden), 404 (not_found)

---

#### POST /projects/:projectId/statuses — Создать статус (колонку)

**Request:**
```json
{
  "key": "BACKLOG",
  "title": "Backlog",
  "order": 0
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| key | string | да | Уникальный код (1–50 символов, A-Z0-9_) |
| title | string | да | Отображаемое название |
| order | int | нет | Порядок колонки (по умолчанию 0) |

**Response 201:** данные созданного статуса.

**Ошибки:** 403 (forbidden), 404 (not_found), 409 (status_key_exists, status_title_exists — отображаемое имя уже занято другой колонкой в этом проекте, без учёта регистра и краевых пробелов)

---

#### PATCH /projects/statuses/:statusId — Обновить статус

**Request (все поля опциональны):**
```json
{
  "key": "TODO",
  "title": "To Do",
  "order": 1
}
```

**Response 200:** обновлённый статус.

**Ошибки:** 403 (forbidden), 404 (not_found), 409 (status_title_exists — как при создании)

---

#### DELETE /projects/statuses/:statusId — Удалить статус

Обязательный query-параметр **`move_to_status_id`** — UUID другой колонки **того же проекта**, в которую переносятся все задачи из удаляемой колонки (обновляются и `status_id`, и ключ `status`). Удаляемый статус и целевой должны различаться. **Последний** статус в проекте удалить нельзя.

**Пример:** `DELETE /api/v1/projects/statuses/{statusId}?move_to_status_id={uuid}`

**Response 204:** No Content

**Ошибки:** 400 (validation_error — нет или невалидный `move_to_status_id`, целевой статус не из этого проекта или совпадает с удаляемым), 403 (forbidden), 404 (not_found), 409 (last_status — в проекте осталась только эта колонка)

---

#### GET /projects/:projectId/board — Канбан-доска

Возвращает колонки (динамические статусы проекта) с задачами в каждой. Порядок колонок определяется полем order статусов.

Опциональные query-параметры те же, что у `GET /projects/:projectId/tasks`: `status`, `assignee_id`, `title`, `type`, `due_from`, `due_to` (см. таблицу выше). В каждой колонке остаются только задачи, подходящие под фильтр; пустые колонки возвращаются с `tasks: []`.

**Response 200:**
```json
{
  "data": {
    "columns": [
      { "status": "TODO", "title": "To Do", "order": 0, "tasks": [...] },
      { "status": "IN_PROGRESS", "title": "In Progress", "order": 1, "tasks": [...] }
    ]
  }
}
```

---

#### POST /projects/:projectId/tasks — Создать задачу

author_id (reporter_id) берётся из текущего аутентифицированного пользователя.

**Request:**
```json
{
  "title": "Реализовать API",
  "description": "Описание задачи",
  "type": "TASK",
  "priority": "MEDIUM",
  "assignee_id": "uuid",
  "due_date": "2025-03-15T12:00:00Z",
  "tags": ["backend", "api"],
  "result_url": "https://example.com/result"
}
```

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| title | string | да | 1–500 символов |
| description | string | нет | Описание задачи |
| type | string | нет | BUG, TASK, STORY (по умолчанию TASK) |
| priority | string | нет | LOW, MEDIUM, HIGH, CRITICAL (по умолчанию MEDIUM) |
| assignee_id | string | нет | UUID исполнителя |
| due_date | string | нет | RFC3339 (ISO дата/время) |
| tags | string[] | нет | Список тегов/меток |
| result_url | string | нет | Ссылка на результат |

**Response 201:**
```json
{
  "data": {
    "id": "uuid",
    "project_id": "uuid",
    "key": "APP-1",
    "title": "Реализовать API",
    "description": "...",
    "status": "TODO",
    "status_id": "uuid",
    "type": "TASK",
    "priority": "MEDIUM",
    "assignee_id": null,
    "reporter_id": "uuid",
    "author_id": "uuid",
    "due_date": "2025-03-15T12:00:00Z",
    "tags": ["backend", "api"],
    "result_url": "https://example.com/result",
    "order": 0,
    "created_at": "2025-03-10T00:00:00Z",
    "updated_at": "2025-03-10T00:00:00Z"
  }
}
```

**status** — ключ колонки из `project_statuses`; **status_id** — UUID строки в `project_statuses` (оба поля согласованы). **type** — BUG, TASK, STORY. **priority** — LOW, MEDIUM, HIGH, CRITICAL. **due_date** — ISO 8601 (RFC3339).

**Ошибки:** 400 (validation_error), 403 (forbidden), 404 (not_found)

---

#### GET /projects/:projectId/tasks — Список задач проекта

Опциональные query-параметры (можно комбинировать):

| Параметр | Описание |
|----------|----------|
| `status` | Фильтр по статусу (ключ колонки), например `TODO` |
| `assignee_id` | UUID исполнителя |
| `title` | Подстрока заголовка (без учёта регистра) |
| `type` | Тип задачи: `BUG`, `TASK`, `STORY` |
| `due_from` | Нижняя граница срока (RFC3339); учитываются только задачи с `due_date` ≥ `due_from` |
| `due_to` | Верхняя граница срока (RFC3339); учитываются только задачи с `due_date` ≤ `due_to` |

Пример: `GET .../tasks?status=IN_PROGRESS&assignee_id=...&title=fix&type=BUG&due_from=2025-03-01T00:00:00Z&due_to=2025-03-31T23:59:59Z`

**Response 200:**
```json
{
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "project_id": "uuid",
        "key": "APP-1",
        "title": "...",
        "description": "...",
        "status": "TODO",
        "status_id": "uuid",
        "type": "TASK",
        "priority": "MEDIUM",
        "assignee_id": null,
        "reporter_id": "uuid",
        "author_id": "uuid",
        "due_date": null,
        "tags": [],
        "result_url": null,
        "order": 0,
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

---

#### GET /tasks/:taskId — Получить задачу

**Response 200:** данные задачи.

---

#### PATCH /tasks/:taskId — Обновить задачу

Частичное обновление (title, description, status, type, priority, assignee_id, due_date, tags, result_url, order).

**Request:**
```json
{
  "title": "Новый заголовок",
  "description": "Обновлённое описание",
  "status": "IN_PROGRESS",
  "type": "BUG",
  "priority": "HIGH",
  "assignee_id": "uuid",
  "due_date": "2025-03-20T00:00:00Z",
  "tags": ["urgent", "hotfix"],
  "result_url": "https://example.com/patch",
  "order": 1
}
```

Все поля опциональны. При смене статуса задача переносится в другую колонку канбана. Пустая строка assignee_id сбрасывает исполнителя. Пустой массив tags очищает теги.

**Response 200:** обновлённая задача (все поля, включая type, tags, result_url).

---

#### DELETE /tasks/:taskId — Удалить задачу

**Response 204:** No Content

---

## Сводная таблица эндпоинтов

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /files/*path | Получить файл (картинку) из хранилища | — |
| POST | /auth/register | Регистрация | — |
| POST | /auth/login | Вход | — |
| GET | /users/me | Текущий пользователь | Bearer |
| PATCH | /users/me | Обновить профиль | Bearer |
| GET | /users/me/tasks | Мои задачи (исполнитель), поле in_work | Bearer |
| PUT | /users/me/current-task | Текущая задача в работе | Bearer |
| PUT | /users/me/avatar | Установить аватарку пользователя | Bearer |
| PUT | /users/me/work-status | Обновить рабочий статус | Bearer |
| GET | /users/me/work-status/history | История изменений статуса | Bearer |
| GET | /organizations | Список моих организаций | Bearer |
| GET | /organizations/:id | Получить организацию по ID | Bearer |
| GET | /organizations/:id/members | Список участников организации | Bearer |
| POST | /organizations | Создать организацию | Bearer |
| PATCH | /organizations/:id | Обновить организацию | Bearer |
| PUT | /organizations/:id/avatar | Установить аватарку | Bearer |
| POST | /organizations/:id/archive | Архивировать организацию | Bearer |
| POST | /organizations/:id/members | Добавить участника | Bearer |
| DELETE | /organizations/:id/members/:user_id | Удалить участника | Bearer |
| GET | /organizations/:id/teams | Команды организации | Bearer |
| POST | /teams | Создать команду | Bearer |
| GET | /teams/:id | Получить команду | Bearer |
| PATCH | /teams/:id | Обновить команду | Bearer |
| PUT | /teams/:id/avatar | Установить аватарку команды | Bearer |
| DELETE | /teams/:id | Удалить команду | Bearer |
| POST | /teams/:id/members | Добавить участника в команду | Bearer |
| DELETE | /teams/:id/members/:user_id | Удалить участника из команды | Bearer |
| GET | /teams/:id/members | Список участников команды | Bearer |
| GET | /teams/:id/projects | Проекты команды | Bearer |
| POST | /projects | Создать проект | Bearer |
| GET | /projects | Список проектов | Bearer |
| GET | /projects/:projectId | Получить проект | Bearer |
| GET | /projects/:projectId/members | Участники проекта | Bearer |
| GET | /projects/:projectId/statuses | Статусы (колонки) проекта | Bearer |
| POST | /projects/:projectId/statuses | Создать статус | Bearer |
| PATCH | /projects/statuses/:statusId | Обновить статус | Bearer |
| DELETE | /projects/statuses/:statusId | Удалить статус | Bearer |
| GET | /projects/:projectId/board | Канбан-доска | Bearer |
| GET | /projects/:projectId/tasks | Список задач проекта | Bearer |
| POST | /projects/:projectId/tasks | Создать задачу | Bearer |
| GET | /tasks/:taskId | Получить задачу | Bearer |
| PATCH | /tasks/:taskId | Обновить задачу | Bearer |
| DELETE | /tasks/:taskId | Удалить задачу | Bearer |

---

## Коды ошибок (error.code)

| code | Описание |
|------|----------|
| validation_error | Ошибка валидации входных данных |
| user_exists | Пользователь с таким email уже существует |
| invalid_credentials | Неверный email или пароль |
| unauthorized | Не авторизован (нет токена или токен невалидный) |
| not_found | Ресурс не найден |
| forbidden | Нет доступа (например, пользователь не в организации) |
| archived | Действие над архивированной организацией запрещено |
| user_already_in_org | Пользователь уже состоит в организации |
| user_already_in_team | Пользователь уже в команде |
| key_exists | Проект с таким key уже существует |
| status_key_exists | Статус с таким key уже существует в проекте |
| status_title_exists | Статус с таким отображаемым названием уже есть в проекте |
| last_status | Нельзя удалить последнюю колонку статуса в проекте |
| internal_error | Внутренняя ошибка сервера |
