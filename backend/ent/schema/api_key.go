package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// APIKey is a programmatic credential. Only the hash is stored; the plaintext
// secret is shown once at creation and never persisted.
type APIKey struct {
	ent.Schema
}

func (APIKey) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").Optional(),
		field.String("key_hash").Unique().NotEmpty(), // sha256 of the secret
		field.String("key_prefix").NotEmpty(),         // first 8 chars for display
		field.String("name").Optional(),
		field.Enum("status").Values("active", "revoked").Default("active"),
		field.Int("permissions").Default(0), // bitmap; extend via migration
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("expires_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (APIKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("api_keys").Unique().Field("user_id"),
	}
}

func (APIKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key_hash"),
		index.Fields("user_id"),
	}
}
