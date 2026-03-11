DROP INDEX IF EXISTS idx_projects_team_id;
ALTER TABLE projects DROP COLUMN IF EXISTS team_id;
