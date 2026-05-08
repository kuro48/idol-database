package auth

import "context"

// TokenVerifier は Bearer トークンを検証して Principal を返すインターフェース
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*Principal, error)
}
