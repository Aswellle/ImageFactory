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
		field.Time("last_login_at").Optional().Nillable(),
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
