package auth

import "context"

// IdentityVerifier は ID token を検証して本人プロフィール claim を返す。
type IdentityVerifier interface {
	Verify(ctx context.Context, rawToken string) (*Principal, error)
}
