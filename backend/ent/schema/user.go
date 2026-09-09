package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds ImageForge user identity. Auth is owned by ImageForge; Sub2API
// owns the upstream credentials.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique().NotEmpty(),
		field.String("password_hash").NotEmpty().Sensitive(),
		field.String("name").Optional(),
		field.Enum("role").Values("user", "admin").Default("user"),
		field.Enum("status").Values("active", "suspended", "deleted").Default("active"),
		field.Int("token_version").Default(0),
		// 2FA / TOTP
		field.String("totp_secret").Optional().Sensitive(), // 加密存储的 TOTP 密钥
		field.Bool("totp_enabled").Default(false),
		field.String("totp_backup_codes").Optional().Sensitive(), // 备用恢复码（JSON 数组）
		// 登录异常检测
		field.String("last_login_ip").Optional(),
		field.String("last_login_country").Optional(),
		field.Time("last_login_at").Optional().Nillable(),
		field.Int("failed_login_count").Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("api_keys", APIKey.Type),
		edge.To("projects", Project.Type),
		edge.To("assets", Asset.Type),
		edge.To("prompt_templates", PromptTemplate.Type),
		edge.To("generation_jobs", GenerationJob.Type),
		edge.To("collections", Collection.Type),
		edge.To("favorites", Favorite.Type),
		edge.To("tags", Tag.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("status"),
	}
}
