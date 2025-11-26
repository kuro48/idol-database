package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger は構造化ログミドルウェア
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// リクエスト処理
		c.Next()

		// レスポンス情報
		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 構造化ログ出力
		log.Printf(
			"[API] method=%s path=%s query=%s status=%d latency=%v ip=%s",
			method,
			path,
			query,
			statusCode,
			latency,
			clientIP,
		)

		// エラーログ
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[ERROR] %v", e.Error())
			}
		}
	}
}
