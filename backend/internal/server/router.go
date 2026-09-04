package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/handler"
	"github.com/imageforge/imageforge/internal/job"
	"github.com/imageforge/imageforge/internal/batchimage"
	"github.com/imageforge/imageforge/internal/repository"
	"github.com/imageforge/imageforge/internal/server/middleware"
	"github.com/imageforge/imageforge/internal/server/routes"
	"github.com/imageforge/imageforge/internal/service"
	"github.com/imageforge/imageforge/internal/storage"
	"github.com/imageforge/imageforge/internal/pkg/response"
	admin "github.com/imageforge/imageforge/internal/handler/admin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Router wraps a Gin engine and its runtime dependencies.
type Router struct {
	Engine    *gin.Engine
	db        *ent.Client
	rdb       *redis.Client
	processor job.Processor
}

// Close gracefully shuts down runtime dependencies (DB, Redis, job processor).
// It is safe to call more than once.
func (r *Router) Close(log *zap.Logger) {
	if r.processor != nil {
		if err := r.processor.Stop(); err != nil {
			log.Warn("job processor stop returned error", zap.Error(err))
		}
	}
	if r.rdb != nil {
		if err := r.rdb.Close(); err != nil {
			log.Warn("redis close returned error", zap.Error(err))
		}
	}
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			log.Warn("database close returned error", zap.Error(err))
		}
	}
}

