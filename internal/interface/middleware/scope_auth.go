package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Scope はAPIキーのスコープ定義
type Scope string

const (
	// ScopeWrite は書き込み操作（POST/PUT/DELETE）を許可
	ScopeWrite Scope = "write"
	// ScopeAdmin は管理操作を許可（write スコープも含む）
	ScopeAdmin Scope = "admin"
)

// APIKeyConfig はスコープ別APIキー設定
type APIKeyConfig struct {
	WriteAPIKey string // write スコープ以上に有効なキー
	AdminAPIKey string // admin スコープのみに有効なキー
}

// ScopeAuth はスコープベースのAPIキー認証ミドルウェア
// 階層構造: admin スコープのキーは write スコープでも使用可能
func ScopeAuth(required Scope, cfg APIKeyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if !isAuthorized(token, required, cfg) {
			c.JSON(http.StatusForbidden, NewForbiddenError())
			c.Abort()
			return
		}

		c.Next()
	}
}

// isAuthorized はトークンが必要スコープを持つか検証する
// スコープ階層: admin ⊇ write
// タイミング攻撃対策のため subtle.ConstantTimeCompare を使用する
func isAuthorized(token string, required Scope, cfg APIKeyConfig) bool {
	switch required {
	case ScopeAdmin:
		// admin スコープは AdminAPIKey のみ許可
		return cfg.AdminAPIKey != "" && constantTimeEqual(token, cfg.AdminAPIKey)
	case ScopeWrite:
		// write スコープは WriteAPIKey または AdminAPIKey を許可
		if cfg.WriteAPIKey != "" && constantTimeEqual(token, cfg.WriteAPIKey) {
			return true
		}
		if cfg.AdminAPIKey != "" && constantTimeEqual(token, cfg.AdminAPIKey) {
			return true
		}
		return false
	default:
		return false
	}
}

// constantTimeEqual は長さが異なる場合もタイミングが一定な文字列比較を行う
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// WriteAuth は write スコープ認証ミドルウェアを返す
func WriteAuth(cfg APIKeyConfig) gin.HandlerFunc {
	if cfg.WriteAPIKey == "" && cfg.AdminAPIKey == "" {
		// キーが未設定の場合は503
		return func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Code:    "SERVICE_UNAVAILABLE",
				Message: "APIキーが設定されていません。WRITE_API_KEY または ADMIN_API_KEY 環境変数を設定してください",
			})
			c.Abort()
		}
	}
	return ScopeAuth(ScopeWrite, cfg)
}
