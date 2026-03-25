-- Текущая задача пользователя (в работе может быть только одна — одна ссылка на задачу).
ALTER TABLE users ADD COLUMN IF NOT EXISTS current_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_users_current_task_id ON users(current_task_id) WHERE current_task_id IS NOT NULL;
