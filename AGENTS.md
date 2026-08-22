# ImageForge 工程规范

## 0. 项目使命

ImageForge 是一个**商业级 AI 图片生产工作台**，以 **Sub2API 本地代码库为基座**构建。

ImageForge 不是一个通用 AI 聊天应用。
ImageForge 不是 Sub2API 的管理后台复刻。
ImageForge 不是一个 API 中转面板。

产品目标：

> "创建、编辑、组织、管理商业化视觉资产，并以 API 形式对外提供能力。"

用户体验必须是：图片优先、极简、专业、Apple 风格。

---

## 1. 核心架构原则：以 Sub2API 为基座

**Sub2API 是本项目的基座代码库**，不是外部依赖服务。

ImageForge 直接复用 Sub2API 的源代码实现：

- **能直接复用的代码** → 直接复制源文件到 ImageForge，适配包路径后使用
- **能复用的开发模式** → 遵循 Sub2API 的架构模式（错误处理、中间件链、队列 worker、Ent 用法）
- **能复用的处理逻辑** → 移植 Sub2API 的业务逻辑（图片管道、账户调度、计费、故障转移）
- **能复用的数据模型** → 通过 Go workspace 直接导入 Sub2API 的 `ent/` 生成包

**最终目标：ImageForge 单独部署上线即可提供完整的商业化服务，无需依赖外部 API（除必要的 AI 上游调用外）。**

### 1.1 Go Workspace 机制

项目使用 Go workspace（`D:\Git-Clone\SRC\Sub2API\go.work`）链接两个模块：

```
github.com/Wei-Shaw/sub2api   ← 基座（本地代码库 D:\Git-Clone\SRC\Sub2API\sub2api\backend）
github.com/imageforge/imageforge  ← 本项目
```

通过 workspace，ImageForge 可以直接 import Sub2API 的公开包：

```go
import "github.com/Wei-Shaw/sub2api/ent"           // Ent 数据模型
import "github.com/Wei-Shaw/sub2api/ent/user"       // 用户模型
import "github.com/Wei-Shaw/sub2api/ent/account"    // 账户模型
```

### 1.2 代码复用层级

| 层级 | 复用方式 | 示例 |
|------|----------|------|
| **Ent 数据模型** | 直接 import `ent/` 包 | `ent.User`、`ent.Account`、`ent.BatchImageJob` |
| **服务层逻辑** | 复制源文件 → 适配包路径 → 编译验证 | `batch_image` 管道、`openai_images` 网关 |
| **开发模式** | 遵循 Sub2API 的架构约定 | 错误码枚举、中间件链、worker 队列、配置结构 |
| **前端组件** | 参考 Sub2API 的 Vue 组件模式 | Pinia store、axios 客户端、路由守卫 |

### 1.3 移植标准流程

将 Sub2API 的 `internal/` 包移植到 ImageForge 时，严格遵循以下步骤：

1. **复制** — 将源文件复制到 ImageForge 对应的包目录
2. **改包名** — 将 `package service` 改为 ImageForge 的目标包名
3. **换导入** — 将 `github.com/Wei-Shaw/sub2api/internal/...` 替换为 ImageForge 本地包或 `ent/` 直接导入
4. **去 internal 依赖** — 将 Sub2API `internal/` 的类型替换为本地 stub 或直接使用 ent 类型
5. **编译验证** — `go build ./...` 通过
6. **测试** — `go test ./...` 通过

---

## 2. 绝对开发规则

修改任何代码之前：

1. 阅读相关现有实现（先查 Sub2API，再查 ImageForge）
2. 理解当前架构和数据流
3. 搜索已有可复用功能（**先搜 Sub2API，再搜 ImageForge**）
4. 不创建重复基础设施
5. 不修改 Sub2API 源文件（只复制适配）
6. **能复用就复用，不能复用再新建**
7. 保持变更最小且可逆
8. 修改后运行相关测试

**绝不盲目重写已有子系统。**

---

## 3. Sub2API 边界与复用策略

### 3.1 直接复用（通过 Go workspace import）

以下 Sub2API 包可直接导入使用：

- `github.com/Wei-Shaw/sub2api/ent/...` — 所有 Ent 生成模型
- `github.com/Wei-Shaw/sub2api/ent/schema/...` — Ent  Schema 定义（参考）

### 3.2 复制移植（internal 包）

以下 Sub2API 功能需要复制源文件到 ImageForge 后适配使用：

| Sub2API 功能 | 源路径 | ImageForge 目标包 |
|---|---|---|
| 异步图片批处理管道 | `internal/service/batch_image*.go` | `internal/batchimage/` |
| 同步图片网关 | `internal/service/openai_images.go` | `internal/imagegateway/` |
| 异步图片任务处理器 | `internal/handler/image_task_handler.go` | `internal/handler/` |
| 账户调度 | `internal/service/account.go` | `internal/account/` |
| 计费服务 | `internal/service/billing_service.go` | `internal/billing/` |
| 错误处理 | `internal/pkg/errors/` | `internal/pkg/errors/` |
| 配置结构 | `internal/config/config.go` | `internal/config/` |
| Redis 队列 | `internal/service/batch_image_queue.go` | `internal/batchimage/` |

### 3.3 ImageForge 独有（新建）

以下功能 Sub2API 不存在，由 ImageForge 全新实现：

- 产品用户系统（`ent/user` 产品级用户，区别于 Sub2API 的网关用户）
- 项目管理（`ent/project`）
- 资产库（`ent/asset`、`ent/asset_version`）
- 提示词模板（`ent/prompt_template`）
- 产品级 API Key（`ent/api_key`）
- 收藏与标签（`ent/favorite`、`ent/tag`、`ent/collection`）
- 产品级生成任务（`ent/generation_job`）
- 前端 UI（Vue 3 工作台、画廊、编辑器）

