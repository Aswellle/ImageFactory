package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AssetTag is the many-to-many join between Asset and Tag.
type AssetTag struct {
	ent.Schema
}

func (AssetTag) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("asset_id"),
		field.Int64("tag_id"),
	}
}

func (AssetTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("asset", Asset.Type).Unique().Required().Field("asset_id"),
		edge.To("tag", Tag.Type).Unique().Required().Field("tag_id"),
	}
}

func (AssetTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("asset_id", "tag_id").Unique(),
	}
}
