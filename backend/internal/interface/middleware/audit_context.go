package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/shared/audit"
)

const (
	// ContextKeyActor はAPIキー識別子（アクター）のコンテキストキー
	ContextKeyActor = "audit_actor"
	// ContextKeySource はリクエストソースのコンテキストキー
	ContextKeySource = "audit_source"

	// SourceAPI はAPIリクエスト由来
	SourceAPI = "api"
)

// AuditContext はAPIキーから監査コンテキストをginコンテキストに格納するミドルウェア
// 認証は行わず、Authorizationヘッダーが存在すればアクターを記録するだけ
func AuditContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor := extractActor(c)
		c.Set(ContextKeyActor, actor)
		c.Set(ContextKeySource, SourceAPI)
		c.Next()
	}
}

// extractActor はAuthorizationヘッダーからアクター識別子を取得する
// キーの末尾8文字をマスク付き識別子として使用
func extractActor(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "anonymous"
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "anonymous"
	}

	token := strings.TrimPrefix(authHeader, prefix)
	if len(token) == 0 {
		return "anonymous"
	}

	// セキュリティのため末尾4文字のみ使用
	if len(token) <= 4 {
		return "key:****"
	}
	return "key:***" + token[len(token)-4:]
}

// GetActor はginコンテキストからアクターを取得する
func GetActor(c *gin.Context) string {
	if actor, exists := c.Get(ContextKeyActor); exists {
		if s, ok := actor.(string); ok {
			return s
		}
	}
	return "anonymous"
}

// GetSource はginコンテキストからソースを取得する
func GetSource(c *gin.Context) string {
	if source, exists := c.Get(ContextKeySource); exists {
		if s, ok := source.(string); ok {
			return s
		}
	}
	return SourceAPI
}

// AuditContextFor は gin.Context から監査情報を取り出し、Go の context.Context に埋め込んで返す
// ハンドラー内で c.Request.Context() の代わりに使用する
func AuditContextFor(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	ctx = audit.WithActor(ctx, GetActor(c))
	ctx = audit.WithSource(ctx, GetSource(c))
	return ctx
}
