package auth

import (
	"context"
	"strings"
)

type principalContextKey struct{}

// Principal は認証済みユーザーの認証情報を保持する値オブジェクト。
// idol-auth はアクセストークンに roles のみ注入する。
type Principal struct {
	SubjectID   string   // Kratos identity ID (sub claim)
	Email       string   // email claim（ID token 由来）
	DisplayName string   // display_name claim（ID token 由来）
	OshiColor   string   // oshi_color claim（ID token 由来）
	Roles       []string // roles claim（idol-auth が注入）
	Scopes      []string // scope claim（ログ・デバッグ用）
}

// HasRole は指定したロールを保持しているか返す（大文字小文字を無視）
func (p *Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

// CanWrite はデータ書き込み操作（POST/PUT/DELETE）を許可するか返す
func (p *Principal) CanWrite() bool {
	return p.HasRole("admin")
}

// CanAdmin は管理操作を許可するか返す
func (p *Principal) CanAdmin() bool {
	return p.HasRole("admin")
}

// WithPrincipal は Principal をコンテキストに埋め込む
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, p)
}

// PrincipalFromContext はコンテキストから Principal を取り出す
func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalContextKey{}).(*Principal)
	return p, ok && p != nil
}
