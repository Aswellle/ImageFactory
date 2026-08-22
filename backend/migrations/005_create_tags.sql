-- ImageForge 迁移 — tags 标签表 + asset_tags 多对多关联表
-- 对应 ent/schema/tag.go / asset_tag.go
-- 依赖：users, assets

CREATE TABLE IF NOT EXISTS tags (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    color           VARCHAR(50),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 同一用户下标签名唯一
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_user_name ON tags(user_id, name);
CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

-- asset_tags：assets ↔ tags 多对多中间表
CREATE TABLE IF NOT EXISTS asset_tags (
    asset_id        BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    tag_id          BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (asset_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_asset_tags_tag_id ON asset_tags(tag_id);
