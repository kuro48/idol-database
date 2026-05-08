package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

// OIDCVerifier は Ory Hydra が発行した JWT アクセストークンを検証する
type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
}

// NewOIDCVerifier は OIDC プロバイダーディスカバリーを実行して OIDCVerifier を作成する。
// issuerURL は Hydra の公開 URL（例: https://auth.example.com）。
// audience は idol-api のリソースサーバー識別子（例: https://api.idol.example.com）。
func NewOIDCVerifier(ctx context.Context, issuerURL, audience string) (*OIDCVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC プロバイダー初期化エラー (issuer=%s): %w", issuerURL, err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: audience,
	})

	return &OIDCVerifier{verifier: verifier}, nil
}

// Verify は JWT アクセストークンを検証して Principal を返す
func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (*domainAuth.Principal, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("トークン検証エラー: %w", err)
	}

	var claims struct {
		Email       string   `json:"email"`
		DisplayName string   `json:"display_name"`
		Roles       []string `json:"roles"`
		Scope       string   `json:"scope"` // Hydra は space-delimited 文字列で返す
	}
	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("クレーム解析エラー: %w", err)
	}

	return &domainAuth.Principal{
		SubjectID:   token.Subject,
		Email:       claims.Email,
		DisplayName: claims.DisplayName,
		Roles:       claims.Roles,
		Scopes:      strings.Fields(claims.Scope),
	}, nil
}
