-- ImageForge 迁移 — image_tasks 异步图片任务表
-- 对应 internal/domain/image_task/task.go 的 Record 结构
--
-- 注意：image_task 的主存储是 Redis（RedisImageTaskStore），按 TTL 过期。
-- 本表作为关系型镜像/兜底，供审计、离线查询与 Redis 失效时恢复使用。
-- 依赖：users

CREATE TABLE IF NOT EXISTS image_tasks (
    id              VARCHAR(200) PRIMARY KEY,                  -- 与 Redis 记录 ID 一致
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    status_code     INT,
    result          BYTEA,                                      -- 二进制结果载荷
    error           BYTEA,                                      -- 错误载荷
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_tasks_user_id ON image_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_image_tasks_status ON image_tasks(status);
CREATE INDEX IF NOT EXISTS idx_image_tasks_created_at ON image_tasks(created_at);
