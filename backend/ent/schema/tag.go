package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Tag is a user-owned label for organizing assets.
type Tag struct {
	ent.Schema
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("name").NotEmpty(),
		field.String("color").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("tags").Unique().Required().Field("user_id"),
		edge.From("assets", Asset.Type).Ref("tags").Through("asset_tags", AssetTag.Type),
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "name").Unique(),
	}
}
