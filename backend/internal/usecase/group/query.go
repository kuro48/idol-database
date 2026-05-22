package group

import "errors"

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

// Validate は検索条件の許可リスト検証を行う。
func (q *ListGroupQuery) Validate() error {
	if q.Sort != nil {
		allowedSorts := []string{"name", "formation_date", "created_at"}
		if !contains(allowedSorts, *q.Sort) {
			return errors.New("無効なソート項目です")
		}
	}
	if q.Order != nil {
		allowedOrders := []string{"asc", "desc"}
		if !contains(allowedOrders, *q.Order) {
			return errors.New("無効なソート順です")
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
	Data []*GroupDTO     `json:"data"`
	Meta *PaginationMeta `json:"meta"`
}

// PaginationMeta はページネーションメタ情報
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}
