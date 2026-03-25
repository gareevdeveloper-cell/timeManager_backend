DROP INDEX IF EXISTS idx_users_current_task_id;
ALTER TABLE users DROP COLUMN IF EXISTS current_task_id;