// NewRouter builds the Gin engine, wires all dependencies, and registers routes.
// It returns an error if required infrastructure (DB, storage) cannot initialize.
func NewRouter(cfg *config.Config, log *zap.Logger) (*Router, error) {
	if cfg.Server.Mode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// Trust X-Forwarded-For/X-Real-IP from the reverse proxy (Caddy).
	// In production Caddy sets these from the real client / TCP peer.
	if len(cfg.Server.TrustedProxies) > 0 {
		if err := engine.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
			return nil, fmt.Errorf("set trusted proxies: %w", err)
		}
	} else {
		// Default: trust no proxies (use direct remote IP). Deployers should
		// set IF_SERVER_TRUSTED_PROXIES to the Caddy subnet in production.
		engine.SetTrustedProxies(nil)
	}


	// --- Global middleware (order matters) ---
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS(cfg.Server.AllowedOrigins))
	engine.Use(middleware.MaxRequestBodySize(1 << 20)) // 1 MB limit
	engine.Use(gin.Recovery())
	engine.Use(requestLogger(log))

	// --- Infrastructure ---
	db, err := repository.NewEntClient(cfg.Database.DSN(), cfg.Database.AutoMigrate)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	store, err := storage.New(cfg.Storage)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init storage: %w", err)
	}


	// --- Job queue + worker ---
	queue := job.NewMemoryQueue(256, log)
	processor := job.NewProcessor(queue, 4)
	if err := processor.Start(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("start job processor: %w", err)
	}

	// --- Redis (optional) ---
	var rdb *redis.Client
	if cfg.Redis.Host != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Error("redis unavailable, falling back to in-memory task store — tasks will be lost on process restart", zap.Error(err))
			rdb = nil
		}
	}

	// --- Services & handlers ---
	jwt, err := service.NewJWTService(cfg.Auth, cfg.Server.Mode)
	if err != nil {
		return nil, fmt.Errorf("init jwt service: %w", err)
	}
	password := service.NewPassword(cfg.Auth)
	users := repository.NewUserRepository(db)
	authService := service.NewAuthService(users, jwt, password)
	authHandler := handler.NewAuthHandler(authService)
	authMW := middleware.NewAuth(jwt)

	// --- Email & password reset ---
	emailSvc := service.NewEmailService(cfg.Email, log)
	resetCodes := repository.NewResetCodeStore(rdb)
	rateLimiter := repository.NewRateLimiter(rdb)
	resetSvc := service.NewPasswordResetService(users, resetCodes, emailSvc, password, rateLimiter)
	resetHandler := handler.NewPasswordResetHandler(resetSvc)

	// --- Account management (Sub2API integration) ---
	accountRepo := repository.NewAccountRepository(db)
	accountSvc := service.NewAccountService(accountRepo)
	accountResolver := service.NewAccountResolver(accountSvc)

	batchSvc := batchimage.NewPublicService(cfg.Sub2API.GeminiAPIKey)

	projectSvc := service.NewProjectService(db)
	assetSvc := service.NewAssetService(db, store)
	projectHandler := handler.NewProjectHandler(projectSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)

	usageSvc := service.NewUsageService(db)
	genService := service.NewGenerationService(db, batchSvc, store, queue, assetSvc, usageSvc, accountResolver)
	genHandler := handler.NewGenerationHandler(genService)
	imageGW := handler.NewOpenAIImagesHandler(handler.NewOpenAIImagesService())

	editSvc := service.NewImageEditService(db, batchSvc, store, queue, accountResolver)
	editHandler := handler.NewImageEditHandler(editSvc)


	apiKeySvc := service.NewAPIKeyService(db)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeySvc)
 	usageHandler := handler.NewUsageHandler(usageSvc)


	promptTemplateSvc := service.NewPromptTemplateService(db)
	promptTemplateHandler := handler.NewPromptTemplateHandler(promptTemplateSvc)
	apiKeyMW := middleware.APIKeyAuth(apiKeySvc)


	asyncImageHandler := handler.NewAsyncImageHandler(
		service.NewImageTaskService(repository.NewRedisImageTaskStore(rdb)),
	)

	// --- Admin ---
	adminSvc := service.NewAdminService(db)
	adminAuth := middleware.NewAdminAuth(authMW)
	adminHandlers := &routes.AdminHandlers{
		Dashboard: admin.NewDashboardHandler(adminSvc),
		User:      admin.NewUserHandler(adminSvc),
		Job:       admin.NewJobHandler(adminSvc),
		APIKey:    admin.NewAPIKeyHandler(adminSvc),
		Account:   admin.NewAccountHandler(accountSvc),
	}

	// --- Public routes ---
	v1 := engine.Group("/v1")
	{
		// Health endpoint. ?deep=true runs dependency checks (DB/Redis/storage).
		// Caddy uses the shallow check (liveness); monitoring uses deep (readiness).
		v1.GET("/health", func(c *gin.Context) {
			if c.Query("deep") == "true" {
				checks := gin.H{}
				healthy := true

				dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
				defer dbCancel()
				if _, err := db.User.Query().Count(dbCtx); err != nil {
					checks["database"] = "unavailable: " + err.Error()
					healthy = false
				} else {
					checks["database"] = "ok"
				}


				if rdb != nil {
					if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
						checks["redis"] = "unavailable: " + err.Error()
						healthy = false
					} else {
						checks["redis"] = "ok"
					}
				}

				checks["storage"] = "ok" // storage is lazily verified on first use
				if healthy {
					response.OK(c, gin.H{"status": "ok", "checks": checks})
				} else {
					c.JSON(503, gin.H{"data": gin.H{"status": "degraded", "checks": checks}})
				}
				return
			}
			response.OK(c, gin.H{"status": "ok"})
		})

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-reset-code", resetHandler.SendResetCode)
			auth.POST("/reset-password", resetHandler.ResetPassword)
		}


		v1.GET("/models", imageGW.Models)


		// OpenAI-compatible image endpoints (public for SDK access).
		v1.POST("/images/generations", imageGW.Generations)
	}

	// --- Protected routes (JWT required) ---
	authorized := v1.Group("")
	authorized.Use(authMW.Require())
	{
		// Image generation. NOTE: POST /images/generations is intentionally
		// duplicated here AND in the public group above. Gin uses first-match
		// routing, so the public OpenAI-compatible endpoint (line ~144) wins.
		// This JWT-protected copy exists as documentation of the intended
		// auth model; it is unreachable but kept for clarity.
		authorized.POST("/images/generations", genHandler.Create)
		authorized.POST("/images/edits", editHandler.Edit)
		authorized.GET("/images/jobs", genHandler.List)
		authorized.GET("/images/jobs/:id", genHandler.Get)


		// Projects.
		authorized.POST("/projects", projectHandler.Create)
		authorized.GET("/projects", projectHandler.List)
		authorized.GET("/projects/:id", projectHandler.Get)

		// Assets.
		authorized.GET("/assets", assetHandler.List)
		authorized.GET("/assets/:id", assetHandler.Get)
		authorized.DELETE("/assets/:id", assetHandler.Delete)

		// Asset versioning + content serving.
		authorized.GET("/assets/:id/content", assetHandler.Content)
		authorized.GET("/assets/:id/versions", assetHandler.Versions)
		authorized.GET("/assets/:id/versions/:vid", assetHandler.GetVersion)

		// API Keys.
		authorized.POST("/api-keys", apiKeyHandler.Create)
		authorized.GET("/api-keys", apiKeyHandler.List)
		authorized.DELETE("/api-keys/:id", apiKeyHandler.Revoke)

		// Usage.
		authorized.GET("/usage", usageHandler.Get)
	authorized.GET("/usage/history", usageHandler.History)

	// Async image tasks (submit-then-poll).
	authorized.POST("/images/tasks", asyncImageHandler.Submit)
	authorized.GET("/images/tasks/:id", asyncImageHandler.Get)

	// Prompt templates.
	authorized.POST("/prompt-templates", promptTemplateHandler.Create)
	authorized.GET("/prompt-templates", promptTemplateHandler.List)
	authorized.GET("/prompt-templates/:id", promptTemplateHandler.Get)
	authorized.PUT("/prompt-templates/:id", promptTemplateHandler.Update)
	authorized.DELETE("/prompt-templates/:id", promptTemplateHandler.Delete)
	authorized.POST("/prompt-templates/:id/apply", promptTemplateHandler.Apply)
}

// --- API-key-authed routes (no JWT required) ---
apiKeyOnly := v1.Group("")
apiKeyOnly.Use(apiKeyMW)
{
	// API-key-authed generation endpoint for programmatic access.
	// Uses /images/generate (not /images/generations) to avoid conflict
	// with the public OpenAI-compatible endpoint.
	apiKeyOnly.POST("/images/generate", genHandler.Create)
}

// --- Admin routes (adminAuth enforced at group level) ---
routes.RegisterAdminRoutes(v1, adminHandlers, adminAuth)

return &Router{Engine: engine, db: db, rdb: rdb, processor: processor}, nil
}



// requestLogger logs method, path, status, duration, and request ID. It never
// logs request bodies or headers (which may carry secrets).
func requestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", rid(c)),
		)
	}
}

func rid(c *gin.Context) string {
	v, _ := c.Get("if.request_id")
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ServeHTTP makes Router implement http.Handler so it can be passed directly
// to http.Server.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Engine.ServeHTTP(w, req)
}

// Ensure ent.Client is referenced even though router uses it via repository.
var _ = (*ent.Client)(nil)
