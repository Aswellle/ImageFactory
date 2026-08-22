-- ImageForge 迁移 — collections 资产集合表
-- 对应 ent/schema/collection.go
-- 依赖：users

CREATE TABLE IF NOT EXISTS collections (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections(user_id);
