package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

// OIDCWriteAuth は idol-auth トークンを検証し admin ロールを持つ場合のみ通過させる。
// verifier が nil（IDOL_AUTH_URL 未設定）の場合は 503 を返す。
func OIDCWriteAuth(verifier domainAuth.TokenVerifier) gin.HandlerFunc {
	return oidcAuth(verifier, (*domainAuth.Principal).CanWrite)
}

// OIDCAdminAuth は idol-auth トークンを検証し admin ロールを持つ場合のみ通過させる。
// verifier が nil（IDOL_AUTH_URL 未設定）の場合は 503 を返す。
func OIDCAdminAuth(verifier domainAuth.TokenVerifier) gin.HandlerFunc {
	return oidcAuth(verifier, (*domainAuth.Principal).CanAdmin)
}

func oidcAuth(verifier domainAuth.TokenVerifier, allowed func(*domainAuth.Principal) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verifier == nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Code:    "SERVICE_UNAVAILABLE",
				Message: "認証サービスが設定されていません。IDOL_AUTH_URL 環境変数を設定してください",
			})
			c.Abort()
			return
		}

		token, ok := extractBearer(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		principal, err := verifier.Verify(c.Request.Context(), token)
		if err != nil {
			slog.Debug("トークン検証失敗", "error", err)
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		if !allowed(principal) {
			c.JSON(http.StatusForbidden, NewForbiddenError())
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(domainAuth.WithPrincipal(c.Request.Context(), principal))
		c.Next()
	}
}

// extractBearer は Authorization: Bearer <token> からトークン文字列を取り出す
func extractBearer(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	return token, token != ""
}
