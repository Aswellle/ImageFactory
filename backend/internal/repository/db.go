package repository

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/imageforge/imageforge/ent"
)

// NewEntClient opens a PostgreSQL connection and returns an Ent client.
// When migrate=true, Ent auto-migrate runs (use sparingly; prefer versioned
// SQL migrations via MigrateUp in production).
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
