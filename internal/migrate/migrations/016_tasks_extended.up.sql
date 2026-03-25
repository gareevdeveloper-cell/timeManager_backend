-- Расширение таблицы tasks: type, tags, result_url
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS type VARCHAR(50) NOT NULL DEFAULT 'TASK',
    ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS result_url TEXT;

-- Дефолтные значения для существующих записей (уже заданы через DEFAULT)
-- type: TASK, tags: [], result_url: NULL
