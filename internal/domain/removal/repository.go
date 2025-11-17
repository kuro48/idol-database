package removal

import "context"

// Repository は削除申請リポジトリのインターフェース
type Repository interface {
	// Save は新しい削除申請を保存する
	Save(ctx context.Context, request *RemovalRequest) error

	// FindByID はIDで削除申請を取得する
	FindByID(ctx context.Context, id RemovalID) (*RemovalRequest, error)

	// FindAll は全ての削除申請を取得する
	FindAll(ctx context.Context) ([]*RemovalRequest, error)

	// FindPending は保留中の削除申請を取得する
	FindPending(ctx context.Context) ([]*RemovalRequest, error)

	// Update は削除申請を更新する
	Update(ctx context.Context, request *RemovalRequest) error

	// Delete は削除申請を削除する
	Delete(ctx context.Context, id RemovalID) error
}
