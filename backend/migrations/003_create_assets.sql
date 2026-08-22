-- ImageForge 迁移 — assets 资产表（图片）
-- 对应 ent/schema/asset.go
-- 依赖：users, projects, generation_jobs

CREATE TABLE IF NOT EXISTS assets (
    id                    BIGSERIAL PRIMARY KEY,
    user_id               BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id            BIGINT REFERENCES projects(id) ON DELETE SET NULL,
    generation_job_id     BIGINT,                            -- FK added in 008 (generation_jobs created later)

    source                VARCHAR(20) NOT NULL CHECK (source IN ('generated', 'uploaded', 'edited')),
    status                VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deleted')),

    title                 VARCHAR(500),
    description           TEXT,

    prompt                TEXT,
    negative_prompt       TEXT,

    model                 VARCHAR(200),
    model_provider        VARCHAR(200),

    width                 INT,
    height                INT,
    aspect_ratio          VARCHAR(20),

    mime_type             VARCHAR(100),
    file_size             BIGINT,

    -- 存储键（非 URL），原图与缩略图分开存储
    storage_key           TEXT NOT NULL UNIQUE,
    thumbnail_key         TEXT,
    medium_key            TEXT,

    current_version       INT NOT NULL DEFAULT 1,

    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_assets_user_status ON assets(user_id, status);
CREATE INDEX IF NOT EXISTS idx_assets_project_id ON assets(project_id);
CREATE INDEX IF NOT EXISTS idx_assets_storage_key ON assets(storage_key);
CREATE INDEX IF NOT EXISTS idx_assets_generation_job_id ON assets(generation_job_id);
CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets(deleted_at);
