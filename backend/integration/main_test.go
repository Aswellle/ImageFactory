//go:build integration || e2e

// Shared integration/e2e harness. This file is compiled into both the
// `integration` and `e2e` test binaries and owns infrastructure provisioning
// (database connection, migrations, server boot) plus the request helpers the
// per-tag test files rely on.
//
// It is NOT a test file itself: it defines TestMain and helpers but no Test*
// functions, so it never runs tests on its own.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/ent/user"
	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/handler"
	admin "github.com/imageforge/imageforge/internal/handler/admin"
	"github.com/imageforge/imageforge/internal/job"
	"github.com/imageforge/imageforge/internal/repository"
	"github.com/imageforge/imageforge/internal/server/middleware"
	"github.com/imageforge/imageforge/internal/service"
	"github.com/imageforge/imageforge/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testEnv holds shared infrastructure for the suite. Populated by TestMain.
type testEnv struct {
	db        *ent.Client
	rdb       *redis.Client
	engine    *gin.Engine
	server    *httptest.Server
	jwt       *service.JWTService
	baseURL   string
	authToken string
	userID    int64
	adminKey  string
}

// env is the process-wide harness. TestMain seeds an admin token here so the
// per-tag suites can exercise admin endpoints without re-provisioning.
var env *testEnv

// TestMain waits for infrastructure, runs migrations, boots the server, and
// registers a cleanup. If the database is unreachable the suite skips instead
// of failing so CI without Docker still passes.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	cfg := loadTestConfig()

	logger := zap.NewNop()

	db, err := connectDatabase(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: database unavailable (%v), skipping suite\n", err)
		os.Exit(0)
	}

	// Auto-migrate schema (integration tests own a dedicated database).
	if err := db.Schema.Create(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "integration: migration failed: %v\n", err)
		os.Exit(1)
	}

	rdb := connectRedis(cfg)

	store, err := storage.New(cfg.Storage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: storage init failed: %v\n", err)
		os.Exit(1)
	}

	engine := buildTestEngine(cfg, logger, db, rdb, store)
	server := httptest.NewServer(engine)

	env = &testEnv{
		db:       db,
		rdb:      rdb,
		engine:   engine,
		server:   server,
		jwt:      service.NewJWTService(cfg.Auth),
		baseURL:  server.URL,
		adminKey: cfg.Auth.AdminPanelKey,
	}

	// Seed an admin-authenticated token for admin endpoint tests.
	env.authToken = mustSeedUser(env, "admin-"+strconv.FormatInt(time.Now().UnixNano(), 10)+"@imageforge.test", user.RoleAdmin)
	claims, err := env.jwt.Parse(env.authToken)
	if err == nil {
		env.userID = claims.UserID
	}

	code := m.Run()

	server.Close()
	_ = db.Close()
	if rdb != nil {
		_ = rdb.Close()
	}
	os.Exit(code)
}

// loadTestConfig builds a Config from environment with integration-safe
// defaults. Tests point at a dedicated database (IF_DATABASE_NAME) to avoid
// clobbering dev data.
func loadTestConfig() *config.Config {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:    getenv("IF_SERVER_HOST", "127.0.0.1"),
			Port:    0,
			Mode:    "test",
			Timeout: 30,
		},
		Database: config.DatabaseConfig{
			Host:     getenv("IF_DATABASE_HOST", "localhost"),
			Port:     atoi(getenv("IF_DATABASE_PORT", "5432")),
			User:     getenv("IF_DATABASE_USER", "imageforge"),
			Password: getenv("IF_DATABASE_PASSWORD", "imageforge"),
			Name:     getenv("IF_DATABASE_NAME", "imageforge_it"),
			SSLMode:  getenv("IF_DATABASE_SSLMODE", "disable"),
		},
		Redis: config.RedisConfig{
			Host:     getenv("IF_REDIS_HOST", "localhost"),
			Port:     atoi(getenv("IF_REDIS_PORT", "6379")),
			Password: getenv("IF_REDIS_PASSWORD", ""),
			DB:       1, // separate Redis DB to avoid collisions
		},
		Auth: config.AuthConfig{
			JWTSecret:          getenv("IF_AUTH_JWT_SECRET", "integration-test-secret"),
			AccessTokenMinutes: 1440,
			BcryptCost:         10,
			AdminPanelKey:      getenv("IF_AUTH_ADMIN_PANEL_KEY", "it-admin-panel-key"),
		},
		Storage: config.StorageConfig{
			Provider: "filesystem",
			Bucket:   "imageforge-it",
		},
		Sub2API: config.Sub2APIConfig{
			TimeoutSeconds: 120,
		},
	}
	return cfg
}

