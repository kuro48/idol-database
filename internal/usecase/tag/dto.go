package tag

import (
	"time"

	"github.com/kuro48/idol-api/internal/domain/tag"
)

// TagDTO はタグのデータ転送オブジェクト
type TagDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToDTO はドメインモデルをDTOに変換する
func ToDTO(t *tag.Tag) TagDTO {
	return TagDTO{
		ID:          t.ID().String(),
		Name:        t.Name().String(),
		Category:    t.Category().String(),
		Description: t.Description(),
		CreatedAt:   t.CreatedAt(),
	}
}

// SearchResult は検索結果
type SearchResult struct {
	Data  []TagDTO        `json:"data"`
	Meta  PaginationMeta  `json:"meta"`
	Links PaginationLinks `json:"links"`
}

// PaginationMeta はページネーション情報
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginationLinks はページネーションリンク
type PaginationLinks struct {
	First string `json:"first"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
	Last  string `json:"last"`
}
