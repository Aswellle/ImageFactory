package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UsageRecord is a lightweight accounting row per billed event.
type UsageRecord struct {
	ent.Schema
}

func (UsageRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("api_key_id").Optional(), // set when usage is via API key

		field.Enum("type").Values("generation", "edit", "upload", "storage"),
		field.String("model").Optional(),
		field.Int("image_count").Default(1),
		field.Int("tokens").Optional(), // for text/vision hybrid models

		// Cost (NUMERIC(20,10) maps to decimal at app layer).
		field.Float("cost").Default(0),

		field.String("request_id").Optional().Unique(), // idempotency / trace
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (UsageRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("api_key_id"),
		index.Fields("request_id"),
	}
}
