# 一、项目总体定义

## 项目暂定名称

建议先使用：

**ImageForge**

中文定位：

> 面向商业场景的 AI 图片生成与创意资产工作台

也可以后续换成：

+ CanvasAI
+ ImageFlow
+ PixelForge
+ VisualForge
+ ArtFlow
+ BrandCanvas
+ ImageStudio

其中我更推荐 **ImageForge** 或 **VisualForge**。

------

# 二、产品核心定位

不要把产品定位成：

> ChatGPT 生图 API 中转站

而应该定位成：

> **商业图片 AI 生产工作台**

用户进入网站后首先看到的不是：

+ API Key
+ Endpoint
+ Token
+ RPM
+ Account
+ Provider
+ Subscription

而应该看到：

+ 创建图片
+ 图片编辑
+ 我的作品
+ 项目
+ 品牌素材
+ 提示词
+ 最近生成
+ 收藏
+ API

也就是说：

> **API 是基础设施，图片生产才是产品。**

------

# 三、总体系统架构

ImageForge 以 **Sub2API 本地代码库为基座**构建。Sub2API 不是外部服务，而是本项目的底层引擎代码库。ImageForge 通过 Go workspace 直接复用 Sub2API 的 Ent 数据模型，并通过代码移植复用其图片管道、账户调度、计费、故障转移等核心能力。

**最终目标：ImageForge 单独部署上线即可提供完整商业化服务，无需依赖外部 API（除必要的 AI 上游调用外）。**

建议采用：

```text
                         ┌──────────────────────┐
                         │      用户浏览器       │
                         └──────────┬───────────┘
                                    │
                                    ▼
                     ┌──────────────────────────┐
                     │     ImageForge Web       │
                     │                          │
                     │  Image Workspace         │
                     │  Asset Library           │
                     │  Projects                │
                     │  Prompt Templates        │
                     │  API Console             │
                     │  User / Settings         │
                     └────────────┬─────────────┘
                                  │
                                  ▼
                     ┌──────────────────────────────────────────┐
                     │         ImageForge Backend               │
                     │                                          │
                     │  ┌─────────────────────────────────┐    │
                     │  │  产品层（ImageForge 独有）       │    │
                     │  │  Auth / Projects / Assets       │    │
                     │  │  Generation Jobs / API Keys     │    │
                     │  │  Usage / Metadata               │    │
                     │  └─────────────────────────────────┘    │
                     │                                          │
                     │  ┌─────────────────────────────────┐    │
                     │  │  引擎层（移植自 Sub2API）        │    │
                     │  │  Batch Image Pipeline           │    │
                     │  │  Account Scheduling             │    │
                     │  │  Model Routing / Failover       │    │
                     │  │  Billing / Worker Queue         │    │
                     │  └─────────────────────────────────┘    │
                     └────────────┬─────────────────────────────┘
                                  │
                         OpenAI-compatible
                                  │
                                  ▼
                     ┌──────────────────────────┐
                     │      Upstream Provider    │
                     │   Authorized AI Service   │
                     └──────────────────────────┘
```

Sub2API 本地代码库已实现完整的图片生成管道（`/v1/images/generations`、`/v1/images/edits`、异步任务、批处理、计费、故障转移），ImageForge 直接移植复用这些底层能力，不重新实现。

------

# 四、产品信息架构

一级导航建议控制在 6 个以内：

```text
ImageForge

├── 工作台
│
├── 创作
│   ├── 文生图
│   └── 图片编辑
│
├── 作品
│   ├── 全部
│   ├── 项目
│   ├── 收藏
│   └── 回收站
│
├── 素材
│   ├── 上传素材
│   ├── 品牌素材
│   └── 参考图
│
├── 提示词
│   ├── 我的提示词
│   └── 模板
│
└── API
    ├── API Keys
    ├── API 文档
    └── 调用记录
```

管理员后台则完全独立：

```text
Admin

├── Dashboard
├── Users
├── Accounts
├── Models
├── Image Jobs
├── Assets
├── API Keys
├── Usage
├── System
└── Logs
```

------

# 五、核心 UX 原则

## 5.1 Apple 风格，但不要做成“苹果官网复制品”

设计关键词：

```text
Minimal
Clean
Premium
Quiet
Spatial
Editorial
Image-first
Professional
```

视觉上应该类似：

