package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	domainAnalytics "github.com/kuro48/idol-api/internal/domain/analytics"
)

// usageRecorder はAPI利用記録のサービス契約
type usageRecorder interface {
	RecordUsage(ctx context.Context, record *domainAnalytics.APIUsageRecord)
}

// UsageTrackerMiddleware はAPI利用をトラッキングするミドルウェアを返す
func UsageTrackerMiddleware(svc usageRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ヘルスチェックエンドポイントは記録しない
		path := c.Request.URL.Path
		if path == "/health" || path == "/health/live" || path == "/health/ready" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		latencyMs := time.Since(start).Milliseconds()

		// APIキーのマスク処理
		maskedKey := maskAPIKey(c)

		record := &domainAnalytics.APIUsageRecord{
			MaskedKey:  maskedKey,
			Endpoint:   path,
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			LatencyMs:  latencyMs,
			RecordedAt: start,
		}

		// 非ブロッキングで記録
		svc.RecordUsage(c.Request.Context(), record)
	}
}

// maskAPIKey はAPIキーをマスクする（先頭4文字 + **** + 末尾4文字）
func maskAPIKey(c *gin.Context) string {
	// X-API-Key ヘッダーをまず確認
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		// Authorization: Bearer <token> から取得
		authHeader := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if strings.HasPrefix(authHeader, prefix) {
			apiKey = strings.TrimPrefix(authHeader, prefix)
		}
	}

	if apiKey == "" {
		return "anonymous"
	}

	if len(apiKey) <= 8 {
		return "key:****"
	}

	// 先頭4文字 + **** + 末尾4文字
	masked := fmt.Sprintf("%s****%s", apiKey[:4], apiKey[len(apiKey)-4:])
	return masked
}
