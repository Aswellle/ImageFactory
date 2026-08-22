-- ImageForge 迁移 — api_keys API 密钥表
-- 对应 ent/schema/api_key.go
-- 依赖：users（可选外键，user_id 可为 NULL）

CREATE TABLE IF NOT EXISTS api_keys (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL, -- 可选关联
    key_hash        VARCHAR(255) NOT NULL UNIQUE,                   -- 密钥 sha256
    key_prefix      VARCHAR(20) NOT NULL,                            -- 前 8 位，用于展示
    name            VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    permissions     INT NOT NULL DEFAULT 0,                          -- 位图，通过迁移扩展
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_status ON api_keys(status);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);
