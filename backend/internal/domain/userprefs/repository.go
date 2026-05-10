package userprefs

import "context"

// Repository はユーザー設定リポジトリのインターフェース
type Repository interface {
	// FindBySub は OIDC sub でユーザー設定を取得する（存在しない場合は nil, nil）
	FindBySub(ctx context.Context, sub string) (*UserPrefs, error)

	// Upsert はユーザー設定を作成または更新する
	Upsert(ctx context.Context, prefs *UserPrefs) error
}
