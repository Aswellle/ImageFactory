package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PromptTemplate stores reusable prompt patterns with variable placeholders.
type PromptTemplate struct {
	ent.Schema
}

func (PromptTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("content").NotEmpty(),   // template body with {{vars}}
		field.String("variables").Optional(), // comma-separated var names
		field.Enum("category").Values("product", "scene", "style", "custom").Default("custom"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PromptTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("prompt_templates").Unique().Required().Field("user_id"),
	}
}

func (PromptTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
