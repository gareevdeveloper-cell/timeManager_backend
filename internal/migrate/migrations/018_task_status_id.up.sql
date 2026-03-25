-- Связь задачи со строкой project_statuses (колонка канбана).
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS status_id UUID REFERENCES project_statuses(id) ON DELETE RESTRICT;

UPDATE tasks t
SET status_id = ps.id
FROM project_statuses ps
WHERE ps.project_id = t.project_id AND ps.key = t.status;

UPDATE tasks t
SET status_id = sub.id
FROM (
  SELECT DISTINCT ON (project_id) id, project_id
  FROM project_statuses
  ORDER BY project_id, "order" ASC, key ASC
) sub
WHERE t.project_id = sub.project_id AND t.status_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_status_id ON tasks(status_id);

ALTER TABLE tasks ALTER COLUMN status_id SET NOT NULL;
