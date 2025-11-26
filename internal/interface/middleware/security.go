package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders はセキュリティヘッダーを設定するミドルウェア
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS対策
		c.Header("X-XSS-Protection", "1; mode=block")

		// クリックジャッキング対策
		c.Header("X-Frame-Options", "DENY")

		// MIMEタイプスニッフィング対策
		c.Header("X-Content-Type-Options", "nosniff")

		// コンテンツセキュリティポリシー
		c.Header("Content-Security-Policy", "default-src 'self'")

		// リファラーポリシー
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy（旧Feature-Policy）
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
