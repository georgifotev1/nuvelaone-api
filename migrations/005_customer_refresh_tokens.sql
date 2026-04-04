-- +goose Up
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS user_id;
ALTER TABLE refresh_tokens ADD COLUMN entity_id VARCHAR(27) NOT NULL;
ALTER TABLE refresh_tokens ADD COLUMN entity_type VARCHAR(10) NOT NULL;
ALTER TABLE refresh_tokens ADD COLUMN tenant_id VARCHAR(27) REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX idx_refresh_tokens_entity_id ON refresh_tokens(entity_id);
CREATE INDEX idx_refresh_tokens_tenant_id ON refresh_tokens(tenant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_tokens_entity_id;
DROP INDEX IF EXISTS idx_refresh_tokens_tenant_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS entity_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS entity_type;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE refresh_tokens ADD COLUMN user_id VARCHAR(27) REFERENCES users(id) ON DELETE CASCADE;
