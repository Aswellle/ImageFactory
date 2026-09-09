-- ImageForge 数据库迁移脚本 — users 用户表
-- PostgreSQL 15+（与 ent schema 对齐）
--
-- 迁移规范：
--   - 幂等：使用 CREATE TABLE IF NOT EXISTS / IF NOT EXISTS
--   - 零填充数字前缀保证执行顺序
--   - 已应用的迁移不应再修改（checksum 校验）

-- 1. users 用户表（无外键依赖，最先创建）
--    对应 ent/schema/user.go
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,                 -- bcrypt 哈希，敏感字段
    name            VARCHAR(255),                          -- 可选显示名
    role            VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    token_version   INT NOT NULL DEFAULT 0,                -- Token 版本号，密码重置时递增以失效所有 JWT
    totp_secret     VARCHAR(255),                           -- TOTP 密钥（加密存储，敏感字段）
    totp_enabled    BOOLEAN NOT NULL DEFAULT FALSE,         -- 是否启用 2FA
    last_login_ip   VARCHAR(45),                            -- 最后登录 IP（支持 IPv6）
    last_login_at   TIMESTAMPTZ,                            -- 可选
    failed_login_count INT NOT NULL DEFAULT 0,               -- 连续登录失败次数
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);



CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
