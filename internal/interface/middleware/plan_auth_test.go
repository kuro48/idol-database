package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	domainusage "github.com/kuro48/idol-api/internal/domain/usage"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
)

// --- インラインスタブ ---

type stubAPIKeyRepo struct {
	keys []*domainapikey.APIKey
	err  error
}

func (r *stubAPIKeyRepo) FindByPrefix(_ context.Context, _ string) ([]*domainapikey.APIKey, error) {
	return r.keys, r.err
}
func (r *stubAPIKeyRepo) Save(_ context.Context, _ *domainapikey.APIKey) error { return nil }
func (r *stubAPIKeyRepo) FindByID(_ context.Context, _ string) (*domainapikey.APIKey, error) {
	return nil, nil
}
func (r *stubAPIKeyRepo) FindByEmail(_ context.Context, _ string) ([]*domainapikey.APIKey, error) {
	return nil, nil
}
func (r *stubAPIKeyRepo) Update(_ context.Context, _ *domainapikey.APIKey) error { return nil }

type stubUsageRepo struct {
	usage *domainusage.MonthlyUsage
	err   error
}

func (r *stubUsageRepo) IncrementAndGet(_ context.Context, _, _ string, _ int) (*domainusage.MonthlyUsage, error) {
	return r.usage, r.err
}
func (r *stubUsageRepo) Get(_ context.Context, _, _ string, _ int) (*domainusage.MonthlyUsage, error) {
	return r.usage, r.err
}

// --- ヘルパー ---

const testRawKey = "ik_live_aabbccddeeff00112233445566778899aabbccddeeff0011"

func newTestAPIKey(t *testing.T) *domainapikey.APIKey {
	t.Helper()
	k, err := domainapikey.New("aabbccddeeff001122334455", testRawKey, "test@example.com", "test", "free")
	if err != nil {
		t.Fatalf("APIKey作成失敗: %v", err)
	}
	return k
}

func withinLimitUsage() *domainusage.MonthlyUsage {
	return domainusage.Reconstruct("ik_live_aabbccdd", "2026-04", 1, 1000, time.Now())
}

func atLimitUsage() *domainusage.MonthlyUsage {
	return domainusage.Reconstruct("ik_live_aabbccdd", "2026-04", 1000, 1000, time.Now())
}

func newRouter(mw ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	for _, m := range mw {
		router.Use(m)
	}
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	router.POST("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return router
}

// --- OptionalAuth ---

func TestOptionalAuth_NoToken_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := middleware.NewPlanAuth(&stubAPIKeyRepo{}, &stubUsageRepo{usage: withinLimitUsage()})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	newRouter(m.OptionalAuth()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuth_ValidToken_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := middleware.NewPlanAuth(
		&stubAPIKeyRepo{keys: []*domainapikey.APIKey{newTestAPIKey(t)}},
		&stubUsageRepo{usage: withinLimitUsage()},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+testRawKey)
	w := httptest.NewRecorder()
	newRouter(m.OptionalAuth()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuth_InvalidToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := middleware.NewPlanAuth(
		&stubAPIKeyRepo{keys: []*domainapikey.APIKey{}},
		&stubUsageRepo{usage: withinLimitUsage()},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ik_live_invalidkeyinvalidkeyinvalidkeyinvalidkeyinvalid")
	w := httptest.NewRecorder()
	newRouter(m.OptionalAuth()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOptionalAuth_LimitExceeded_Returns429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := middleware.NewPlanAuth(
		&stubAPIKeyRepo{keys: []*domainapikey.APIKey{newTestAPIKey(t)}},
		&stubUsageRepo{usage: atLimitUsage()},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+testRawKey)
	w := httptest.NewRecorder()
	newRouter(m.OptionalAuth()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// --- Auth ---

func TestAuth_NoToken_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := middleware.NewPlanAuth(&stubAPIKeyRepo{}, &stubUsageRepo{usage: withinLimitUsage()})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	newRouter(m.Auth()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ValidToken_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := middleware.NewPlanAuth(
		&stubAPIKeyRepo{keys: []*domainapikey.APIKey{newTestAPIKey(t)}},
		&stubUsageRepo{usage: withinLimitUsage()},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+testRawKey)
	w := httptest.NewRecorder()
	newRouter(m.Auth()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- RequireWrite ---

func TestRequireWrite_WithWriteEnabled_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setWrite := func(c *gin.Context) { c.Set(middleware.CtxKeyWriteEnabled, true); c.Next() }

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	newRouter(setWrite, middleware.RequireWrite()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireWrite_WriteDisabled_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setWrite := func(c *gin.Context) { c.Set(middleware.CtxKeyWriteEnabled, false); c.Next() }

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	newRouter(setWrite, middleware.RequireWrite()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireWrite_NoContext_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	newRouter(middleware.RequireWrite()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
