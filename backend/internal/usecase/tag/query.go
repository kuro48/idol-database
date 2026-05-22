package tag

import (
	"github.com/kuro48/idol-api/internal/domain/tag"
)

// SearchQuery はタグ検索クエリ
type SearchQuery struct {
	Name     *string
	Category *string
	Page     int
	Limit    int
}

// ToCriteria はクエリをドメインの検索条件に変換する
func (q SearchQuery) ToCriteria() (tag.SearchCriteria, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	criteria := tag.SearchCriteria{
		Name:  q.Name,
		Page:  page,
		Limit: limit,
	}

	if q.Category != nil && *q.Category != "" {
		category, err := tag.NewTagCategory(*q.Category)
		if err != nil {
			return tag.SearchCriteria{}, err
		}
		criteria.Category = &category
	}

	return criteria, nil
}
