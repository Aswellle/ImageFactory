# ImageForge 后端代码质量修复计划

> 基于 `/code-quality` 全面审查报告（83 个发现），分 4 个阶段推进。
> 每阶段独立可验证，按优先级排序。

---

## 阶段 1：P0 致命问题（8 项）— 必须立即修复

> 目标：消除生产环境崩溃、数据竞争、安全绕过风险。

### 1.1 crypto 包假加密

**文件**: `internal/pkg/crypto/crypto.go`

- [ ] 1.1.1 `Encrypt` 在 key 未设置时返回错误而非 base64 明文
- [ ] 1.1.2 `Decrypt` 在 key 未设置时返回错误
- [ ] 1.1.3 添加 `IsEnabled()` 方法供调用方检测
- [ ] 1.1.4 启动时若 `IF_CREDENTIAL_KEY` 未配置则 `log.Fatal`（生产环境强制要求）

### 1.2 batchimage provider 数据竞争

**文件**: `internal/batchimage/provider.go`

- [ ] 2.1.1 `withCause` 返回新分配的 Error 副本而非修改全局 sentinel
- [ ] 2.1.2 使用 `go test -race` 验证无数据竞争

### 1.3 public_service 死锁

**文件**: `internal/batchimage/public_service.go`

- [ ] 3.1.1 将 `persistState` 拆分为 `persistStateLocked`（不加锁）和 `persistState`（加锁）
- [ ] 3.1.2 `markItemStatus` 调用 `persistStateLocked`（已持写锁）
- [ ] 3.1.3 `Cancel` 调用 `persistStateLocked`（已持写锁）
- [ ] 3.1.4 `UpdateJobStatus` 调用 `persistStateLocked`（已持写锁）
- [ ] 3.1.5 外部调用路径（如 `restoreFromDB`）调用 `persistState`（自动加锁）

### 1.4 memory_queue worker 死亡螺旋

**文件**: `internal/job/memory_queue.go`

- [ ] 4.1.1 `runTask` 添加 `recover()` 捕获恐慌并记录日志
- [ ] 4.1.2 Submit/Stop 使用 `sync.Once` 或 select+done channel 防止向已关闭通道发送

### 1.5 refresh_token 丢失请求上下文

**文件**: `internal/service/refresh_token.go`

- [ ] 5.1.1 将 `ctx()` 改为从结构体字段获取（构造时传入）
- [ ] 5.1.2 所有使用 `ctx()` 的地方改为 `r.ctx`

### 1.6 scheduling_bridge thresholds 无同步

**文件**: `internal/service/scheduling_bridge.go`

- [ ] 6.1.1 添加 `sync.RWMutex` 保护 `thresholds` map
- [ ] 6.1.2 `GetThresholds` 使用读锁
- [ ] 6.1.3 `SetThresholds` 使用写锁

### 1.7 jwt debug 模式自动生成 secret

**文件**: `internal/service/jwt.go`

- [ ] 7.1.1 生产环境（非 debug）若 `IF_JWT_SECRET` 未设置则启动失败
- [ ] 7.1.2 debug 模式生成临时 secret 时输出醒目警告日志

### 1.8 image_edit 占位逻辑返回源图片

**文件**: `internal/service/image_edit.go`

- [ ] 8.1.1 `createEditVersion` 在编辑逻辑未实现时返回明确错误而非存储源图片
- [ ] 8.1.2 添加 `ErrImageEditNotImplemented` 错误码

---

## 阶段 2：P1 高优先级（15 项）— 功能 bug 和安全漏洞

> 目标：修复功能缺陷、安全漏洞、数据一致性问题。

### 2.1 Gemini 结果引用为空

**文件**: `internal/batchimage/gemini_provider.go`

- [ ] 1.1.1 `mapGeminiState` 读取 `ResponsesFileSnake`（snake_case）而非 `ResponsesFile`
- [ ] 1.1.2 添加单元测试覆盖 snake_case 和 camelCase 两种格式

### 2.2 public_service Get 吞错误

**文件**: `internal/batchimage/public_service.go`

- [ ] 2.2.1 `Get` 方法在 `provider.Get` 失败时返回错误
- [ ] 2.2.2 调用方处理错误并返回 500

### 2.3 worker pollJob 传入 nil account

**文件**: `internal/batchimage/worker.go`

