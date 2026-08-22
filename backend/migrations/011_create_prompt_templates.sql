-- ImageForge 迁移 — prompt_templates 提示词模板表
-- 对应 ent/schema/prompt_template.go
-- 依赖：users

CREATE TABLE IF NOT EXISTS prompt_templates (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    content         TEXT NOT NULL,                                      -- 含 {{vars}} 的模板体
    variables       TEXT,                                               -- 逗号分隔的变量名
    category        VARCHAR(20) NOT NULL DEFAULT 'custom'
                        CHECK (category IN ('product', 'scene', 'style', 'custom')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prompt_templates_user_id ON prompt_templates(user_id);
CREATE INDEX IF NOT EXISTS idx_prompt_templates_category ON prompt_templates(category);
