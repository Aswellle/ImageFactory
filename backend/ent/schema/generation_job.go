package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GenerationJob models an async image-generation or edit task. All image work
// flows through jobs so the UI can show status, retry, and cancel.
type GenerationJob struct {
	ent.Schema
}

func (GenerationJob) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
	field.String("external_id").Optional().Unique(),
		field.Int64("project_id").Optional(),

		field.Enum("type").Values("generation", "edit").Immutable(),

		field.Enum("status").
			Values("pending", "processing", "completed", "failed", "cancelled").
			Default("pending"),

		field.String("provider").Optional(), // e.g. "openai", "sub2api"
		field.String("model").Optional(),
		field.String("sub2api_task_id").Optional(),

		// Request parameters (provider-independent).
		field.String("prompt").Optional(),
		field.String("negative_prompt").Optional(),
		field.String("aspect_ratio").Optional(),
		field.Int("image_count").Default(1),
		field.String("output_format").Optional(), // png, jpeg, webp

		// Failure preservation — never silently swallow.
		field.String("error_code").Optional(),
		field.String("error_message").Optional(),
		field.Int("retry_count").Default(0),

		// Batch image state — JSONB snapshot of the batch-image job/item state
		// for persistence across restarts. Null for non-batch jobs.
		field.JSON("batch_image_state", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		// Timing.
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (GenerationJob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("generation_jobs").Unique().Required().Field("user_id"),
		edge.From("project", Project.Type).Ref("generation_jobs").Unique().Field("project_id"),
		edge.To("inputs", GenerationInput.Type),
		edge.To("output_assets", Asset.Type),
	}
}

func (GenerationJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status"),
		index.Fields("status", "created_at"),
		index.Fields("sub2api_task_id"),
	}
}
