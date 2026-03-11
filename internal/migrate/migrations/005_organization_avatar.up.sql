-- Аватар организации (URL в MinIO)
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(1024);
