package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rl := middleware.NewRateLimiter(10, 5)
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_ExceedLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// バースト1、レートを低く設定してすぐにレート制限が発動するようにする
	rl := middleware.NewRateLimiter(0.001, 1)
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// 最初のリクエストは成功（バーストで1回許可）
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2回目のリクエストはレート制限に引っかかる
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// バースト1でレートを低く設定
	rl := middleware.NewRateLimiter(0.001, 1)
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// IP1のリクエスト
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "10.0.0.1:11111"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// IP2のリクエスト（別IPなので許可される）
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.2:22222"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
