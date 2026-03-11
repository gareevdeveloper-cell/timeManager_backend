DROP INDEX IF EXISTS idx_organization_members_user_id;
DROP INDEX IF EXISTS idx_organization_members_org_id;
DROP TABLE IF EXISTS organization_members;

ALTER TABLE organizations DROP COLUMN IF EXISTS updated_at;
ALTER TABLE organizations DROP COLUMN IF EXISTS status;
