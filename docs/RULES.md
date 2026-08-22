# ImageForge 不可违反规则

你正在开发的是 ImageForge——一个以 Sub2API 本地代码库为基座的商业图片生产工作台。

## 核心原则

**Sub2API 是基座，不是外部服务。**

ImageForge 直接复用 Sub2API 的源代码、开发模式、处理逻辑。最终目标是 ImageForge 单独部署即可提供完整商业化服务。

---

## 必须立即停止并重新评估的情况

1. 开始重复造轮子（Sub2API 已有实现的功能，不要自己重写）
2. 开始把 Sub2API 当外部 HTTP 服务调用（应直接复用代码）
3. 开始修改 Sub2API 源文件（只复制适配，不改源）
4. 开始实现 OAuth/upstream account scheduler/provider protocol（Sub2API 已有，直接移植）
5. 开始把 API Key、refresh token 暴露到前端
6. 开始把图片二进制直接存进数据库
7. 开始把生成过程设计成永久同步 HTTP 请求
8. 开始为了"漂亮"加入大量渐变、玻璃拟态和装饰
9. 开始为了未来可能需求创建大量抽象层
10. 开始修改与当前任务无关的文件
11. 开始大规模重构已有代码
12. 开始假设不存在的 API
13. 开始假设不存在的数据库字段
14. 开始绕过测试
15. 开始把上游错误直接返回给用户
16. 开始忽略 Go workspace 的 ent 直接导入能力

---

## 正确行为

**先搜索。先搜 Sub2API，再搜 ImageForge。**
**再阅读。理解实现和数据流。**
**再计划。确定复用还是新建。**
**最后修改。**

---

## 决策优先级

如果已有实现可以复用：
**REUSE > REWRITE**
> 先查 Sub2API 有没有。有 → 复制移植。没有 → 新建。

如果不确定：
**INSPECT > GUESS**
> 先读代码确认，不猜。

如果需求不明确：
**ASK > ASSUME**
> 先问清楚，不假设。

如果修改范围过大：
**MINIMIZE > REFACTOR**
> 最小变更，不大改。

如果出现安全问题：
**STOP > SHIP**
> 停下来，不上线。

如果 Sub2API 有现成实现：
**PORT > REIMPLEMENT**
> 移植适配，不重新实现。

---

## 代码复用清单

开发新功能前，按此清单检查：

- [ ] Sub2API ent/ 有现成数据模型？→ 直接 import
- [ ] Sub2API internal/ 有现成服务？→ 复制源文件移植
- [ ] Sub2API 有现成处理逻辑？→ 复用模式
- [ ] Sub2API handler 有现成路由？→ 复制适配
- [ ] 以上都没有？→ 在 ImageForge 新建

---

## 移植 Sub2API 代码的标准流程

1. **复制** — 源文件复制到 ImageForge 对应目录
2. **改包名** — `package service` → ImageForge 目标包名
3. **换导入** — `github.com/Wei-Shaw/sub2api/internal/...` → ImageForge 本地包
4. **去 internal 依赖** — 替换为本地 stub 或 ent 类型
5. **编译验证** — `go build ./...`
6. **测试验证** — `go test ./...`
