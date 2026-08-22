package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GenerationInput records the source image(s) for an edit job.
type GenerationInput struct {
	ent.Schema
}

func (GenerationInput) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("job_id"),
		field.Int64("source_asset_id").Optional(), // existing asset used as input
		field.String("storage_key").Optional(),  // or a freshly uploaded source
		field.String("role").Optional(),         // "reference", "mask", etc.
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (GenerationInput) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", GenerationJob.Type).Ref("inputs").Unique().Required().Field("job_id"),
	}
}

func (GenerationInput) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("job_id"),
	}
}
