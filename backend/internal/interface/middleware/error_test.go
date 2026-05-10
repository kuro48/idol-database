package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
)

func TestWriteError_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		middleware.WriteError(c, errors.New("アイドルが見つかりません"), middleware.ErrorContext{Resource: "アイドル"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWriteError_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		errMsg   string
		wantCode int
	}{
		{
			name:     "無効なID形式エラー",
			errMsg:   "無効なID形式です",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "不正な入力値エラー",
			errMsg:   "不正な入力値が含まれています",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "必須フィールドエラー",
			errMsg:   "必須フィールドがありません",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "入力値エラー",
			errMsg:   "入力値が正しくありません",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "形式エラー",
			errMsg:   "形式が正しくありません",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				middleware.WriteError(c, errors.New(tt.errMsg), middleware.ErrorContext{Resource: "リソース"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestWriteError_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		middleware.WriteError(c, errors.New("データベース接続エラー"), middleware.ErrorContext{
			Resource: "リソース",
			Message:  "内部エラーが発生しました",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteError_NilError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		middleware.WriteError(c, nil, middleware.ErrorContext{
			Message: "エラーメッセージ",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestErrorHandler_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		panic("テストパニック")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewNotFoundError(t *testing.T) {
	resp := middleware.NewNotFoundError("アイドル")
	assert.Equal(t, "NOT_FOUND", resp.Code)
	assert.Contains(t, resp.Message, "アイドル")
}

func TestNewBadRequestError(t *testing.T) {
	resp := middleware.NewBadRequestError("不正なリクエスト")
	assert.Equal(t, "BAD_REQUEST", resp.Code)
	assert.Equal(t, "不正なリクエスト", resp.Message)
}

func TestNewUnauthorizedError(t *testing.T) {
	resp := middleware.NewUnauthorizedError()
	assert.Equal(t, "UNAUTHORIZED", resp.Code)
}

func TestNewForbiddenError(t *testing.T) {
	resp := middleware.NewForbiddenError()
	assert.Equal(t, "FORBIDDEN", resp.Code)
}

func TestNewConflictError(t *testing.T) {
	resp := middleware.NewConflictError("重複しています")
	assert.Equal(t, "CONFLICT", resp.Code)
}

func TestNewTooManyRequestsError(t *testing.T) {
	resp := middleware.NewTooManyRequestsError("レート制限超過")
	assert.Equal(t, "TOO_MANY_REQUESTS", resp.Code)
}
