package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AssetVersion records an edit of an asset without overwriting the original.
// The original asset is immutable; each edit creates a new version row.
type AssetVersion struct {
	ent.Schema
}

func (AssetVersion) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("asset_id"),
		field.Int("version").Positive(), // 1-based, increments per asset
		field.Int64("edit_job_id").Optional(),

		field.String("prompt").Optional(),
		field.String("negative_prompt").Optional(),
		field.String("model").Optional(),

		field.Int("width").Optional(),
		field.Int("height").Optional(),

		field.String("mime_type").Optional(),
		field.Int64("file_size").Optional(),

		field.String("storage_key").NotEmpty().Unique(),
		field.String("thumbnail_key").Optional(),

		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (AssetVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("asset", Asset.Type).Ref("versions").Unique().Required().Field("asset_id"),
	}
}

func (AssetVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("asset_id", "version").Unique(),
		index.Fields("storage_key"),
	}
}
