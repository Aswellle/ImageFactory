package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Project groups assets into a commercial work context (e.g. "Nike summer ads").
type Project struct {
	ent.Schema
}

func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.Enum("status").Values("active", "archived").Default("active"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("projects").Unique().Required().Field("user_id"),
		edge.To("assets", Asset.Type),
		edge.To("generation_jobs", GenerationJob.Type),
	}
}

func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "status"),
	}
}
