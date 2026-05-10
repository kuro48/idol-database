package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

// RequireOIDC は OIDC トークン専用の認証ミドルウェアを返す。
// verifier が nil（OIDC 未設定）の場合は 503 を返す。
func RequireOIDC(verifier domainAuth.TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verifier == nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Code:    "OIDC_NOT_CONFIGURED",
				Message: "OIDC 認証が設定されていません",
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
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(domainAuth.WithPrincipal(c.Request.Context(), principal))
		c.Next()
	}
}
