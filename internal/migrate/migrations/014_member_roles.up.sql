-- Роли для organization_members: administrator, participant, user
ALTER TABLE organization_members ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'participant';
UPDATE organization_members om SET role = 'administrator'
FROM organizations o WHERE om.organization_id = o.id AND om.user_id = o.owner_id;

-- Роли для team_members
ALTER TABLE team_members ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'participant';
UPDATE team_members tm SET role = 'administrator'
FROM teams t WHERE tm.team_id = t.id AND tm.user_id = t.creator_id;

-- Таблица участников проекта с ролями
CREATE TABLE IF NOT EXISTS project_members (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(50) NOT NULL DEFAULT 'participant',
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id);

-- Владельцы проектов получают роль administrator
INSERT INTO project_members (project_id, user_id, role)
SELECT id, owner_id, 'administrator' FROM projects
ON CONFLICT (project_id, user_id) DO NOTHING;
