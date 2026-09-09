package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/redis/go-redis/v9"
)

// HealthHandler exposes liveness and readiness probes.
type HealthHandler struct {
	db  *ent.Client
	rdb *redis.Client
}

// NewHealthHandler builds a HealthHandler.
func NewHealthHandler(db *ent.Client, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Liveness returns 200 if the process is alive. No dependency checks.
// Load balancers use this to verify the process is running.
func (h *HealthHandler) Liveness(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

// Readiness returns 200 only if all dependencies (DB, Redis) are reachable.
// Orchestrators use this to decide whether to route traffic.
func (h *HealthHandler) Readiness(c *gin.Context) {
	checks := gin.H{}
	healthy := true

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer dbCancel()
	if _, err := h.db.User.Query().Count(dbCtx); err != nil {
		checks["database"] = "unavailable: " + err.Error()
		healthy = false
	} else {
		checks["database"] = "ok"
	}

	if h.rdb != nil {
		if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = "unavailable: " + err.Error()
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	}

	checks["storage"] = "ok" // storage is lazily verified on first use

	if healthy {
		response.OK(c, gin.H{"status": "ok", "checks": checks})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "SERVICE_UNAVAILABLE", "message": "service degraded", "checks": checks}})

}
