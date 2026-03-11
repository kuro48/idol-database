package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth は管理系APIの認証ミドルウェア
// Authorization: Bearer <ADMIN_API_KEY> ヘッダーを検証する
func AdminAuth(adminAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if adminAPIKey == "" {
			// APIキーが未設定の場合はサービス側の設定ミスとして503を返す
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Code:    "SERVICE_UNAVAILABLE",
				Message: "管理APIキーが設定されていません。ADMIN_API_KEY 環境変数を設定してください",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "Authorization ヘッダーの形式は 'Bearer <token>' である必要があります",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, prefix)
		if token != adminAPIKey {
			c.JSON(http.StatusForbidden, NewForbiddenError())
			c.Abort()
			return
		}

		c.Next()
	}
}
