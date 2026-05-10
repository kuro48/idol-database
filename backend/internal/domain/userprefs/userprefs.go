package userprefs

import (
	"errors"
	"regexp"
	"time"
)

var hexColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// UserPrefs は OIDC ユーザーのアプリケーション設定を保持する
type UserPrefs struct {
	sub       string // OIDC subject ID（Kratos identity ID）
	oshiColor string // CSS hex color（例: "#FF69B4"）、空文字はデフォルト
	updatedAt time.Time
}

// New は新しい UserPrefs を作成する
func New(sub string) (*UserPrefs, error) {
	if sub == "" {
		return nil, errors.New("sub は必須です")
	}
	return &UserPrefs{sub: sub, updatedAt: time.Now()}, nil
}

// Reconstruct は DB から取得したデータで UserPrefs を再構築する
func Reconstruct(sub, oshiColor string, updatedAt time.Time) (*UserPrefs, error) {
	if sub == "" {
		return nil, errors.New("sub は必須です")
	}
	return &UserPrefs{sub: sub, oshiColor: oshiColor, updatedAt: updatedAt}, nil
}

// UpdateOshiColor は推しメンカラーを更新する（空文字はクリア）
func (p *UserPrefs) UpdateOshiColor(color string) error {
	if color != "" && !hexColorPattern.MatchString(color) {
		return errors.New("推しメンカラーは #RGB または #RRGGBB 形式で指定してください")
	}
	p.oshiColor = color
	p.updatedAt = time.Now()
	return nil
}

func (p *UserPrefs) Sub() string          { return p.sub }
func (p *UserPrefs) OshiColor() string    { return p.oshiColor }
func (p *UserPrefs) UpdatedAt() time.Time { return p.updatedAt }
