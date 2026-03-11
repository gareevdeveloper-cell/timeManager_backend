-- Динамические статусы (колонки) для проектов
CREATE TABLE IF NOT EXISTS project_statuses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key        VARCHAR(50) NOT NULL,
    title      VARCHAR(255) NOT NULL,
    "order"    INTEGER NOT NULL DEFAULT 0,
    UNIQUE(project_id, key)
);

CREATE INDEX IF NOT EXISTS idx_project_statuses_project_id ON project_statuses(project_id);

-- Дефолтные статусы для существующих проектов
INSERT INTO project_statuses (project_id, key, title, "order")
SELECT id, 'TODO', 'To Do', 0 FROM projects
ON CONFLICT (project_id, key) DO NOTHING;

INSERT INTO project_statuses (project_id, key, title, "order")
SELECT id, 'IN_PROGRESS', 'In Progress', 1 FROM projects
ON CONFLICT (project_id, key) DO NOTHING;

INSERT INTO project_statuses (project_id, key, title, "order")
SELECT id, 'IN_REVIEW', 'In Review', 2 FROM projects
ON CONFLICT (project_id, key) DO NOTHING;

INSERT INTO project_statuses (project_id, key, title, "order")
SELECT id, 'DONE', 'Done', 3 FROM projects
ON CONFLICT (project_id, key) DO NOTHING;
