DROP TABLE IF EXISTS user_work_status_history;
ALTER TABLE users DROP COLUMN IF EXISTS work_status;
DROP TYPE IF EXISTS work_status_enum;
