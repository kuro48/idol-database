package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ErrorResponse は標準エラーレスポンス形式
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorContext はエラーレスポンスの補足情報
type ErrorContext struct {
	Resource string
	Message  string
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

// NewConflictError は競合エラーを作成
func NewConflictError(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "CONFLICT",
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

// NewTooManyRequestsError はレート制限エラーを作成
func NewTooManyRequestsError(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "TOO_MANY_REQUESTS",
		Message: message,
	}
}

// WriteError はエラー内容に応じて標準レスポンスを返す
func WriteError(c *gin.Context, err error, ctx ErrorContext) {
	status, response := classifyError(err, ctx)
	c.JSON(status, response)
}

func classifyError(err error, ctx ErrorContext) (int, ErrorResponse) {
	if err == nil {
		msg := ctx.Message
		if msg == "" {
			msg = "内部サーバーエラーが発生しました"
		}
		return http.StatusInternalServerError, NewInternalError(msg)
	}

	switch {
	case isNotFoundError(err):
		resource := ctx.Resource
		if resource == "" {
			resource = "対象"
		}
		return http.StatusNotFound, NewNotFoundError(resource)
	case isConflictError(err):
		return http.StatusConflict, NewConflictError(err.Error())
	case isBadRequestError(err):
		return http.StatusBadRequest, NewBadRequestError(err.Error())
	default:
		msg := ctx.Message
		if msg == "" {
			msg = "内部サーバーエラーが発生しました"
		}
		return http.StatusInternalServerError, NewInternalError(msg)
	}
}

func isNotFoundError(err error) bool {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "見つかりません") || strings.Contains(strings.ToLower(msg), "not found")
}

func isConflictError(err error) bool {
	if mongo.IsDuplicateKeyError(err) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "既に") || strings.Contains(msg, "重複") || strings.Contains(strings.ToLower(msg), "duplicate")
}

func isBadRequestError(err error) bool {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "IDの生成エラー"):
		return true
	case strings.Contains(msg, "名前の生成エラー"):
		return true
	case strings.Contains(msg, "国の生成エラー"):
		return true
	case strings.Contains(msg, "無効"):
		return true
	case strings.Contains(msg, "不正"):
		return true
	case strings.Contains(msg, "必須"):
		return true
	case strings.Contains(msg, "形式"):
		return true
	case strings.Contains(msg, "入力"):
		return true
	default:
		return false
	}
}
