package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

// OIDCWriteAuth は idol.write / idol.admin スコープを持つ OIDC トークン、
// または既存の write/admin API キーを受け付ける複合認証ミドルウェアを返す。
// verifier が nil の場合は API キー認証のみにフォールバックする（後方互換）。
func OIDCWriteAuth(verifier domainAuth.TokenVerifier, apiKeyCfg APIKeyConfig) gin.HandlerFunc {
	apikeyFallback := WriteAuth(apiKeyCfg)

	return func(c *gin.Context) {
		if verifier != nil {
			if token, ok := extractBearer(c); ok {
				principal, err := verifier.Verify(c.Request.Context(), token)
				if err == nil && principal.CanWrite() {
					c.Request = c.Request.WithContext(domainAuth.WithPrincipal(c.Request.Context(), principal))
					c.Next()
					return
				}
				if err != nil {
					slog.Debug("OIDC トークン検証失敗（APIキーへフォールバック）", "error", err)
				}
			}
		}
		apikeyFallback(c)
	}
}

// OIDCAdminAuth は idol.admin スコープまたは admin ロールを持つ OIDC トークン、
// または既存の admin API キーを受け付ける複合認証ミドルウェアを返す。
// verifier が nil の場合は API キー認証のみにフォールバックする（後方互換）。
func OIDCAdminAuth(verifier domainAuth.TokenVerifier, adminAPIKey string) gin.HandlerFunc {
	apikeyFallback := AdminAuth(adminAPIKey)

	return func(c *gin.Context) {
		if verifier != nil {
			if token, ok := extractBearer(c); ok {
				principal, err := verifier.Verify(c.Request.Context(), token)
				if err == nil && principal.CanAdmin() {
					c.Request = c.Request.WithContext(domainAuth.WithPrincipal(c.Request.Context(), principal))
					c.Next()
					return
				}
				if err != nil {
					slog.Debug("OIDC トークン検証失敗（APIキーへフォールバック）", "error", err)
				}
			}
		}
		apikeyFallback(c)
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
