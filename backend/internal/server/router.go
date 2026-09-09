package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/batchimage"
	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/handler"
	"github.com/imageforge/imageforge/internal/job"
	"github.com/imageforge/imageforge/internal/pkg/crypto"
	"github.com/imageforge/imageforge/internal/repository"
	"github.com/imageforge/imageforge/internal/server/middleware"
	"github.com/imageforge/imageforge/internal/server/routes"
	"github.com/imageforge/imageforge/internal/service"
	"github.com/imageforge/imageforge/internal/storage"
	"github.com/imageforge/imageforge/internal/web"
	"github.com/imageforge/imageforge/migrations"
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
	engine.Use(middleware.MaxRequestBodySize(8 << 20)) // 8 MB limit (image edits w/ base64 can exceed 1MB)

	engine.Use(requestLogger(log))
	// 全局限流：按 IP 限制请求速率，防止暴力破解和 DoS。
	engine.Use(middleware.RateLimitByIP(10, 20)) // 10 req/s per IP, burst 20


	// --- Infrastructure ---
	db, err := repository.NewEntClient(cfg.Database.DSN(), cfg.Database.AutoMigrate)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	// --- Versioned SQL migrations (production-safe) ---
	// Run embedded SQL migrations before serving traffic. Idempotent: only
	// applies untracked migrations.
	if err := repository.MigrateUp(db, migrations.FS); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	store, err := storage.New(cfg.Storage)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init storage: %w", err)
	}

	// --- Credential encryption (optional) ---
	// 从环境变量 IF_CREDENTIAL_KEY 加载加密密钥。
	// 未设置时凭证以明文存储（开发模式）。
	if err := crypto.SetKey(os.Getenv("IF_CREDENTIAL_KEY")); err != nil {
		return nil, fmt.Errorf("init credential encryption: %w", err)
	}
	if crypto.IsEnabled() {
		log.Info("credential encryption enabled (AES-256-GCM)")
	} else {
		log.Warn("credential encryption DISABLED — credentials stored in plaintext")
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
			Addr:     net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port)),
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

	// --- Health handler (liveness + readiness) ---
	healthHandler := handler.NewHealthHandler(db, rdb)

	// --- Account management (Sub2API integration) ---
	accountRepo := repository.NewAccountRepository(db)
	accountSvc := service.NewAccountService(accountRepo)
	schedulingSvc := service.NewSchedulingService(accountRepo)
	schedulingSvc.SetThresholds(map[string]int{
		"openai":    80,
		"anthropic": 80,
		"grok":      80,
		"kimi":      80,
		"zhipu":     80,
	})
	accountResolver := service.NewAccountResolver(accountRepo, accountSvc, schedulingSvc)

	// --- Batch image service (with DB persistence) ---
	batchSvc := batchimage.NewPublicServiceWithDB(cfg.Sub2API.GeminiAPIKey, db)

	projectSvc := service.NewProjectService(db)
	assetSvc := service.NewAssetService(db, store)
	projectHandler := handler.NewProjectHandler(projectSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)

	usageSvc := service.NewUsageService(db)
	genService := service.NewGenerationService(service.GenerationConfig{
		DB:              db,
		Batch:           batchSvc,
		Store:           store,
		Queue:           queue,
		Assets:          assetSvc,
		Usage:           usageSvc,
		AccountResolver: accountResolver,
	})
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
		Scheduling: admin.NewSchedulingHandler(schedulingSvc),
	}


	// --- Public routes ---
	v1 := engine.Group("/v1")
	{
		// Liveness probe — always returns 200 if the process is alive.
		// Load balancers use this to verify the process is running.
		v1.GET("/healthz", healthHandler.Liveness)

		// Readiness probe — returns 200 only if DB/Redis are reachable.
		// Orchestrators use this to decide whether to route traffic.
		v1.GET("/ready", healthHandler.Readiness)

		// Legacy health endpoint (backward-compatible). ?deep=true runs dependency checks.
		v1.GET("/health", func(c *gin.Context) {
			if c.Query("deep") == "true" {
				healthHandler.Readiness(c)
				return
			}
			healthHandler.Liveness(c)
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
		authorized.POST("/images/generations", genHandler.Create)
		authorized.POST("/images/edits", editHandler.Edit)
		authorized.GET("/images/jobs", genHandler.List)
		authorized.GET("/images/jobs/:id", genHandler.Get)

		authorized.POST("/projects", projectHandler.Create)
		authorized.GET("/projects", projectHandler.List)
		authorized.GET("/projects/:id", projectHandler.Get)

		authorized.GET("/assets", assetHandler.List)
		authorized.GET("/assets/:id", assetHandler.Get)
		authorized.DELETE("/assets/:id", assetHandler.Delete)

		authorized.GET("/assets/:id/content", assetHandler.Content)
		authorized.GET("/assets/:id/versions", assetHandler.Versions)
		authorized.GET("/assets/:id/versions/:vid", assetHandler.GetVersion)

		authorized.POST("/api-keys", apiKeyHandler.Create)
		authorized.GET("/api-keys", apiKeyHandler.List)
		authorized.DELETE("/api-keys/:id", apiKeyHandler.Revoke)

		authorized.GET("/usage", usageHandler.Get)
		authorized.GET("/usage/history", usageHandler.History)

		authorized.POST("/images/tasks", asyncImageHandler.Submit)
		authorized.GET("/images/tasks/:id", asyncImageHandler.Get)

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
		apiKeyOnly.POST("/images/generate", genHandler.Create)
	}

	// --- Admin routes (adminAuth enforced at group level) ---
	routes.RegisterAdminRoutes(v1, adminHandlers, adminAuth)

	// --- Embedded SPA (must be registered last) ---
	// Serve the compiled frontend from go:embed. Static assets are served
	// directly; all other paths fall back to index.html for client-side routing.
	engine.StaticFS("/", http.FS(web.Dist))
	engine.NoRoute(func(c *gin.Context) {
		// Only fall back to index.html for non-API GET requests.
		if c.Request.Method != "GET" || strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.FileFromFS("/", http.FS(web.Dist))
	})

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
