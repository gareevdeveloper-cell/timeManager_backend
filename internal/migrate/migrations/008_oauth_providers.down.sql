DROP INDEX IF EXISTS idx_users_oauth;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider_id;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
