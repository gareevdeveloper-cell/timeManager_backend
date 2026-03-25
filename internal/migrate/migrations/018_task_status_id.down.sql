ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_id_fkey;
DROP INDEX IF EXISTS idx_tasks_status_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS status_id;
