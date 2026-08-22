package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// AuthHandler handles register/login.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
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
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
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
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	result, err := h.auth.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(c, err)
		return
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