- [ ] 2.3.1 为 `Worker` 添加 `BatchImageAccountResolver` 字段
- [ ] 2.3.2 `pollJob` 按 `job.AccountID` 解析 account
- [ ] 2.3.3 解析失败时记录错误并跳过该任务

### 2.4 worker provider.Get 错误静默忽略

**文件**: `internal/batchimage/worker.go`

- [ ] 2.4.1 `provider.Get` 失败时返回错误
- [ ] 2.4.2 实现指数退避重试（最多 3 次）

### 2.5 public_service Delete 未校验归属

**文件**: `internal/batchimage/public_service.go`

- [ ] 2.5.1 `Delete` 方法校验 `job.UserID == owner.UserID`
- [ ] 2.5.2 不匹配时返回 `ErrBatchImageJobNotFound`

### 2.6 generation_service handleCompleted 数据不一致

**文件**: `internal/service/generation_service.go`

- [ ] 2.6.1 资产创建失败时将任务标记为 `failed` 而非 `completed`
- [ ] 2.6.2 添加错误信息到任务记录

### 2.7 image_task randomString 取模有偏

**文件**: `internal/service/image_task.go`

- [ ] 2.7.1 使用 `math/rand/v2` 的 `rand.N` 替代时间取模
- [ ] 2.7.2 `account_resolver.go` 同步修复

### 2.8 admin SetUserStatus 静默回退

**文件**: `internal/service/admin.go`

- [ ] 2.8.1 未知状态返回 `ErrInvalidStatus` 错误
- [ ] 2.8.2 `UpdateUser` 校验枚举值

### 2.9 apikey_service GetUsage 是 stub

**文件**: `internal/service/apikey_service.go`

- [ ] 2.9.1 实现真实用量查询（从 usage_records 表聚合）
- [ ] 2.9.2 或返回 `ErrNotImplemented` 明确标识

### 2.10 filesystem 路径遍历

**文件**: `internal/storage/filesystem.go`

- [ ] 2.10.1 验证 key 不包含 `..`
- [ ] 2.10.2 `Get` 返回前检查解析路径是否在 base 目录下
- [ ] 2.10.3 添加单元测试覆盖路径遍历攻击向量

### 2.11 image_task_store 返回共享指针

**文件**: `internal/repository/image_task_store.go`

- [ ] 2.11.1 `Get` 返回深拷贝（或只读视图）
- [ ] 2.11.2 `List` 返回切片副本

### 2.12 rate_limiter TOCTOU 竞争

**文件**: `internal/repository/rate_limiter.go`

- [ ] 2.12.1 Redis 路径使用 Lua 脚本实现原子 Get-then-Set
- [ ] 2.12.2 内存路径使用 `sync.Mutex`

### 2.13 config DSN 拼接注入

**文件**: `internal/config/config.go`

- [ ] 2.13.1 使用 `url.UserPassword` 进行 URL 编码
- [ ] 2.13.2 特殊字符测试用例

### 2.14 HTTP Server 缺少 ReadTimeout

**文件**: `cmd/server/main.go`

- [ ] 2.14.1 添加 `ReadTimeout: 30 * time.Second`
- [ ] 2.14.2 添加 `WriteTimeout: 30 * time.Second`
- [ ] 2.14.3 添加 `IdleTimeout: 120 * time.Second`

### 2.15 password_reset token version 原子性

**文件**: `internal/service/password_reset_service.go`

- [ ] 2.15.1 在事务中完成密码更新和 token version 递增
- [ ] 2.15.2 任一失败则回滚整个事务

---

## 阶段 3：P2 中等优先级（35 项）— 性能、健壮性、可维护性

> 目标：提升性能、消除边界条件、改善代码质量。

### 3.1 并发与性能

- [ ] 3.1.1 `thumbnail.go GenerateBatch` 使用 goroutine 池并发处理
- [ ] 3.1.2 `asset_service.go GetContent` 使用 io.Reader 流式传输
- [ ] 3.1.3 `admin.go Dashboard` 使用批量查询消除 N+1
- [ ] 3.1.4 `prompt_template.go Update` 在事务中完成检查和更新

### 3.2 安全与边界

