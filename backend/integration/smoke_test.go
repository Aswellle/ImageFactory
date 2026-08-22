//go:build e2e

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestLifecycle walks the happy path a real client takes against a
// running ImageForge instance:
//
//	Register -> Login -> Create API Key -> Health Check
//
// It reuses the harness (env) built by TestMain in main_test.go.
func TestRequestLifecycle(t *testing.T) {
	if env == nil || env.engine == nil {
		t.Skip("e2e harness not initialized; database likely unavailable")
	}

	email := fmt.Sprintf("smoke-%d@imageforge.test", time.Now().UnixNano())
	password := "supersecret123"

	// 1. Register a fresh user.
	reg := doRequest(env, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":    email,
		"password": password,
		"name":     "Smoke User",
	}, "")
	require.Equal(t, http.StatusCreated, reg.Code, "register must succeed")
	require.NotEmpty(t, readBody(t, reg)["data"].(map[string]any)["token"], "register returns a token")

	// 2. Login with the new credentials issues a fresh token.
	login := doRequest(env, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    email,
		"password": password,
	}, "")
	require.Equal(t, http.StatusOK, login.Code, "login must succeed")
	loginBody := readBody(t, login)
	token, _ := loginBody["data"].(map[string]any)["token"].(string)
	require.NotEmpty(t, token, "login returns a token")

	// 3. Use the authenticated session to create an API key.
	keys := doRequest(env, http.MethodPost, "/v1/api-keys", map[string]any{
		"name": "smoke-key",
	}, token)
	require.Equal(t, http.StatusCreated, keys.Code, "api-key create must succeed")
	keyMaterial, _ := readBody(t, keys)["data"].(map[string]any)["key"].(string)
	assert.NotEmpty(t, keyMaterial, "api key material is returned exactly once")

	// 4. Health endpoint is reachable without authentication.
	health := doRequest(env, http.MethodGet, "/v1/health", nil, "")
	assert.Equal(t, http.StatusOK, health.Code)
	assert.Equal(t, "ok", readBody(t, health)["status"])

	// 5. /v1/auth/me resolves the token's identity.
	me := doRequest(env, http.MethodGet, "/v1/auth/me", nil, token)
	assert.Equal(t, http.StatusOK, me.Code)
}

// TestRegisterDuplicateConflict verifies the auth layer rejects a duplicate
// email registration with 409, proving the error contract end-to-end.
func TestRegisterDuplicateConflict(t *testing.T) {
	if env == nil || env.engine == nil {
		t.Skip("e2e harness not initialized; database likely unavailable")
	}

	email := fmt.Sprintf("dup-%d@imageforge.test", time.Now().UnixNano())

	first := doRequest(env, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":    email,
		"password": "supersecret123",
	}, "")
	require.Equal(t, http.StatusCreated, first.Code)

	second := doRequest(env, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":    email,
		"password": "supersecret123",
	}, "")
	assert.Equal(t, http.StatusConflict, second.Code, "duplicate email must conflict")
}

// TestSeededAdminAccess confirms the admin dashboard is reachable for the
// admin identity seeded in TestMain, guarding the admin auth wiring.
func TestSeededAdminAccess(t *testing.T) {
	if env == nil || env.engine == nil {
		t.Skip("e2e harness not initialized; database likely unavailable")
	}
	require.NotEmpty(t, env.authToken, "TestMain must seed an admin token")

	rec := doRequest(env, http.MethodGet, "/v1/admin/dashboard", nil, env.authToken)
	assert.Equal(t, http.StatusOK, rec.Code)
}
