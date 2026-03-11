DROP TABLE IF EXISTS project_members;
ALTER TABLE team_members DROP COLUMN IF EXISTS role;
ALTER TABLE organization_members DROP COLUMN IF EXISTS role;
