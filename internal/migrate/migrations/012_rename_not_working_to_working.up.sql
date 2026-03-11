-- Переименование not_working -> working (для БД, где уже применена 011 с not_working)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_enum e JOIN pg_type t ON e.enumtypid = t.oid WHERE t.typname = 'work_status_enum' AND e.enumlabel = 'not_working') THEN
    ALTER TYPE work_status_enum RENAME VALUE 'not_working' TO 'working';
  END IF;
END $$;
ALTER TABLE users ALTER COLUMN work_status SET DEFAULT 'working';
