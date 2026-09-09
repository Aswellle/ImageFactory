package handler

import (
	"github.com/gin-gonic/gin"
	appErrors "github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/server/middleware"
	"github.com/imageforge/imageforge/internal/service"
)

// AuthHandler handles register/login.
type AuthHandler struct {
	auth         *service.AuthService
	loginTracker *middleware.LoginAttemptTracker
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(auth *service.AuthService, opts ...AuthOption) *AuthHandler {
	h := &AuthHandler{auth: auth}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// AuthOption configures AuthHandler.
type AuthOption func(*AuthHandler)

// WithLoginTracker enables login attempt tracking.
func WithLoginTracker(tracker *middleware.LoginAttemptTracker) AuthOption {
	return func(h *AuthHandler) {
		h.loginTracker = tracker
	}
}

// registerRequest is the register body.
type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
}

// loginRequest is the login body.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles POST /v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, string(appErrors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	result, err := h.auth.Register(c.Request.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, result)
}

// Login handles POST /v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, string(appErrors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}

	// 检查账户是否被锁定
	if h.loginTracker != nil {
		locked, remaining := h.loginTracker.IsLocked(req.Email)
		if locked {
			response.Error(c, 423, string(appErrors.ErrorCodeForbidden),
				"account locked due to too many failed login attempts, try again later", requestID(c))
			return
		}
		_ = remaining
	}

	result, err := h.auth.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// 记录失败
		if h.loginTracker != nil {
			_, _ = h.loginTracker.RecordFailure(req.Email)
		}
		writeError(c, err)
		return
	}

	// 登录成功，清除失败记录
	if h.loginTracker != nil {
		h.loginTracker.RecordSuccess(req.Email)
	}
	response.OK(c, result)
}

// Me returns the current user (populated by auth middleware).
func (h *AuthHandler) Me(c *gin.Context) {
	uid, _ := c.Get(ctxKeyUserID)
	role, _ := c.Get(ctxKeyRole)
	response.OK(c, gin.H{
		"user_id": uid,
		"role":    role,
	})
}
