-- ImageForge 迁移 — accounts AI API 账号表
-- 对应 ent/schema/account.go
-- 无外部依赖（独立表）

CREATE TABLE IF NOT EXISTS accounts (
    id                    BIGSERIAL PRIMARY KEY,
    name                  VARCHAR(100) NOT NULL,
    platform              VARCHAR(50) NOT NULL,
    type                  VARCHAR(20) NOT NULL,
    credentials           JSONB NOT NULL DEFAULT '{}',
    priority              INT NOT NULL DEFAULT 50,
    status                VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'error', 'disabled')),
    error_message         TEXT,
    last_used_at          TIMESTAMPTZ,
    expires_at            TIMESTAMPTZ,
    schedulable           BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limited_at       TIMESTAMPTZ,
    rate_limit_reset_at   TIMESTAMPTZ,
    overload_until        TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引：按平台查询可调度账号（调度器最常用查询）
CREATE INDEX IF NOT EXISTS idx_accounts_platform_schedulable_status
    ON accounts(platform, schedulable, status);

CREATE INDEX IF NOT EXISTS idx_accounts_platform ON accounts(platform);
CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);
CREATE INDEX IF NOT EXISTS idx_accounts_schedulable ON accounts(schedulable);
CREATE INDEX IF NOT EXISTS idx_accounts_priority ON accounts(priority);
CREATE INDEX IF NOT EXISTS idx_accounts_last_used_at ON accounts(last_used_at);
