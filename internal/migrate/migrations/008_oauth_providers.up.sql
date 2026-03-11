-- OAuth: провайдер и внешний ID для входа через Google, Yandex, GitHub
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider_id VARCHAR(255);
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_provider_id) WHERE oauth_provider IS NOT NULL AND oauth_provider_id IS NOT NULL;
