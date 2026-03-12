package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAdminAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "有効なadminキー",
			authHeader: "Bearer admin-key",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Authorizationヘッダーなし",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "無効なAPIキー",
			authHeader: "Bearer wrong-key",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Bearerプレフィックスなし",
			authHeader: "admin-key",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.AdminAuth("admin-key"))
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

func TestAdminAuth_EmptyAdminKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.AdminAuth(""))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer any-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