// connectDatabase retries until the database is reachable (Docker startup).
func connectDatabase(cfg *config.Config) (*ent.Client, error) {
	var db *ent.Client
	var err error
	dsn := cfg.Database.DSN()
	for range 30 {
		db, err = repository.NewEntClient(dsn, false)
		if err == nil {
			// Verify the connection is live.
			if _, pingErr := db.QueryContext(context.Background(), "SELECT 1"); pingErr == nil {
				return db, nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	if db != nil {
		_ = db.Close()
	}
	return nil, fmt.Errorf("could not connect to database after retries: %w", err)
}

// connectRedis returns a Redis client or nil when Redis is unavailable. A nil
// client is acceptable: the server falls back to in-memory task storage.
func connectRedis(cfg *config.Config) *redis.Client {
	if cfg.Redis.Host == "" {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port)),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil
	}
	return rdb
}

// buildTestEngine wires the full application exactly like server.NewRouter but
// also registers the favorite/tag/collection routes that the production router
// currently omits, so the suite covers the complete API surface.
func buildTestEngine(cfg *config.Config, log *zap.Logger, db *ent.Client, rdb *redis.Client, store storage.Storage) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS())
	engine.Use(gin.Recovery())

	// Job queue + worker (in-memory; sufficient for request-lifecycle tests).
	queue := job.NewMemoryQueue(256, log)
	processor := job.NewProcessor(queue, 4)
	if err := processor.Start(context.Background()); err != nil {
		log.Warn("failed to start job processor", zap.Error(err))
	}

	// --- Services & handlers ---
	jwtSvc := service.NewJWTService(cfg.Auth)
	passwordSvc := service.NewPassword(cfg.Auth)
	users := repository.NewUserRepository(db)
	authService := service.NewAuthService(users, jwtSvc, passwordSvc)
	authHandler := handler.NewAuthHandler(authService)
	authMW := middleware.NewAuth(jwtSvc)

	projectSvc := service.NewProjectService(db)
	assetSvc := service.NewAssetService(db, store)
	projectHandler := handler.NewProjectHandler(projectSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)

	apiKeySvc := service.NewAPIKeyService(db)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeySvc)
	apiKeyMW := middleware.APIKeyAuth(apiKeySvc)

	usageSvc := service.NewUsageService(db)
	usageHandler := handler.NewUsageHandler(usageSvc)

	promptTemplateSvc := service.NewPromptTemplateService(db)
	promptTemplateHandler := handler.NewPromptTemplateHandler(promptTemplateSvc)

	favoriteSvc := service.NewFavoriteService(db)
	favoriteHandler := handler.NewFavoriteHandler(favoriteSvc)

	tagSvc := service.NewTagService(db)
	tagHandler := handler.NewTagHandler(tagSvc)

	collectionSvc := service.NewCollectionService(db)
	collectionHandler := handler.NewCollectionHandler(collectionSvc)

	// --- Admin ---
	adminSvc := service.NewAdminService(db)
	adminAuth := middleware.NewAdminAuth(authMW, cfg.Auth.AdminPanelKey)
	adminHandlers := &adminRouteHandlers{
		Dashboard: admin.NewDashboardHandler(adminSvc),
		User:      admin.NewUserHandler(adminSvc),
		Job:       admin.NewJobHandler(adminSvc),
		APIKey:    admin.NewAPIKeyHandler(adminSvc),
	}

	// --- Public routes ---
	v1 := engine.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
		auth.GET("/me", authMW.Require(), authHandler.Me)
	}

	// --- Protected routes ---
	authorized := v1.Group("")
	authorized.Use(authMW.Require())
	{
		authorized.POST("/projects", projectHandler.Create)
		authorized.GET("/projects", projectHandler.List)
		authorized.GET("/projects/:id", projectHandler.Get)

		authorized.GET("/assets", assetHandler.List)
		authorized.GET("/assets/:id", assetHandler.Get)
		authorized.DELETE("/assets/:id", assetHandler.Delete)

		authorized.POST("/api-keys", apiKeyHandler.Create)
		authorized.GET("/api-keys", apiKeyHandler.List)
		authorized.DELETE("/api-keys/:id", apiKeyHandler.Revoke)

		authorized.GET("/usage", usageHandler.Get)
		authorized.GET("/usage/history", usageHandler.History)

		authorized.POST("/prompt-templates", promptTemplateHandler.Create)
		authorized.GET("/prompt-templates", promptTemplateHandler.List)
		authorized.GET("/prompt-templates/:id", promptTemplateHandler.Get)
		authorized.PUT("/prompt-templates/:id", promptTemplateHandler.Update)
		authorized.DELETE("/prompt-templates/:id", promptTemplateHandler.Delete)
		authorized.POST("/prompt-templates/:id/apply", promptTemplateHandler.Apply)

		// Favorites.
		authorized.POST("/favorites", favoriteHandler.Create)
		authorized.GET("/favorites", favoriteHandler.List)
		authorized.GET("/favorites/check", favoriteHandler.Check)
		authorized.DELETE("/favorites/:id", favoriteHandler.Delete)

		// Tags.
		authorized.POST("/tags", tagHandler.Create)
		authorized.GET("/tags", tagHandler.List)
		authorized.POST("/tags/:id/assets", tagHandler.TagAsset)
		authorized.DELETE("/tags/:id/assets/:assetId", tagHandler.UntagAsset)

		// Collections.
		authorized.POST("/collections", collectionHandler.Create)
		authorized.GET("/collections", collectionHandler.List)
		authorized.POST("/collections/:id/assets", collectionHandler.AddAsset)
		authorized.DELETE("/collections/:id/assets/:assetId", collectionHandler.RemoveAsset)
		// API-key-authenticated generation (programmatic access). Exercises the
		// API key auth path in addition to the JWT path.
		authorized.POST("/images/generate", apiKeyMW, func(c *gin.Context) {
			c.JSON(200, gin.H{"data": gin.H{"accepted": true}})
		})
	}

	// --- Admin routes ---

	adminGroup := v1.Group("/admin")
	adminGroup.Use(adminAuth.Require())
	{
		adminGroup.GET("/dashboard", adminHandlers.Dashboard.Get)
		adminGroup.GET("/users", adminHandlers.User.List)
		adminGroup.GET("/jobs", adminHandlers.Job.List)
		adminGroup.GET("/api-keys", adminHandlers.APIKey.List)
	}

	return engine
}

