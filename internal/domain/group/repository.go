package group

import "context"

type Repository interface {
	// Save は新しいグループを保存する
	Save(ctx context.Context, group *Group) error

	// FindByID はIDでグループを検索する
	FindByID(ctx context.Context, id GroupID) (*Group, error)

	// FindAll は全てのグループを取得する
	FindAll(ctx context.Context) ([]*Group, error)

	// Update は既存のグループを更新する
	Update(ctx context.Context, group *Group) error

	// Delete はグループを削除する
	Delete(ctx context.Context, id GroupID) error

	// ExistsByName は同じ名前のグループが存在するかチェック
	ExistsByName(ctx context.Context, name GroupName) (bool, error)
}