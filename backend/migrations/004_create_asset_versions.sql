-- ImageForge 迁移 — asset_versions 资产版本表
-- 对应 ent/schema/asset_version.go
-- 依赖：assets

CREATE TABLE IF NOT EXISTS asset_versions (
    id                BIGSERIAL PRIMARY KEY,
    asset_id          BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    version           INT NOT NULL CHECK (version > 0), -- 1-based，每个资产递增
    edit_job_id       BIGINT,                            -- 关联的编辑 job（可选）

    prompt            TEXT,
    negative_prompt   TEXT,
    model             VARCHAR(200),

    width             INT,
    height            INT,

    mime_type         VARCHAR(100),
    file_size         BIGINT,

    storage_key       TEXT NOT NULL UNIQUE,
    thumbnail_key     TEXT,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_versions_asset_version ON asset_versions(asset_id, version);
CREATE INDEX IF NOT EXISTS idx_asset_versions_storage_key ON asset_versions(storage_key);
CREATE INDEX IF NOT EXISTS idx_asset_versions_edit_job_id ON asset_versions(edit_job_id);
