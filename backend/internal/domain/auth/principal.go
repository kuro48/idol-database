package auth

import (
	"context"
	"strings"
)

type principalContextKey struct{}

// Principal は認証済みユーザーの認証情報を保持する値オブジェクト
type Principal struct {
	SubjectID   string   // Kratos identity ID (sub claim)
	Email       string   // email claim
	DisplayName string   // display_name claim
	Roles       []string // roles claim
	Scopes      []string // scope claim (space-delimited → slice)
}

// HasScope は指定したスコープを保持しているか返す
func (p *Principal) HasScope(scope string) bool {
	for _, s := range p.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasRole は指定したロールを保持しているか返す
func (p *Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

// CanWrite は idol.write / idol.admin スコープまたは admin ロールを持つか返す
func (p *Principal) CanWrite() bool {
	return p.HasScope("idol.write") || p.HasScope("idol.admin") || p.HasRole("admin")
}

// CanAdmin は idol.admin スコープまたは admin ロールを持つか返す
func (p *Principal) CanAdmin() bool {
	return p.HasScope("idol.admin") || p.HasRole("admin")
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
