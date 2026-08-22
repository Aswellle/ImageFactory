-- ImageForge 迁移 — projects 项目表
-- 对应 ent/schema/project.go
-- 依赖：users

CREATE TABLE IF NOT EXISTS projects (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_user_status ON projects(user_id, status);
