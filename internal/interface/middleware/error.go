package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse は標準エラーレスポンス形式
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler はエラーハンドリングミドルウェア
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// パニック回復
		defer func() {
			if err := recover(); err != nil {
				log.Printf("パニック回復: %v", err)

				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Code:    "INTERNAL_ERROR",
					Message: "内部サーバーエラーが発生しました",
				})
				c.Abort()
			}
		}()

		c.Next()

		// エラーが設定されている場合の処理
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// すでにレスポンスが書き込まれている場合はスキップ
			if c.Writer.Written() {
				return
			}

			statusCode := c.Writer.Status()
			if statusCode == http.StatusOK {
				statusCode = http.StatusInternalServerError
			}

			c.JSON(statusCode, ErrorResponse{
				Code:    "REQUEST_ERROR",
				Message: err.Error(),
			})
		}
	}
}

// ValidationError はバリデーションエラー
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationErrorResponse はバリデーションエラーレスポンスを作成
func NewValidationErrorResponse(errors []ValidationError) ErrorResponse {
	return ErrorResponse{
		Code:    "VALIDATION_ERROR",
		Message: "入力値が不正です",
		Details: errors,
	}
}

// NewBadRequestError は不正なリクエストエラーを作成
func NewBadRequestError(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

// NewNotFoundError はリソース不在エラーを作成
func NewNotFoundError(resource string) ErrorResponse {
	return ErrorResponse{
		Code:    "NOT_FOUND",
		Message: resource + "が見つかりません",
	}
}

// NewInternalError は内部エラーを作成
func NewInternalError(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}
}

// NewUnauthorizedError は認証エラーを作成
func NewUnauthorizedError() ErrorResponse {
	return ErrorResponse{
		Code:    "UNAUTHORIZED",
		Message: "認証が必要です",
	}
}

// NewForbiddenError は権限エラーを作成
func NewForbiddenError() ErrorResponse {
	return ErrorResponse{
		Code:    "FORBIDDEN",
		Message: "この操作を実行する権限がありません",
	}
}
