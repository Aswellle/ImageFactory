//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

import (
	"github.com/imageforge/imageforge/ent/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	rec := doRequest(env, http.MethodGet, "/v1/health", nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)

	body := readBody(t, rec)
	assert.Equal(t, "ok", body["status"])
}

func TestUserRegistration(t *testing.T) {
	email := fmt.Sprintf("reg-%d@imageforge.test", time.Now().UnixNano())
	rec := doRequest(env, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":    email,
		"password": "supersecret123",
		"name":     "Reg User",
	}, "")
	assert.Equal(t, http.StatusCreated, rec.Code)

	body := readBody(t, rec)
	data, ok := body["data"].(map[string]any)
	require.True(t, ok, "expected data envelope, got: %v", body)

	token, _ := data["token"].(string)
	assert.NotEmpty(t, token, "registration must return an access token")

	usr, ok := data["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, email, usr["email"])

	// The returned token must be parseable and carry the user's ID.
	claims, err := env.jwt.Parse(token)
	require.NoError(t, err)
	assert.Greater(t, claims.UserID, int64(0))
}

func TestUserLogin(t *testing.T) {
	email := fmt.Sprintf("login-%d@imageforge.test", time.Now().UnixNano())
	// Register first so the account exists.
	reg := doRequest(env, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":    email,
		"password": "supersecret123",
	}, "")
	require.Equal(t, http.StatusCreated, reg.Code)

	rec := doRequest(env, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    email,
		"password": "supersecret123",
	}, "")
	assert.Equal(t, http.StatusOK, rec.Code)

	body := readBody(t, rec)
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	token, _ := data["token"].(string)
	assert.NotEmpty(t, token, "login must return an access token")

	_, err := env.jwt.Parse(token)
	require.NoError(t, err)
}

