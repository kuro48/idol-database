package tag

import "context"

// Repository はタグリポジトリのインターフェース
type Repository interface {
	// Save はタグを保存する
	Save(ctx context.Context, tag *Tag) error

	// Update はタグを更新する
	Update(ctx context.Context, tag *Tag) error

	// Delete はタグを削除する
	Delete(ctx context.Context, id TagID) error

	// FindByID はIDでタグを検索する
	FindByID(ctx context.Context, id TagID) (*Tag, error)

	// FindByName は名前でタグを検索する（完全一致）
	FindByName(ctx context.Context, name string) (*Tag, error)

	// FindByCategory はカテゴリでタグを検索する
	FindByCategory(ctx context.Context, category TagCategory) ([]*Tag, error)

	// Search は検索条件に基づいてタグを検索する
	Search(ctx context.Context, criteria SearchCriteria) ([]*Tag, int64, error)

	// Exists はタグが存在するか確認する
	Exists(ctx context.Context, id TagID) (bool, error)
}

// SearchCriteria はタグ検索の条件
type SearchCriteria struct {
	Name     *string      // 名前（部分一致）
	Category *TagCategory // カテゴリ
	Page     int          // ページ番号（1始まり）
	Limit    int          // 1ページあたりの件数
}
