package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders はセキュリティヘッダーを設定するミドルウェア
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HTTPS強制（1年間、サブドメイン含む）
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

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
