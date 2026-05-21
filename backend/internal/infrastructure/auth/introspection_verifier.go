package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
)

const introspectPath = "/v1/public/api/token/introspect"

// IntrospectionVerifier は idol-auth のイントロスペクションエンドポイント経由で
// アクセストークンを検証する。Hydra URL の設定不要で idol-auth URL のみ必要。
type IntrospectionVerifier struct {
	introspectURL string
	httpClient    *http.Client
}

// NewIntrospectionVerifier は IntrospectionVerifier を作成する。
// idolAuthURL は idol-auth の公開 URL（例: https://auth.example.com）。
func NewIntrospectionVerifier(idolAuthURL string) (*IntrospectionVerifier, error) {
	if idolAuthURL == "" {
		return nil, fmt.Errorf("IDOL_AUTH_URL が未設定です")
	}
	return &IntrospectionVerifier{
		introspectURL: strings.TrimRight(idolAuthURL, "/") + introspectPath,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
	}, nil
}

type introspectResponse struct {
	Active bool   `json:"active"`
	Sub    string `json:"sub"`
	Scope  string `json:"scope"`
	// Hydra はアクセストークンの拡張クレームを ext に格納する
	Ext struct {
		Roles []string `json:"roles"`
	} `json:"ext"`
	// Hydra バージョンによっては top-level に出る場合もあるフォールバック
	Roles []string `json:"roles"`
}

// Verify はイントロスペクションエンドポイントを呼び出してトークンを検証し Principal を返す
func (v *IntrospectionVerifier) Verify(ctx context.Context, rawToken string) (*domainAuth.Principal, error) {
	body := url.Values{"token": {rawToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.introspectURL,
		strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("イントロスペクションリクエスト作成エラー: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("イントロスペクションリクエスト失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("イントロスペクションエンドポイントエラー: status=%d", resp.StatusCode)
	}

	var result introspectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("イントロスペクションレスポンス解析エラー: %w", err)
	}

	if !result.Active {
		return nil, fmt.Errorf("トークンが無効または期限切れです")
	}
	if result.Sub == "" {
		return nil, fmt.Errorf("イントロスペクションレスポンスに sub クレームがありません")
	}

	roles := result.Ext.Roles
	if len(roles) == 0 {
		roles = result.Roles
	}

	return &domainAuth.Principal{
		SubjectID: result.Sub,
		Roles:     roles,
		Scopes:    strings.Fields(result.Scope),
	}, nil
}
