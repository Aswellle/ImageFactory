-- ImageForge 迁移 — generation_jobs 生成任务表 + generation_inputs 输入表
-- 对应 ent/schema/generation_job.go / generation_input.go
-- 依赖：users, projects

CREATE TABLE IF NOT EXISTS generation_jobs (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    external_id       VARCHAR(200) UNIQUE,                   -- 外部/上游任务 ID
    project_id        BIGINT REFERENCES projects(id) ON DELETE SET NULL,

    type              VARCHAR(20) NOT NULL CHECK (type IN ('generation', 'edit')),

    status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),

    provider          VARCHAR(100),                          -- e.g. "openai", "sub2api"
    model             VARCHAR(200),
    sub2api_task_id   VARCHAR(200),

    -- 请求参数（与 provider 无关）
    prompt            TEXT,
    negative_prompt   TEXT,
    aspect_ratio      VARCHAR(20),
    image_count       INT NOT NULL DEFAULT 1,
    output_format     VARCHAR(20),                           -- png, jpeg, webp

    -- 失败保留 — 绝不静默吞掉
    error_code        VARCHAR(100),
    error_message     TEXT,
    retry_count       INT NOT NULL DEFAULT 0,

    -- 时间戳
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generation_jobs_user_status ON generation_jobs(user_id, status);
CREATE INDEX IF NOT EXISTS idx_generation_jobs_status_created ON generation_jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_generation_jobs_sub2api_task ON generation_jobs(sub2api_task_id);
CREATE INDEX IF NOT EXISTS idx_generation_jobs_external_id ON generation_jobs(external_id);
CREATE INDEX IF NOT EXISTS idx_generation_jobs_project_id ON generation_jobs(project_id);

-- generation_inputs：编辑任务的源图输入
CREATE TABLE IF NOT EXISTS generation_inputs (
    id                BIGSERIAL PRIMARY KEY,
    job_id            BIGINT NOT NULL REFERENCES generation_jobs(id) ON DELETE CASCADE,
    source_asset_id   BIGINT REFERENCES assets(id) ON DELETE SET NULL, -- 用作输入的已有资产
    storage_key       TEXT,                                             -- 或刚上传的源图
    role              VARCHAR(50),                                      -- "reference", "mask" 等
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generation_inputs_job_id ON generation_inputs(job_id);
CREATE INDEX IF NOT EXISTS idx_generation_inputs_source_asset ON generation_inputs(source_asset_id);

-- Deferred FK: assets.generation_job_id -> generation_jobs (003 created assets first).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_assets_generation_job_id'
          AND table_name = 'assets'
    ) THEN
        ALTER TABLE assets
            ADD CONSTRAINT fk_assets_generation_job_id
            FOREIGN KEY (generation_job_id) REFERENCES generation_jobs(id) ON DELETE SET NULL;
    END IF;
END$$;
