-- ImageForge 迁移 — favorites 收藏表
-- 对应 ent/schema/favorite.go
-- 依赖：users, assets

CREATE TABLE IF NOT EXISTS favorites (
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id        BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, asset_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_asset_id ON favorites(asset_id);
CREATE INDEX IF NOT EXISTS idx_favorites_created_at ON favorites(user_id, created_at);
