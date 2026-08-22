-- ImageForge 迁移 — usage_records 用量记录表
-- 对应 ent/schema/usage_record.go
-- 依赖：users, api_keys

CREATE TABLE IF NOT EXISTS usage_records (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id      BIGINT REFERENCES api_keys(id) ON DELETE SET NULL, -- 通过 API key 调用时设置

    type            VARCHAR(20) NOT NULL CHECK (type IN ('generation', 'edit', 'upload', 'storage')),
    model           VARCHAR(200),
    image_count     INT NOT NULL DEFAULT 1,
    tokens          INT,                                                -- 文本/视觉混合模型

    -- 费用（NUMERIC(20,10) 在应用层映射为 decimal）
    cost            NUMERIC(20, 10) NOT NULL DEFAULT 0,

    request_id      VARCHAR(200) UNIQUE,                                -- 幂等/追踪

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_records_user_created ON usage_records(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_usage_records_api_key ON usage_records(api_key_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_request_id ON usage_records(request_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_type ON usage_records(type);
