package main

import (
	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/server"
	"go.uber.org/zap"
)

// InitializeApplication builds the application and a cleanup func.
//
// NOTE: This is a hand-written initializer for Phase 1 Foundation so the app
// boots without Wire code-generation. When the dependency graph grows (Phase 2+),
// replace the body with a google/wire injector (wire.go wireinject + wire_gen.go).
func InitializeApplication(cfg *config.Config, log *zap.Logger) (*app, func(), error) {
	router, err := server.NewRouter(cfg, log)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		// Close DB, Redis, worker pools here in later phases.
		log.Info("cleanup complete")
	}
	return &app{Router: router}, cleanup, nil
}
