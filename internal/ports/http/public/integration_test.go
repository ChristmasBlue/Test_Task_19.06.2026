package public_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"test_task/internal/adapters/storage/cache/redis"
	"test_task/internal/adapters/storage/database/ms"
	"test_task/internal/cases"
	"test_task/internal/ports/http/public"
)

func setupHTTPServer(t *testing.T) (*gin.Engine, func()) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	storage, err := ms.NewStorage(&testConfig{})
	require.NoError(t, err)

	cache, err := redis.NewRedisCache(&testConfig{})
	require.NoError(t, err)

	svc, err := cases.NewService(storage, cache, &testConfig{})
	require.NoError(t, err)

	server, err := public.NewServer(svc, &testConfig{})
	require.NoError(t, err)

	router := gin.New()
	server.RegisterRoutes()

	cleanup := func() {
		storage.Stop(context.Background())
		cache.Stop(context.Background())
	}

	return router, cleanup
}

func TestHTTPIntegration_RegisterAndLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	router, cleanup := setupHTTPServer(t)
	defer cleanup()

	reqBody := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "12345678",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// 2. Логин
	loginBody := map[string]string{
		"email":    "john@example.com",
		"password": "12345678",
	}
	loginJSON, _ := json.Marshal(loginBody)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(loginJSON))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code)

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	require.Contains(t, resp, "token")
	require.NotEmpty(t, resp["token"])
}

func TestHTTPIntegration_CreateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	router, cleanup := setupHTTPServer(t)
	defer cleanup()

	reqBody := map[string]string{
		"name":     "Jane Doe",
		"email":    "jane@example.com",
		"password": "12345678",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody := map[string]string{
		"email":    "jane@example.com",
		"password": "12345678",
	}
	loginJSON, _ := json.Marshal(loginBody)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(loginJSON))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	var loginResp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	teamBody := map[string]string{
		"name": "Team A",
	}
	teamJSON, _ := json.Marshal(teamBody)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamJSON))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusCreated, w3.Code)
}
