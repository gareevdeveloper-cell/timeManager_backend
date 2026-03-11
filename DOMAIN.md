# Domain Model

Краткое описание предметной области.

## Основные сущности

- `User` — Пользователь системы. id(UUID), email, password_hash, firstname, lastname, middlename, birthday, role, status, created_at, updated_at. Пароль хранится как bcrypt-хеш.
- `Organization` — Организация. Создаётся аутентифицированным пользователем. Поля: id(UUID), name, slug, owner_id, status (active/archived), created_at, updated_at. Связь один-ко-многим с пользователями через `organization_members` (пользователь может быть только в одной организации). Роли участников: administrator, participant, user.
- `Team` — Команда. Принадлежит организации (один-ко-многим). В команде могут состоять несколько пользователей (team_members). Поля: id, name, description, organization_id, creator_id, created_at. Добавлять в команду можно только пользователей организации. Роли участников: administrator, participant, user. 
- `Project` — Проект. Может принадлежать команде (team_id опционален) или быть личным. Поля: id(UUID), key (уникальный короткий код, например APP), name, description, team_id (nullable), owner_id, next_task_number, created_at, updated_at. Участники хранятся в `project_members` с ролями: administrator, participant, user. Владелец при создании получает роль administrator.
- `ProjectStatus` — Статус (колонка) канбан-доски проекта. Динамический — каждый проект имеет свой набор. Поля: id, project_id, key (уникален в проекте, например TODO), title (отображаемое название), order (порядок колонки).
- `Task` — Задача (issue) в проекте. Поля: id(UUID), project_id, key (уникальный в системе, например APP-1), title, description, status (ключ из project_statuses проекта), priority (LOW, MEDIUM, HIGH, CRITICAL), assignee_id (nullable), reporter_id, due_date (nullable), order (позиция в колонке), created_at, updated_at. key генерируется как PROJECTKEY-номер. 
- `Chat` - Чат. Переписка между нескольки пользователями и тет-а-тет
- `Message` - Сообщения в чате.

## Инварианты и бизнес‑правила

- Примеры:
  - Если пользователь в статусе на месте - то тогда он может выполнять действия.
  - Пользователь может быть только в одной организации.

Любой код (от человека или агента) не должен нарушать эти правила.
