ALTER TYPE work_status_enum RENAME VALUE 'working' TO 'not_working';
ALTER TABLE users ALTER COLUMN work_status SET DEFAULT 'not_working';
