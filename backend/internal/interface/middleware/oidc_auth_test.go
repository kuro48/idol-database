package middleware_test

// D-11: oidc_auth.go の認証分岐テスト。
// verifier が nil の場合 503、Bearer なし/検証失敗は 401、
// スコープ不足は 403 を返すことを固定する。

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
)

// stubTokenVerifier は domain/auth.TokenVerifier のテスト用実装
type stubTokenVerifier struct {
	principal *domainAuth.Principal
	err       error
}

func (s *stubTokenVerifier) Verify(_ context.Context, _ string) (*domainAuth.Principal, error) {
	return s.principal, s.err
}

// stubIdentityVerifier は domain/auth.IdentityVerifier のテスト用実装
type stubIdentityVerifier struct {
	principal *domainAuth.Principal
	err       error
}

func (s *stubIdentityVerifier) Verify(_ context.Context, _ string) (*domainAuth.Principal, error) {
	return s.principal, s.err
}

// --- OIDCWriteAuth / OIDCAdminAuth テスト ---

func TestOIDCWriteAuth_NilVerifier_Returns503(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.OIDCWriteAuth(nil))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestOIDCAdminAuth_NilVerifier_Returns503(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.OIDCAdminAuth(nil))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestOIDCWriteAuth_NoBearerToken_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{Roles: []string{"admin"}}}
	router := gin.New()
	router.Use(middleware.OIDCWriteAuth(verifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOIDCWriteAuth_InvalidToken_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{err: errors.New("invalid token")}
	router := gin.New()
	router.Use(middleware.OIDCWriteAuth(verifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOIDCWriteAuth_InsufficientRole_Returns403(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	// write は admin ロールが必要（Principal.CanWrite() = HasRole("admin")）
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{Roles: []string{"viewer"}}}
	router := gin.New()
	router.Use(middleware.OIDCWriteAuth(verifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOIDCWriteAuth_AdminRole_Passes(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{Roles: []string{"admin"}}}
	router := gin.New()
	router.Use(middleware.OIDCWriteAuth(verifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- OIDCUserAuth テスト ---

func TestOIDCUserAuth_NilVerifier_Returns503(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.OIDCUserAuth(nil, nil))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestOIDCUserAuth_NoBearerToken_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{SubjectID: "user-1"}}
	idVerifier := &stubIdentityVerifier{principal: &domainAuth.Principal{SubjectID: "user-1", Email: "user@example.com"}}
	router := gin.New()
	router.Use(middleware.OIDCUserAuth(verifier, idVerifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOIDCUserAuth_NoIDToken_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{SubjectID: "user-1"}}
	idVerifier := &stubIdentityVerifier{principal: &domainAuth.Principal{SubjectID: "user-1", Email: "user@example.com"}}
	router := gin.New()
	router.Use(middleware.OIDCUserAuth(verifier, idVerifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	// X-ID-Token ヘッダーなし
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOIDCUserAuth_SubjectMismatch_Returns401(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{SubjectID: "user-1"}}
	idVerifier := &stubIdentityVerifier{principal: &domainAuth.Principal{SubjectID: "user-DIFFERENT", Email: "user@example.com"}}
	router := gin.New()
	router.Use(middleware.OIDCUserAuth(verifier, idVerifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-ID-Token", "valid-id-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOIDCUserAuth_ValidTokens_Passes(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	verifier := &stubTokenVerifier{principal: &domainAuth.Principal{SubjectID: "user-1"}}
	idVerifier := &stubIdentityVerifier{principal: &domainAuth.Principal{SubjectID: "user-1", Email: "user@example.com"}}
	router := gin.New()
	router.Use(middleware.OIDCUserAuth(verifier, idVerifier))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-ID-Token", "valid-id-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