- [ ] 3.2.1 `image.Decode` 添加图片大小限制（防解压炸弹）
- [ ] 3.2.2 `thumbnail.go` 验证尺寸参数（>0）
- [ ] 3.2.3 `reset_code_store.go` 添加尝试次数限制（防暴力破解）
- [ ] 3.2.4 `rate_limiter.go` 内存 map 添加 TTL 清理（防内存泄漏）
- [ ] 3.2.5 `account_repository.go` 返回副本而非原地修改

### 3.3 错误处理与上下文

- [ ] 3.3.1 `totp.go GenerateBackupCodesJSON` 处理 json.Marshal 错误
- [ ] 3.3.2 `email_service.go stripHTML` 使用 html.UnescapeString
- [ ] 3.3.3 `errors.go WithRequestID` 返回新副本而非修改原对象
- [ ] 3.3.4 `migrate.go` 启动时验证迁移校验和

### 3.4 配置与部署

- [ ] 3.4.1 `config.go Config` 实现 String() 方法脱敏敏感字段
- [ ] 3.4.2 `config.go` 使用 strconv.Itoa 替代自定义 itoa
- [ ] 3.4.3 `main.go` 部分 TLS 配置（只有 cert 或 key）返回错误
- [ ] 3.4.4 `r2.go` 使用可取消的 context 替代 context.Background()
- [ ] 3.4.5 `storage.go` 未知 provider 返回错误而非静默回退

### 3.5 移植代码修复

- [ ] 3.5.1 `scheduling/pool.go ScoreCandidates` 修正 loadFactor 语义
- [ ] 3.5.2 `scheduling/ratelimit.go` 429 默认回退添加最大退避时间
- [ ] 3.5.3 `scheduling/session_window.go` 处理字符串时间戳格式
- [ ] 3.5.4 `scheduling/threshold.go` 改进 utilizationAsPercent 启发式

### 3.6 其他 P2

- [ ] 3.6.1 `crypto.go` 添加密钥版本标识（支持轮换）
- [ ] 3.6.2 `crypto.go` 使用更独特的版本化前缀
- [ ] 3.6.3 `image_task.go ToPublic` 创建新切片复制数据
- [ ] 3.6.4 `admin-cli/main.go` 使用参数化查询替代原始 SQL
- [ ] 3.6.5 `admin-cli/main.go` 角色比较使用 domain 常量
- [ ] 3.6.6 `web/web.go` 添加构建时 dist/ 存在性检查

---

## 阶段 4：P3 低优先级（12 项）— 代码质量微调

> 目标：统一代码风格、消除技术债务。

- [ ] 4.1 `public_service.go List` 返回内部 map 的副本
- [ ] 4.2 `public_service.go restoreFromDB` 检查 rows.Err()
- [ ] 4.3 `gemini_client.go` 日志中脱敏 API key
- [ ] 4.4 `scheduling/ratelimit.go` 529 冷却时间可配置
- [ ] 4.5 `scheduling/threshold.go` 改进利用率计算
- [ ] 4.6 `project_service.go` 使用 strconv 替代 fmt.Sscanf
- [ ] 4.7 `response/response.go` 使用类型 switch 替代裸 assertion
- [ ] 4.8 `logger/logger.go` 添加秘密脱敏中间件
- [ ] 4.9 `job/job.go` 移除 Sub2APITaskID 或添加文档说明
- [ ] 4.10 `domain/constants.go` 移除重复错误码定义
- [ ] 4.11 `user_repository.go` 统一注释语言
- [ ] 4.12 `db.go migration` 使用可取消 context

---

## 执行规则

1. **每项完成后**运行 `go build ./...` 和 `go test -tags=unit ./...`
2. **阶段完成后**运行 `go test -race ./...` 验证并发安全
3. **提交粒度**：每个阶段完成后提交一次，commit message 格式 `fix: 阶段N — 简要描述`
4. **验证顺序**：编译 → 单元测试 → race 检测 → 推送

## 预计工作量

| 阶段 | 项数 | 预计时间 | 风险 |
|------|------|----------|------|
| P0 | 8 | 2-3 小时 | 高（涉及并发和加密） |
| P1 | 15 | 4-6 小时 | 中（功能修复） |
| P2 | 35 | 6-8 小时 | 低（渐进改进） |
| P3 | 12 | 2-3 小时 | 低（风格调整） |
| **合计** | **70** | **14-20 小时** | — |
