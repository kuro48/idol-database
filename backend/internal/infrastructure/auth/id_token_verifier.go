package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

type IDTokenVerifier struct {
	issuerURL  string
	clientID   string
	httpClient *http.Client
	mu         sync.RWMutex
	jwksURL    string
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
}

func NewIDTokenVerifier(issuerURL, clientID string) (*IDTokenVerifier, error) {
	if strings.TrimSpace(issuerURL) == "" {
		return nil, fmt.Errorf("IDOL_AUTH_ISSUER_URL が未設定です")
	}
	return &IDTokenVerifier{
		issuerURL:  strings.TrimRight(issuerURL, "/"),
		clientID:   strings.TrimSpace(clientID),
		httpClient: &http.Client{Timeout: 5 * time.Second},
		keys:       map[string]*rsa.PublicKey{},
	}, nil
}

type oidcDiscovery struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type idTokenClaims struct {
	Issuer      string          `json:"iss"`
	Subject     string          `json:"sub"`
	Audience    json.RawMessage `json:"aud"`
	ExpiresAt   int64           `json:"exp"`
	NotBefore   int64           `json:"nbf"`
	Email       string          `json:"email"`
	DisplayName string          `json:"display_name"`
	OshiColor   string          `json:"oshi_color"`
	Roles       []string        `json:"roles"`
}

func (v *IDTokenVerifier) Verify(ctx context.Context, rawToken string) (*domainAuth.Principal, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("IDトークン形式が不正です")
	}

	var header jwtHeader
	if err := decodeJWTPart(parts[0], &header); err != nil {
		return nil, fmt.Errorf("IDトークンヘッダー解析エラー: %w", err)
	}
	if header.Alg != "RS256" || header.Kid == "" {
		return nil, errors.New("未対応のIDトークン署名方式です")
	}

	key, err := v.key(ctx, header.Kid)
	if err != nil {
		return nil, err
	}
	signed := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(signed))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("IDトークン署名デコードエラー: %w", err)
	}
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return nil, fmt.Errorf("IDトークン署名検証エラー: %w", err)
	}

	var claims idTokenClaims
	if err := decodeJWTPart(parts[1], &claims); err != nil {
		return nil, fmt.Errorf("IDトークンclaim解析エラー: %w", err)
	}
	now := time.Now().Unix()
	if claims.Issuer != v.issuerURL {
		return nil, errors.New("IDトークンissuerが一致しません")
	}
	if claims.Subject == "" {
		return nil, errors.New("IDトークンsubがありません")
	}
	if claims.ExpiresAt <= now {
		return nil, errors.New("IDトークンが期限切れです")
	}
	if claims.NotBefore != 0 && claims.NotBefore > now {
		return nil, errors.New("IDトークンはまだ有効ではありません")
	}
	if v.clientID != "" && !claims.hasAudience(v.clientID) {
		return nil, errors.New("IDトークンaudienceが一致しません")
	}

	return &domainAuth.Principal{
		SubjectID:   claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.DisplayName,
		OshiColor:   claims.OshiColor,
		Roles:       claims.Roles,
	}, nil
}

func decodeJWTPart(part string, dest any) error {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c idTokenClaims) hasAudience(clientID string) bool {
	var single string
	if err := json.Unmarshal(c.Audience, &single); err == nil {
		return single == clientID
	}
	var many []string
	if err := json.Unmarshal(c.Audience, &many); err != nil {
		return false
	}
	for _, aud := range many {
		if aud == clientID {
			return true
		}
	}
	return false
}

func (v *IDTokenVerifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key := v.keys[kid]
	fresh := time.Since(v.fetchedAt) < 15*time.Minute
	v.mu.RUnlock()
	if key != nil && fresh {
		return key, nil
	}

	if err := v.refresh(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	key = v.keys[kid]
	if key == nil {
		return nil, errors.New("IDトークン署名キーが見つかりません")
	}
	return key, nil
}

func (v *IDTokenVerifier) refresh(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.jwksURL == "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.issuerURL+"/.well-known/openid-configuration", nil)
		if err != nil {
			return err
		}
		resp, err := v.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("OIDC discovery取得失敗: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("OIDC discovery取得失敗: status=%d", resp.StatusCode)
		}
		var discovery oidcDiscovery
		if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
			return fmt.Errorf("OIDC discovery解析失敗: %w", err)
		}
		if discovery.Issuer != "" && strings.TrimRight(discovery.Issuer, "/") != v.issuerURL {
			return errors.New("OIDC discovery issuerが設定値と一致しません")
		}
		if discovery.JWKSURI == "" {
			return errors.New("OIDC discoveryにjwks_uriがありません")
		}
		v.jwksURL = discovery.JWKSURI
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("JWKS取得失敗: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS取得失敗: status=%d", resp.StatusCode)
	}
	var jwks jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("JWKS解析失敗: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" || jwk.Kid == "" || jwk.N == "" || jwk.E == "" {
			continue
		}
		key, err := rsaPublicKey(jwk.N, jwk.E)
		if err != nil {
			continue
		}
		keys[jwk.Kid] = key
	}
	v.keys = keys
	v.fetchedAt = time.Now()
	return nil
}

func rsaPublicKey(n64, e64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(n64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(e64)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("JWK exponentが不正です")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}
