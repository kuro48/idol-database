package tag

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/tag"
)

// TagAppPort は tag.Usecase が tag application サービスに要求する契約
type TagAppPort interface {
	CreateTag(ctx context.Context, input TagCreateInput) (*domain.Tag, error)
	UpdateTag(ctx context.Context, input TagUpdateInput) error
	DeleteTag(ctx context.Context, id string) error
	GetTag(ctx context.Context, id string) (*domain.Tag, error)
	SearchTags(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Tag, int64, error)
}

// TagCreateInput はタグ作成の入力
type TagCreateInput struct {
	Name        string
	Category    string
	Description string
}

// TagUpdateInput はタグ更新の入力
type TagUpdateInput struct {
	ID          string
	Name        string
	Category    string
	Description string
}
