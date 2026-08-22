package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Asset is a produced or uploaded image. Original binaries live in object
// storage; the DB stores only keys, dimensions, and metadata.
type Asset struct {
	ent.Schema
}

func (Asset) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("project_id").Optional(),
		field.Int64("generation_job_id").Optional(),

		field.Enum("source").Values("generated", "uploaded", "edited").Immutable(),
		field.Enum("status").Values("active", "deleted").Default("active"),

		field.String("title").Optional(),
		field.String("description").Optional(),

		field.String("prompt").Optional(),
		field.String("negative_prompt").Optional(),

		field.String("model").Optional(),
		field.String("model_provider").Optional(),

		field.Int("width").Optional(),
		field.Int("height").Optional(),
		field.String("aspect_ratio").Optional(),

		field.String("mime_type").Optional(),
		field.Int64("file_size").Optional(),

		// storage keys (not URLs). Originals + thumbnails are separate objects.
		field.String("storage_key").NotEmpty().Unique(),
		field.String("thumbnail_key").Optional(),
		field.String("medium_key").Optional(),

		field.Int("current_version").Default(1),

		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Asset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("assets").Unique().Required().Field("user_id"),
		edge.From("project", Project.Type).Ref("assets").Unique().Field("project_id"),
		edge.From("generation_job", GenerationJob.Type).Ref("output_assets").Unique().Field("generation_job_id"),
		edge.To("versions", AssetVersion.Type),
		edge.To("tags", Tag.Type).Through("asset_tags", AssetTag.Type),
	}
}

func (Asset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "status"),
		index.Fields("project_id"),
		index.Fields("storage_key"),
	}
}
