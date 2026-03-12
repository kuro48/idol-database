package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
)

func TestWriteAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := middleware.APIKeyConfig{
		WriteAPIKey: "test-write-key",
		AdminAPIKey: "test-admin-key",
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "有効なwriteキー",
			authHeader: "Bearer test-write-key",
			wantStatus: http.StatusOK,
		},
		{
			name:       "有効なadminキー（adminはwriteも可能）",
			authHeader: "Bearer test-admin-key",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Authorizationヘッダーなし",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "無効なAPIキー",
			authHeader: "Bearer invalid-key",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Bearerプレフィックスなし",
			authHeader: "test-write-key",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.WriteAuth(cfg))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestWriteAuth_NoKeysConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := middleware.APIKeyConfig{
		WriteAPIKey: "",
		AdminAPIKey: "",
	}

	router := gin.New()
	router.Use(middleware.WriteAuth(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer any-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestScopeAuth_WriteScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := middleware.APIKeyConfig{
		WriteAPIKey: "write-key",
		AdminAPIKey: "admin-key",
	}

	tests := []struct {
		name       string
		authHeader string
		scope      middleware.Scope
		wantStatus int
	}{
		{
			name:       "writeキーでwriteスコープ認証成功",
			authHeader: "Bearer write-key",
			scope:      middleware.ScopeWrite,
			wantStatus: http.StatusOK,
		},
		{
			name:       "adminキーでwriteスコープ認証成功",
			authHeader: "Bearer admin-key",
			scope:      middleware.ScopeWrite,
			wantStatus: http.StatusOK,
		},
		{
			name:       "adminキーでadminスコープ認証成功",
			authHeader: "Bearer admin-key",
			scope:      middleware.ScopeAdmin,
			wantStatus: http.StatusOK,
		},
		{
			name:       "writeキーでadminスコープ認証失敗",
			authHeader: "Bearer write-key",
			scope:      middleware.ScopeAdmin,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.ScopeAuth(tt.scope, cfg))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
