-- Связь проектов с командами: проект принадлежит команде.
ALTER TABLE projects ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_projects_team_id ON projects(team_id);
