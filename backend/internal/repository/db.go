package repository

import (
	"context"
	"fmt"

	"github.com/imageforge/imageforge/ent"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewEntClient opens a PostgreSQL connection and returns an Ent client.
// Auto-migrate runs schema migration when migrate=true (use sparingly; prefer
// versioned SQL migrations in production).
func NewEntClient(dsn string, migrate bool) (*ent.Client, error) {
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if migrate {
		if err := client.Schema.Create(context.Background()); err != nil {
			client.Close()
			return nil, fmt.Errorf("schema migration: %w", err)
		}
	}
	return client, nil
}