func TestCreateProject(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("proj-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)

	rec := doRequest(env, http.MethodPost, "/v1/projects", map[string]any{
		"name":        "My Project",
		"description": "created by integration test",
	}, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	body := readBody(t, rec)
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "My Project", data["name"])
}

func TestListAssets(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("assets-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)

	rec := doRequest(env, http.MethodGet, "/v1/assets", nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := readBody(t, rec)
	// Paginated list returns {data: [...], pagination: {...}}.
	_, hasData := body["data"]
	_, hasPagination := body["pagination"]
	require.True(t, hasData || hasPagination, "expected list envelope, got: %v", body)
}

func TestAPIKeyCRUD(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("apikey-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)

	createRec := doRequest(env, http.MethodPost, "/v1/api-keys", map[string]any{
		"name": "ci-key",
	}, token)
	assert.Equal(t, http.StatusCreated, createRec.Code)

	createBody := readBody(t, createRec)
	data, ok := createBody["data"].(map[string]any)
	require.True(t, ok)
	keyStr, _ := data["key"].(string)
	prefix, _ := data["prefix"].(string)
	assert.NotEmpty(t, keyStr, "created API key material must be returned once")
	assert.NotEmpty(t, prefix)
	keyID := int64(data["id"].(float64))

	// List.
	listRec := doRequest(env, http.MethodGet, "/v1/api-keys", nil, token)
	assert.Equal(t, http.StatusOK, listRec.Code)

	// Revoke.
	revokeRec := doRequest(env, http.MethodDelete, fmt.Sprintf("/v1/api-keys/%d", keyID), nil, token)
	assert.Equal(t, http.StatusNoContent, revokeRec.Code)
}

func TestFavoriteToggle(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("fav-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)
	claims, err := env.jwt.Parse(token)
	require.NoError(t, err)
	assetID := seedAsset(env, claims.UserID)

	// Add favorite.
	addRec := doRequest(env, http.MethodPost, fmt.Sprintf("/v1/favorites?asset_id=%d", assetID), nil, token)
	assert.Equal(t, http.StatusCreated, addRec.Code)

	// Check favorite (should be true).
	checkRec := doRequest(env, http.MethodGet, fmt.Sprintf("/v1/favorites/check?asset_id=%d", assetID), nil, token)
	assert.Equal(t, http.StatusOK, checkRec.Code)
	assert.Equal(t, true, readBody(t, checkRec)["data"])

	// Remove favorite.
	delRec := doRequest(env, http.MethodDelete, fmt.Sprintf("/v1/favorites/%d", assetID), nil, token)
	assert.Equal(t, http.StatusNoContent, delRec.Code)

	// Check favorite (should be false).
	checkRec2 := doRequest(env, http.MethodGet, fmt.Sprintf("/v1/favorites/check?asset_id=%d", assetID), nil, token)
	assert.Equal(t, http.StatusOK, checkRec2.Code)
	assert.Equal(t, false, readBody(t, checkRec2)["data"])
}

func TestTagCRUD(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("tag-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)
	claims, err := env.jwt.Parse(token)
	require.NoError(t, err)
	assetID := seedAsset(env, claims.UserID)

	createRec := doRequest(env, http.MethodPost, "/v1/tags", map[string]any{
		"name":  "integration-tag",
		"color": "#ff0000",
	}, token)
	assert.Equal(t, http.StatusCreated, createRec.Code)
	tagID := int64(readBody(t, createRec)["data"].(map[string]any)["id"].(float64))

	// Tag asset.
	tagRec := doRequest(env, http.MethodPost, fmt.Sprintf("/v1/tags/%d/assets", tagID), map[string]any{
		"asset_id": assetID,
	}, token)
	assert.Equal(t, http.StatusOK, tagRec.Code)

	// Untag asset.
	untagRec := doRequest(env, http.MethodDelete, fmt.Sprintf("/v1/tags/%d/assets/%d", tagID, assetID), nil, token)
	assert.Equal(t, http.StatusOK, untagRec.Code)
}

func TestCollectionCRUD(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("col-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)
	claims, err := env.jwt.Parse(token)
	require.NoError(t, err)
	assetID := seedAsset(env, claims.UserID)

	createRec := doRequest(env, http.MethodPost, "/v1/collections", map[string]any{
		"name":        "integration-collection",
		"description": "test",
	}, token)
	assert.Equal(t, http.StatusCreated, createRec.Code)
	colID := int64(readBody(t, createRec)["data"].(map[string]any)["id"].(float64))

	// Add asset.
	addRec := doRequest(env, http.MethodPost, fmt.Sprintf("/v1/collections/%d/assets", colID), map[string]any{
		"asset_id": assetID,
	}, token)
	assert.Equal(t, http.StatusOK, addRec.Code)

	// Remove asset.
	remRec := doRequest(env, http.MethodDelete, fmt.Sprintf("/v1/collections/%d/assets/%d", colID, assetID), nil, token)
	assert.Equal(t, http.StatusOK, remRec.Code)
}

func TestPromptTemplateCRUD(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("tmpl-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)

	createRec := doRequest(env, http.MethodPost, "/v1/prompt-templates", map[string]any{
		"name":      "integration-template",
		"content":   "A {{subject}} in the style of {{artist}}",
		"variables": []string{"subject", "artist"},
		"category":  "style",
	}, token)
	assert.Equal(t, http.StatusCreated, createRec.Code)
	tmplID := int64(readBody(t, createRec)["data"].(map[string]any)["id"].(float64))

	// Apply (fill variables).
	applyRec := doRequest(env, http.MethodPost, fmt.Sprintf("/v1/prompt-templates/%d/apply", tmplID), map[string]any{
		"variables": map[string]string{"subject": "cat", "artist": "picasso"},
	}, token)
	assert.Equal(t, http.StatusOK, applyRec.Code)
	assert.Equal(t, "A cat in the style of picasso", readBody(t, applyRec)["data"])

	// List.
	listRec := doRequest(env, http.MethodGet, "/v1/prompt-templates", nil, token)
	assert.Equal(t, http.StatusOK, listRec.Code)
}

func TestUsageEndpoint(t *testing.T) {
	token := mustSeedUser(env, fmt.Sprintf("usage-%d@imageforge.test", time.Now().UnixNano()), user.RoleUser)

	rec := doRequest(env, http.MethodGet, "/v1/usage", nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)
	_, ok := readBody(t, rec)["data"]
	assert.True(t, ok, "usage must return a data envelope")
}

func TestAdminEndpoints(t *testing.T) {
	// JWT admin path: the admin token was seeded in TestMain.
	rec := doRequest(env, http.MethodGet, "/v1/admin/dashboard", nil, env.authToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	_, ok := readBody(t, rec)["data"]
	assert.True(t, ok, "admin dashboard must return a data envelope")

	// Admin API key auth path: x-api-key header authenticates as synthetic admin.
	keyReq := httptest.NewRequest(http.MethodGet, env.baseURL+"/v1/admin/dashboard", nil)
	keyReq.Header.Set("x-api-key", env.adminKey)
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, keyReq)
	assert.Equal(t, http.StatusOK, w.Code)
}