> Apple + Linear + Raycast + Arc + 专业创意工具

而不是：

> SaaS 后台 + 大量卡片 + 渐变按钮 + AI 紫色霓虹。

------

# 六、视觉规范

## 6.1 色彩

主色：

```text
Background:
#F5F5F7

Surface:
#FFFFFF

Primary Text:
#1D1D1F

Secondary Text:
#6E6E73

Border:
#E5E5E7

Accent:
#0071E3
```

深色模式：

```text
Background:
#000000

Surface:
#1C1C1E

Primary:
#F5F5F7

Secondary:
#98989D
```

原则：

**禁止默认 AI SaaS 紫蓝渐变。**

图片本身承担视觉色彩。

------

# 七、核心页面 PRD

## 7.1 首页 / 工作台

目标：

> 用户打开网站后 3 秒内知道“这里就是用来做商业图片的”。

布局：

```text
┌─────────────────────────────────────────────┐
│ ImageForge                  作品  API  头像 │
├─────────────────────────────────────────────┤
│                                             │
│          创作你的下一张商业图片              │
│                                             │
│   描述你想要的画面、风格或修改方式……         │
│                                             │
│   ┌─────────────────────────────────────┐   │
│   │                                     │   │
│   │  Prompt                              │   │
│   │                                     │   │
│   │                        [生成图片 →]  │   │
│   └─────────────────────────────────────┘   │
│                                             │
│   文生图    图片编辑    参考图    模板       │
│                                             │
├─────────────────────────────────────────────┤
│ 最近作品                                    │
│                                             │
│ [ image ] [ image ] [ image ] [ image ]    │
│                                             │
└─────────────────────────────────────────────┘
```

------

# 八、创作工作台

这是整个项目的**核心页面**。

建议：

```text
┌─────────────────────────────────────────────────┐
│ ← 创作                          保存     导出     │
├─────────────────────────────────────────────────┤
│                                                 │
│                图片预览区域                     │
│                                                 │
│                Generated Image                  │
│                                                 │
├──────────────────────┬──────────────────────────┤
│ Prompt               │ 参数                     │
│                      │                          │
│ 描述你想要的图片...  │ 模型                     │
│                      │ 尺寸                     │
│                      │ 质量                     │
│                      │ 比例                     │
│                      │ 输出格式                 │
│                      │                          │
│ [参考图]             │ [生成 4 张]              │
└──────────────────────┴──────────────────────────┘
```

------

# 九、文生图功能

必须支持：

### Prompt

```text
prompt
```

限制：

+ 前端实时字数统计
+ 支持长 Prompt
+ 支持粘贴
+ 支持 Prompt 模板
+ 支持最近 Prompt
+ 支持重新生成

------

## 图片数量

```text
1
2
4
```

如果底层能力允许，可以继续扩展。

------

## 比例

第一阶段：

```text
1:1
16:9
9:16
4:3
3:4
```

UI 不直接显示：

```text
1024x1024
1536x1024
```

而显示：

```text
正方形
横版
竖版
标准横图
标准竖图
```

高级设置再展示具体分辨率。

------

# 十、图片编辑

这是商业场景非常重要的功能。

用户可以：

```text
上传图片
+
输入修改要求
=
新图片
```

典型操作：

```text
替换背景
删除物体
增加物体
改变颜色
修改服装
调整构图
商业产品摄影
广告图优化
电商主图
社交媒体配图
```

UI：

```text
原图                         编辑后

┌──────────────┐             ┌──────────────┐
│              │             │              │
│   Original   │     →       │   Result     │
│              │             │              │
└──────────────┘             └──────────────┘

Prompt:

“将背景替换成高级白色摄影棚，
保持产品本身形状和材质不变……”

             [开始编辑]
```

------

# 十一、图片生成任务系统

不要把图片生成做成普通同步 HTTP 请求。

应该采用：

```text
Create Job
    ↓
pending
    ↓
processing
    ↓
completed
    ↓
asset created
```

异常：

```text
pending
 ↓
processing
 ↓
failed
```

需要支持：

```text
retry
cancel
duplicate
regenerate
edit
```

