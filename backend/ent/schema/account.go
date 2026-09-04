package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Account 定义 AI API 账号实体的 schema。
//
// 账户是图像生成的核心资源，代表一个可用于调用 AI 图像 API 的凭证。
// 例如：一个 Gemini API 账户、一个 OpenAI DALL-E 账户等。
//
// ImageForge 从 Sub2API 移植了核心的账号调度能力，但做了简化：
// - 聚焦图像生成场景，去除与图像无关的字段
// - 保留账号选择、速率限制、过载保护等核心调度逻辑
// - 通过 credentials 字段灵活存储不同平台的认证信息
type Account struct {
	ent.Schema
}

// Fields 定义账户实体的所有字段。
func (Account) Fields() []ent.Field {
	return []ent.Field{
		// name: 账户显示名称，用于在管理界面中标识账户
		field.String("name").
			MaxLen(100).
			NotEmpty(),

		// platform: 所属平台，如 "gemini", "openai" 等
		field.String("platform").
			MaxLen(50).
			NotEmpty(),

		// type: 认证类型，如 "api_key", "oauth" 等
		// 不同类型决定了 credentials 中存储的数据结构
		field.String("type").
			MaxLen(20).
			NotEmpty(),

		// credentials: 认证凭证，以 JSONB 格式存储
		// 结构取决于 type 字段：
		// - api_key: {"api_key": "sk-xxx"}
		// - oauth: {"access_token": "...", "refresh_token": "..."}
		field.JSON("credentials", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),

		// priority: 账户优先级，数值越小优先级越高
		// 调度器会优先使用高优先级的账户
		field.Int("priority").
			Default(50),

		// status: 账户状态
		field.Enum("status").
			Values("active", "error", "disabled").
			Default("active"),

		// error_message: 错误信息，记录账户异常时的详细信息
		field.String("error_message").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// last_used_at: 最后使用时间，用于轮转调度
		field.Time("last_used_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// expires_at: 账户过期时间（可为空）
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// schedulable: 是否可被调度器选中
		field.Bool("schedulable").
			Default(true),

		// rate_limited_at: 触发速率限制的时间
		field.Time("rate_limited_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// rate_limit_reset_at: 速率限制预计解除的时间
		field.Time("rate_limit_reset_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// overload_until: 过载状态解除时间
		field.Time("overload_until").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// created_at: 创建时间
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// updated_at: 更新时间
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}


// Edges 定义账户实体的关联关系。
func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		// 账户的使用日志
		edge.To("usage_logs", UsageRecord.Type),
	}
}

// Indexes 定义数据库索引，优化查询性能。
func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform"),
		index.Fields("status"),
		index.Fields("schedulable"),
		index.Fields("priority"),
		// 复合索引：按平台查询可调度账号（调度器最常用查询）
		index.Fields("platform", "schedulable", "status"),
	}
}
