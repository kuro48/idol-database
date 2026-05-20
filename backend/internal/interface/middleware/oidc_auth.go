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

// OIDCUserAuth は idol-auth access token と ID token を検証し、本人情報を context に入れる。
func OIDCUserAuth(verifier domainAuth.TokenVerifier, identityVerifier domainAuth.IdentityVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verifier == nil || identityVerifier == nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Code:    "SERVICE_UNAVAILABLE",
				Message: "認証サービスが設定されていません。IDOL_AUTH_URL と IDOL_AUTH_ISSUER_URL を設定してください",
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

		idToken := c.GetHeader("X-ID-Token")
		if idToken == "" {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}
		identity, err := identityVerifier.Verify(c.Request.Context(), idToken)
		if err != nil {
			slog.Debug("IDトークン検証失敗", "error", err)
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}
		if identity.SubjectID == "" || identity.SubjectID != principal.SubjectID {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}
		if identity.Email == "" {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		principal.Email = identity.Email
		principal.DisplayName = identity.DisplayName
		principal.OshiColor = identity.OshiColor
		if len(identity.Roles) > 0 {
			principal.Roles = identity.Roles
		}
		c.Request = c.Request.WithContext(domainAuth.WithPrincipal(c.Request.Context(), principal))
		c.Next()
	}
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
