package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

var allowedAccessTokenAlgs = []jose.SignatureAlgorithm{
	jose.RS256, jose.RS384, jose.RS512,
	jose.ES256, jose.ES384, jose.ES512,
	jose.PS256, jose.PS384, jose.PS512,
	jose.EdDSA,
}

// OIDCVerifier は Ory Hydra が発行した JWT アクセストークンを検証する
type OIDCVerifier struct {
	issuer   string
	audience string
	keySet   *oidc.RemoteKeySet
}

// NewOIDCVerifier は OIDC プロバイダーディスカバリーを実行して OIDCVerifier を作成する。
// issuerURL は Hydra の公開 URL（例: https://auth.example.com）。
// audience は idol-api のリソースサーバー識別子（例: https://api.idol.example.com）。
func NewOIDCVerifier(ctx context.Context, issuerURL, audience string) (*OIDCVerifier, error) {
	if audience == "" {
		return nil, fmt.Errorf("OIDC audience が未設定です")
	}

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC プロバイダー初期化エラー (issuer=%s): %w", issuerURL, err)
	}

	var metadata struct {
		Issuer  string `json:"issuer"`
		JWKSURL string `json:"jwks_uri"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("OIDC プロバイダーメタデータ解析エラー: %w", err)
	}
	if metadata.Issuer == "" || metadata.JWKSURL == "" {
		return nil, fmt.Errorf("OIDC プロバイダーメタデータに issuer または jwks_uri がありません")
	}

	return &OIDCVerifier{
		issuer:   metadata.Issuer,
		audience: audience,
		keySet:   oidc.NewRemoteKeySet(ctx, metadata.JWKSURL),
	}, nil
}

// Verify は JWT アクセストークンを検証して Principal を返す
func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (*domainAuth.Principal, error) {
	if _, err := jwt.ParseSigned(rawToken, allowedAccessTokenAlgs); err != nil {
		return nil, fmt.Errorf("アクセストークン形式エラー: %w", err)
	}

	payload, err := v.keySet.VerifySignature(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("アクセストークン署名検証エラー: %w", err)
	}

	var registered jwt.Claims
	if err := json.Unmarshal(payload, &registered); err != nil {
		return nil, fmt.Errorf("アクセストークン登録済みクレーム解析エラー: %w", err)
	}
	if registered.Expiry == nil {
		return nil, fmt.Errorf("アクセストークンに exp クレームがありません")
	}
	if err := registered.ValidateWithLeeway(jwt.Expected{
		Issuer:      v.issuer,
		AnyAudience: jwt.Audience{v.audience},
		Time:        time.Now(),
	}, time.Minute); err != nil {
		return nil, fmt.Errorf("アクセストークンクレーム検証エラー: %w", err)
	}

	var claims struct {
		Email       string   `json:"email"`
		DisplayName string   `json:"display_name"`
		Roles       []string `json:"roles"`
		Scope       string   `json:"scope"` // Hydra は space-delimited 文字列で返す
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("アクセストークン認可クレーム解析エラー: %w", err)
	}

	return &domainAuth.Principal{
		SubjectID:   registered.Subject,
		Email:       claims.Email,
		DisplayName: claims.DisplayName,
		Roles:       claims.Roles,
		Scopes:      strings.Fields(claims.Scope),
	}, nil
}
