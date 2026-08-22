package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Collection is a user-curated group of assets (e.g. "favorites", "Q3 campaign").
type Collection struct {
	ent.Schema
}

func (Collection) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Collection) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("collections").Unique().Required().Field("user_id"),
		edge.To("assets", Asset.Type),
	}
}

func (Collection) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
