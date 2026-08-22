package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Favorite is a join table recording which assets a user has favorited.
// Relationships are managed at the application layer via user_id / asset_id.
type Favorite struct {
	ent.Schema
}

func (Favorite) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("asset_id"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (Favorite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "asset_id").Unique(),
		index.Fields("user_id"),
		index.Fields("asset_id"),
	}
}