---

## 4. 架构流向

```
浏览器
  → ImageForge Web (Vue 3 SPA)
  → ImageForge API (Gin)
    → 认证 (JWT)
    → 授权 (中间件)
    → 产品服务层 (项目/资产/提示词/API Key)
    → 图片生成引擎 (移植自 Sub2API)
      → 账户调度 (移植自 Sub2API)
      → 模型路由 (移植自 Sub2API)
      → 上游 AI 调用 (OpenAI/DALL·E/Gemini)
      → 故障转移 (移植自 Sub2API)
    → Worker 队列 (移植自 Sub2API)
    → 对象存储
    → 资产记录
    → 浏览器
```

**关键区别：** 图片生成引擎不是外部服务调用，而是 ImageForge 内部直接运行的移植代码。

---

## 5. 安全

所有受保护资源必须验证：

认证 + 授权 + 资源所有权

绝不信任客户端传入的 ID。

绝不使用客户端 user_id 做授权。

API Key 不明文存储，存储哈希。

密钥始终在服务端。

绝不提交：API Key、OAuth Token、refresh token、cookie、session 密钥、生产凭证。

---

## 6. 图片存储

绝不将图片二进制存入数据库字段。

使用对象存储（S3/MinIO/R2）。

数据库只存：storage key、MIME 类型、尺寸、大小、校验和、元数据。

私有资源使用签名 URL。

缩略图异步生成。

---

## 7. 图片任务

图片生成必须建模为异步任务。

有效状态：pending、processing、completed、failed、cancelled。

失败任务必须保留：request ID、provider、model、error code、脱敏 error message、时间戳。

绝不静默吞掉生成失败。

绝不把 HTTP 200 当作生成成功的充分证据，必须验证实际图片输出。

---

## 8. 错误处理

绝不把原始上游异常直接暴露给用户。

将内部错误映射为稳定的 ImageForge 错误码。

示例：IMAGE_GENERATION_FAILED、IMAGE_EDIT_FAILED、UPSTREAM_TIMEOUT、UPSTREAM_RATE_LIMITED、MODEL_UNAVAILABLE、AUTHENTICATION_EXPIRED、STORAGE_FAILED。

每个请求必须有 request_id。

---

## 9. 数据库

每个用户拥有表必须有 user 所有权关系。

用户资产优先软删除。

文件在数据库生命周期规则确认不再引用前不物理删除。

以下操作使用事务：任务完成、资产创建、版本创建、项目移动、API 用量记录。

---

## 10. API 设计

公开 ImageForge API 必须保持 provider 独立。

不在公开 API 合同中暴露 Sub2API 特有概念。

Bad：`provider_account_id`、`sub2api_account_id`
Good：`model`、`generation_id`、`task_id`

---

## 11. 前端

UI 必须是：极简、图片优先、留白、专业、安静、快速。

避免：过度渐变、过度卡片、过度边框、AI 霓虹美学、过大装饰元素、无意义仪表盘、密集企业风格。

图片是第一视觉元素。

---

## 12. UX

每次生成必须传达状态：idle → submitting → processing → completed，或 idle → submitting → processing → failed → retry。

绝不让用户盯着无解释的转圈。

---

## 13. Prompt UX

Prompt 输入必须支持：多行、粘贴、历史、模板、重新生成、编辑前次 prompt。

不无故截断用户 prompt。

---

## 14. 资产 UX

每张生成图片自动成为 Asset。

Asset 可属于：项目、收藏、标签、生成任务。

绝不强制用户手动保存生成图片。

---

## 15. 版本管理

图片编辑创建新版本，绝不覆盖原图。

版本关系必须可追溯。

---

## 16. API Key

API Key 是敏感凭证。

创建时仅显示一次完整密钥，之后显示掩码。

支持：创建、撤销、重命名、最后使用时间、创建时间。

绝不记录完整 API Key。

---

## 17. 日志

日志绝不包含：API Key、refresh token、OAuth 凭证、cookie、授权头、私有图片数据、完整签名 URL。

使用 request_id 做关联。

---

## 18. 测试

每个功能必须有适当的测试：单元测试、集成测试、API 测试。

关键图片路径必须测试：success、timeout、429、500、502、503、invalid response、provider failure、storage failure。

---

## 19. 编码前检查清单

对每个任务：

1. 检查仓库（先查 Sub2API，再查 ImageForge）
2. 找到相关实现
3. 识别依赖
4. 写简短实施计划
5. 实施最小正确变更
6. 运行测试
7. 检查 diff
8. 删除不必要的变更
9. 报告确切改动

---

## 20. 不过度工程

不引入：microservices、event buses、Kubernetes、不必要的消息队列、复杂抽象。

除非当前规模或架构明确要求。

MVP 优先模块化单体。

---

## 21. 产品原则

在技术优雅和图片工作流用户体验之间，优先用户体验。

产品必须感觉像专业创作工具，而非基础设施控制台。

---

## 22. 合规门控

不要假设"技术上能调用上游订阅"等于"可以商业化转售"。

生产环境上线前必须：

1. 验证适用上游条款
2. 验证账户类型授权
3. 验证商业化使用权
4. 验证转售/API 访问权
5. 记录结果

法律/商业状态不明确时，明确标记风险，绝不静默假设许可。

---

## 23. 完成定义

代码编译不等于完成。

完成意味着：实现完成、错误状态处理、加载状态处理、空状态处理、权限检查实现、测试通过、数据库迁移（如需）、文档更新、UI 响应式、无密钥暴露、无多余变更。