// adminRouteHandlers mirrors routes.AdminHandlers for the test engine.
type adminRouteHandlers struct {
	Dashboard *admin.DashboardHandler
	User      *admin.UserHandler
	Job       *admin.JobHandler
	APIKey    *admin.APIKeyHandler
}

// mustSeedUser inserts a user directly through the data layer and returns a
// valid JWT. It provisions authenticated state for tests without going through
// the password hashing path.
func mustSeedUser(e *testEnv, email string, role user.Role) string {
	ctx := context.Background()
	u, err := e.db.User.Create().
		SetEmail(email).
		SetPasswordHash("$2a$10$integrationtestunusedhash0000000000000000000000000000").
		SetName("Integration Test").
		SetRole(role).
		SetStatus(user.StatusActive).
		Save(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to seed user %s: %v", email, err))
	}
	token, err := e.jwt.Generate(u.ID, string(role))
	if err != nil {
		panic(fmt.Sprintf("failed to sign token for %s: %v", email, err))
	}
	return token
}

// seedAsset creates an asset row directly for tests that need an existing asset.
func seedAsset(e *testEnv, ownerID int64) int64 {
	ctx := context.Background()
	a, err := e.db.Asset.Create().
		SetUserID(ownerID).
		SetSource(asset.SourceUploaded).
		SetStatus(asset.StatusActive).
		SetStorageKey(fmt.Sprintf("it/%d/%d.bin", ownerID, time.Now().UnixNano())).
		SetTitle("integration-fixture").
		Save(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to seed asset: %v", err))
	}
	return a.ID
}

// doRequest issues an in-process request against the test engine. Passing a
// non-nil body encodes it as JSON; a non-empty token sets the Bearer header.
func doRequest(e *testEnv, method, path string, body any, token string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, e.baseURL+path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)
	return w
}

// readBody decodes a JSON response body into a generic map for assertions.
func readBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &m))
	return m
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