Sub2API 当前已经提供异步 Image Tasks，适合长耗时图片任务；其异步接口通过 task_id 轮询最终结果，并且要求对象存储能力。([GitHub](https://github.com/Wei-Shaw/sub2api/blob/main/docs/ASYNC_IMAGE_TASKS.md?utm_source=chatgpt.com))

因此 ImageForge 应优先采用：

```text
ImageForge Job
        ↓
Sub2API Async Image Task
        ↓
poll / callback
        ↓
ImageForge Asset
```

而不是让浏览器一直等待上游 HTTP 请求。

------

# 十二、作品库

这是第二核心模块。

不要设计成传统文件管理器。

应该采用：

> **Pinterest / Apple Photos / Linear 风格的视觉资产库**

默认：

```text
全部作品

[图片] [图片] [图片]
[图片] [图片] [图片]
[图片] [图片] [图片]
```

图片使用 Masonry / responsive grid。

------

# 十三、作品卡片

Hover：

```text
┌─────────────────┐
│                 │
│      IMAGE      │
│                 │
│   ♡       ⋯     │
└─────────────────┘
```

点击进入：

```text
图片
Prompt
Model
Size
Created At
Project
Tags
Generation ID
```

操作：

```text
重新生成
编辑
下载
复制
收藏
移动到项目
删除
```

------

# 十四、项目管理

商业用户通常不是“生成一张图片”，而是：

> 做一个项目。

例如：

```text
Nike 夏季广告
淘宝店铺主图
小红书宣传图
品牌官网 Banner
产品包装
公众号配图
```

因此：

```text
Projects
```

是必须的。

项目：

```text
Project
 ├── Images
 ├── References
 ├── Prompts
 ├── Versions
 └── Metadata
```

------

# 十五、版本管理

这是商业图片工具非常值得加入的功能。

例如：

```text
产品主图 v1
       ↓
产品主图 v2
       ↓
产品主图 v3
       ↓
最终版本
```

不能简单覆盖原图。

数据库应该保存：

```text
asset
asset_version
```

形成：

```text
Asset
 ├── Version 1
 ├── Version 2
 ├── Version 3
 └── Current Version
```

------

# 十六、素材系统

素材分为：

```text
Reference
Brand Asset
Generated Asset
Uploaded Asset
```

例如：

```text
品牌 Logo
产品照片
人物照片
包装图
字体
参考图片
历史作品
```

------

# 十七、标签系统

至少支持：

```text
项目
标签
收藏
创建日期
模型
生成类型
```

例如：

```text
#电商
#产品图
#小红书
#广告
#Banner
#人物
#摄影
#品牌
```

支持：

```text
搜索
筛选
排序
```

------

# 十八、API 产品设计

API 页面必须和管理后台彻底分离。

用户看到的是：

```text
API

Your API

sk-xxxxxxxxxxxxxxxx

[复制]

Base URL

https://api.example.com/v1
```

------

## API Endpoint

第一阶段：

```http
POST /v1/images/generations

POST /v1/images/edits

GET /v1/images/tasks/{id}

GET /v1/models
```

底层通过 Sub2API 对接。

Sub2API 当前已经提供 OpenAI Images 兼容接口，因此你的 API 层主要负责**权限、用户隔离、资产归档、业务 metadata、配额与产品级抽象**。([GitHub](https://github.com/Ember1012/sub2api/blob/main/docs/API.md?utm_source=chatgpt.com))

------

# 十九、不要直接暴露 Sub2API

这是一个非常重要的架构原则。

错误：

```text
User
 ↓
Sub2API
```

正确：

```text
User
 ↓
ImageForge API
 ↓
ImageForge Auth
 ↓
Quota
 ↓
Job
 ↓
Sub2API
 ↓
Upstream
```

否则以后：

+ 更换 Sub2API
+ 接入其他模型
+ 接入其他图片 Provider
+ 调整价格
+ 增加商业工作流

都会非常困难。

------

# 二十、数据库设计

建议至少：

```text
users

projects

assets

asset_versions

asset_tags

tags

generation_jobs

generation_inputs

generation_outputs

prompt_templates

api_keys

api_requests

usage_records

model_configs

collections

favorites
```

核心关系：

```text
User
 ├── Projects
 ├── Assets
 ├── API Keys
 ├── Prompt Templates
 └── Generation Jobs

Project
 └── Assets

Asset
 ├── AssetVersions
 ├── Tags
 └── GenerationJob

GenerationJob
 ├── Inputs
 └── Outputs
```

------

# 二十一、Asset 数据模型

核心字段建议：

```text
id
user_id
project_id

type
source
status

title
description

prompt
negative_prompt

model
model_provider

width
height
aspect_ratio

mime_type
file_size

storage_key
thumbnail_key

generation_job_id

created_at
updated_at
deleted_at
```

------

# 二十二、Generation Job 数据模型

```text
id

user_id
project_id

type
    generation
    edit

status
    pending
    processing
    completed
    failed
    cancelled

provider
sub2api_task_id

model

prompt

input_assets

parameters

output_assets

error_code
error_message

started_at
completed_at

created_at
updated_at
```

------

# 二十三、Storage 架构

不要把图片 Blob 全部放数据库。

采用：

```text
Object Storage
```

例如：

```text
S3
Cloudflare R2
MinIO
OSS
```

目录：

```text
/users/{user_id}/
    projects/{project_id}/
        assets/
        thumbnails/
        originals/
        versions/
```

------

# 二十四、缩略图系统

上传/生成图片以后：

```text
Original
   ↓
Metadata Extraction
   ↓
Thumbnail Generation
   ↓
Preview
   ↓
Object Storage
```

至少生成：

```text
thumbnail
medium
original
```

避免图库加载原图。

------

# 二十五、后台任务

建议：

```text
Web API
   ↓
Queue
   ↓
Worker
   ↓
Sub2API
```

任务：

```text
image_generation
image_edit
thumbnail
metadata
asset_cleanup
storage_cleanup
```

------

# 二十六、错误处理

必须统一错误结构：

```json
{
  "error": {
    "code": "IMAGE_GENERATION_FAILED",
    "message": "Image generation failed",
    "request_id": "req_xxx"
  }
}
```

不要直接把：

```text
Sub2API
OpenAI
upstream
OAuth
```

的内部错误原样暴露给用户。

------

# 二十七、可靠性设计

尤其需要考虑：

```text
429
500
502
503
504
timeout
upstream server_error
account unavailable
authentication expired
```

Sub2API 的图像链路历史上已经出现过上游 image `server_error` 与 failover 相关问题，因此 Codex 在开发时必须把**真实上游失败 → Sub2API → ImageForge Job**整条链路纳入集成测试，而不能只测试 HTTP 200。相关问题后来已有对应修复/关闭记录，但它说明 image pipeline 的错误语义必须单独验证。([GitHub](https://github.com/Wei-Shaw/sub2api/issues/3130?utm_source=chatgpt.com))

------

# 二十八、安全设计

必须做到：

### API Key

数据库：

```text
hash(api_key)
```

而不是：

```text
plain_text_api_key
```

只在创建时显示一次。

------

### 用户隔离

所有查询必须包含：

```text
user_id
```

禁止：

```text
GET /assets/:id
```

直接返回。

必须：

```text
authenticate
↓
authorize
↓
resource ownership
↓
return
```

------

# 二十九、图片访问安全

不要永久暴露：

```text
https://storage.xxx/user/xxx/image.png
```

推荐：

```text
Signed URL
```

并设置：

```text
expiration
```

------

# 三十、商业图片导出

图片详情页：

```text
下载

PNG
JPG
WebP
```

未来：

```text
2x
4x
Upscale
Remove Background
```

都可以变成高级工具。

------

# 三十一、Dashboard

不要做传统“数据大屏”。

展示：

```text
本月生成
图片数量
项目数量
存储空间
API 调用
```

以及：

```text
最近作品
最近项目
最近 API 请求
```

------

# 三十二、管理员后台

管理员应该能：

```text
查看用户
禁用用户
查看任务
查看失败任务
查看 API 调用
查看模型
查看 Sub2API 状态
查看账号状态
查看错误日志
```

但是：

> Sub2API 管理功能不要重新复制。

管理员需要时直接进入 Sub2API Admin。

------

# 三十三、模型抽象

前端不要写死：

```text
gpt-image-2
```

而是：

```text
Image Model
```

数据库：

```text
model_id
display_name
provider
capabilities
supported_sizes
supported_formats
enabled
```

例如：

```text
gpt-image-2
    文生图 ✓
    图生图 ✓
    高清 ✓

future-model
    文生图 ✓
    图生图 ✓
```

这样以后可以加入其他图片模型。

------

# 三十四、Prompt Template

这是商业工具区别于普通 AI 生图页面的重要功能。

模板：

```text
电商产品摄影

你是一名专业商业摄影师。

产品：
{{product}}

场景：
{{scene}}

风格：
{{style}}

光线：
{{lighting}}

要求：
{{requirements}}
```

用户只需要填写变量。

------

# 三十五、商业工作流

第二阶段可以实现：

```text
产品图片
      ↓
生成白底图
      ↓
生成场景图
      ↓
生成广告图
      ↓
生成社媒 Banner
```

也就是：

> 从“AI 生图”升级成“AI 商业视觉生产”。

这是整个产品未来真正有商业价值的方向。

------

# 三十六、MVP 必须砍掉的东西

第一版**不要做**：

```text
在线支付
团队协作
复杂 RBAC
工作流编排
视频生成
音频
AI 聊天
Prompt marketplace
社区
社交系统
评论
点赞
复杂统计
```

MVP 只做：

```text
登录
↓
工作台
↓
文生图
↓
图生图
↓
作品库
↓
项目
↓
素材
↓
API
↓
基础用量
```

------

# 三十七、推荐开发阶段

## Phase 0 — Reverse Engineering & Reuse Mapping

Codex 第一阶段：

**禁止写代码。**

目标：识别 Sub2API 中可复用的代码、模式、逻辑，建立复用映射。

只允许：

```text
阅读 Sub2API 源码
阅读 package 结构
阅读 API 接口
阅读数据库 schema
阅读 image pipeline 实现
阅读 frontend 组件
```

输出：

```text
docs/sub2api-analysis.md      — Sub2API 能力清单与复用评估
docs/integration-map.md        — 复用映射：哪些直接 import、哪些复制移植、哪些新建
```

**复用原则：**
- Ent 数据模型 → 通过 Go workspace 直接 import
- internal/ 服务层 → 复制源文件到 ImageForge，适配包路径
- 开发模式/处理逻辑 → 遵循 Sub2API 约定
- ImageForge 独有功能 → 新建

------

## Phase 1 — Foundation

实现：

```text
项目初始化（Go workspace 链接 Sub2API）
数据库（直接复用 Sub2API ent/ 模型 + ImageForge 独有模型）
Auth（JWT，参考 Sub2API 中间件模式）
User（产品级用户，区别于 Sub2API 网关用户）
Storage（对象存储，参考 Sub2API image_storage）
API framework（Gin 路由 + 中间件链，参考 Sub2API）
Job framework（异步任务队列，移植自 Sub2API batch_image_worker）
```

**复用检查：** 每项实现前，先检查 Sub2API 是否有现成实现。有 → 移植。没有 → 新建。

------

## Phase 2 — Image Core

实现：

```text
Image Generation（移植 Sub2API batch_image 管道 + openai_images 网关）
Image Editing（移植 Sub2API 图片编辑能力）
Job（产品级生成任务，使用 Sub2API BatchImageJob ent 模型）
Asset（资产记录，ImageForge 独有）
Storage（对象存储，参考 Sub2API image_storage）
Thumbnail（缩略图生成）
```

**核心：** 图片生成引擎不是外部 API 调用，而是移植到 ImageForge 内部的 Sub2API 代码。

------

## Phase 3 — Workspace

实现：

```text
Dashboard
Create
Gallery
Project
Asset Detail
```

------

## Phase 4 — API

实现：

```text
API Key
API Gateway
Usage
Documentation
```

------

## Phase 5 — Polish

实现：

```text
Apple-like UI
动画
快捷键
响应式
Dark Mode
性能优化
错误体验
```

------

# 三十八、验收标准

## 生图

必须：

+  可以输入 Prompt
+  可以选择模型
+  可以选择尺寸
+  可以选择数量
+  可以生成
+  可以看到实时任务状态
+  失败可以重试
+  成功自动进入图库

------

## 图生图

必须：

+  上传图片
+  预览
+  输入编辑指令
+  生成
+  保存原图
+  保存结果
+  建立版本关系

------

## 图库

必须：

+  Masonry/Grid
+  分页/无限滚动
+  搜索
+  标签
+  收藏
+  下载
+  删除
+  项目归档

------

## API

必须：

+  创建 API Key
+  删除 API Key
+  API 鉴权
+  图片生成
+  图片编辑
+  Task 查询
+  Models
+  Usage
+  Request ID