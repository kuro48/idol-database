package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/shared/logger"
)

// generateRequestID はランダムなリクエストIDを生成する
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// Logger は構造化ログミドルウェア
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// リクエストIDを生成してコンテキストに設定
		requestID := generateRequestID()
		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)

		// リクエスト処理
		c.Next()

		// レスポンス情報
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 構造化ログ出力
		slog.InfoContext(ctx, "HTTPリクエスト",
			"request_id", requestID,
			"method", method,
			"path", path,
			"query", query,
			"status", statusCode,
			"latency_ms", latency.Milliseconds(),
			"ip", clientIP,
		)

		// エラーログ
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				slog.ErrorContext(ctx, "リクエストエラー",
					"request_id", requestID,
					"error", e.Error(),
				)
			}
		}
	}
}
