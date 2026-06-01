package venue

import "errors"

// GetVenueQuery は会場詳細取得クエリ
type GetVenueQuery struct {
	ID string
}

// ListVenueQuery は会場一覧取得クエリ
type ListVenueQuery struct {
	Name       *string `form:"name"`
	Prefecture *string `form:"prefecture"`
	Sort       *string `form:"sort"`
	Order      *string `form:"order"`
	Page       *int    `form:"page"`
	Limit      *int    `form:"limit"`
}

func (q *ListVenueQuery) Normalize() {
	if q.Page == nil || *q.Page < 1 {
		p := 1
		q.Page = &p
	}
	if q.Limit == nil || *q.Limit < 1 {
		l := 20
		q.Limit = &l
	}
	if *q.Limit > 100 {
		l := 100
		q.Limit = &l
	}
	if q.Sort == nil {
		s := "created_at"
		q.Sort = &s
	}
	if q.Order == nil {
		o := "desc"
		q.Order = &o
	}
}

func (q *ListVenueQuery) Validate() error {
	if q.Sort != nil {
		allowed := []string{"name", "created_at"}
		if !containsStr(allowed, *q.Sort) {
			return errors.New("無効なソート項目です")
		}
	}
	if q.Order != nil {
		if *q.Order != "asc" && *q.Order != "desc" {
			return errors.New("無効なソート順です")
		}
	}
	return nil
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// VenueDTO は会場の転送オブジェクト
type VenueDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	NameEn      *string `json:"name_en,omitempty"`
	Prefecture  *string `json:"prefecture,omitempty"`
	City        *string `json:"city,omitempty"`
	Address     *string `json:"address,omitempty"`
	Capacity    *int    `json:"capacity,omitempty"`
	OfficialURL *string `json:"official_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// VenueSearchResult は会場検索結果
type VenueSearchResult struct {
	Data []*VenueDTO     `json:"data"`
	Meta *PaginationMeta `json:"meta"`
}

// PaginationMeta はページネーションのメタ情報
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}
