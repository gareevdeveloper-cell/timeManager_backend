-- Рабочий статус сотрудника: не работает, отдыхает, обед, отпуск, больничный, командировка
DO $$ BEGIN
    CREATE TYPE work_status_enum AS ENUM (
        'working',       -- работает
        'resting',       -- отдыхает
        'lunch',         -- обед
        'vacation',      -- отпуск
        'sick_leave',    -- больничный
        'business_trip'  -- командировка
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TABLE users ADD COLUMN IF NOT EXISTS work_status work_status_enum DEFAULT 'working';

-- История изменений статуса
CREATE TABLE IF NOT EXISTS user_work_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    work_status work_status_enum NOT NULL,
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changed_by  UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_user_work_status_history_user ON user_work_status_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_work_status_history_changed_at ON user_work_status_history(changed_at);
