package agency

// GetAgencyQuery は事務所取得クエリ
type GetAgencyQuery struct {
	ID      string   `uri:"id" binding:"required"`
	Include []string `form:"include"` // (未実装) idols, groups
}

// ListAgenciesQuery は事務所一覧取得クエリ（標準検索仕様準拠）
type ListAgenciesQuery struct {
	// フィルタ
	Name    *string `form:"name"`    // 部分一致検索
	Country *string `form:"country"` // 国コード完全一致

	// ソート
	Sort  *string `form:"sort"`  // name, founded_date, created_at
	Order *string `form:"order"` // asc, desc

	// ページネーション
	Page  *int `form:"page"`
	Limit *int `form:"limit"`
}

// Normalize はクエリパラメータをデフォルト値で正規化する
func (q *ListAgenciesQuery) Normalize() {
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

// AgencyDTO は事務所のデータ転送オブジェクト
type AgencyDTO struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	NameEn          *string `json:"name_en,omitempty"`
	FoundedDate     *string `json:"founded_date,omitempty"` // YYYY-MM-DD
	Country         string  `json:"country"`
	OfficialWebsite *string `json:"official_website,omitempty"`
	Description     *string `json:"description,omitempty"`
	LogoURL         *string `json:"logo_url,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
