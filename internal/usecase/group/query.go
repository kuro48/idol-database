package group

// GetGroupQuery はグループ取得クエリ
type GetGroupQuery struct {
	ID string
}

// ListGroupQuery はグループ一覧取得クエリ（標準検索仕様準拠）
type ListGroupQuery struct {
	// フィルタ
	Name *string `form:"name"` // 部分一致検索

	// ソート
	Sort  *string `form:"sort"`  // name, formation_date, created_at
	Order *string `form:"order"` // asc, desc

	// ページネーション
	Page  *int `form:"page"`
	Limit *int `form:"limit"`
}

// Normalize はクエリパラメータをデフォルト値で正規化する
func (q *ListGroupQuery) Normalize() {
	if q.Page == nil || *q.Page < 1 {
		defaultPage := 1
		q.Page = &defaultPage
	}
	if q.Limit == nil || *q.Limit < 1 {
		defaultLimit := 20
		q.Limit = &defaultLimit
	}
	if *q.Limit > 100 {
		maxLimit := 100
		q.Limit = &maxLimit
	}
	if q.Sort == nil {
		defaultSort := "created_at"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "desc"
		q.Order = &defaultOrder
	}
}

// GroupDTO はグループのデータ転送オブジェクト
type GroupDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	FormationDate string `json:"formation_date,omitempty"`
	DisbandDate   string `json:"disband_date,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// GroupSearchResult はグループ検索結果
type GroupSearchResult struct {
	Data  []*GroupDTO      `json:"data"`
	Meta  *PaginationMeta  `json:"meta"`
}

// PaginationMeta はページネーションメタ情報
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}
