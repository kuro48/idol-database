package apikey

import "context"

// Repository はAPIキーリポジトリのインターフェース
type Repository interface {
	// Save は新しいAPIキーを保存する
	Save(ctx context.Context, key *APIKey) error

	// FindByPrefix はプレフィックスでAPIキーを取得する（複数件の可能性あり）
	FindByPrefix(ctx context.Context, prefix string) ([]*APIKey, error)

	// FindByID はIDでAPIキーを取得する
	FindByID(ctx context.Context, id string) (*APIKey, error)

	// FindByEmail はメールアドレスで全APIキーを取得する
	FindByEmail(ctx context.Context, email string) ([]*APIKey, error)

	// Update はAPIキーを更新する（isActive の変更等）
	Update(ctx context.Context, key *APIKey) error
}
