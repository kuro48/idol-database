package idol

import "context"

// Repository はアイドル集約のリポジトリインターフェース
type Repository interface {
	// Save は新しいアイドルを保存する
	Save(ctx context.Context, idol *Idol) error

	// FindByID はIDでアイドルを検索する
	FindByID(ctx context.Context, id IdolID) (*Idol, error)

	// FindAll は全てのアイドルを取得する
	FindAll(ctx context.Context) ([]*Idol, error)

	// Update は既存のアイドルを更新する
	Update(ctx context.Context, idol *Idol) error

	// Delete はアイドルを削除する
	Delete(ctx context.Context, id IdolID) error

	// ExistsByName は同じ名前のアイドルが存在するかチェック
	ExistsByName(ctx context.Context, name IdolName) (bool, error)

	Search(ctx context.Context, criteria SearchCriteria) ([]*Idol, error)
	
    Count(ctx context.Context, criteria SearchCriteria) (int64, error)
}
