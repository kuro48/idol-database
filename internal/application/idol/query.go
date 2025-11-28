package idol

import "errors"

// GetIdolQuery はアイドル取得クエリ
type GetIdolQuery struct {
	ID string
}

// IdolDTO はアイドルのデータ転送オブジェクト
type IdolDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Birthdate   string `json:"birthdate,omitempty"`
	Age         *int   `json:"age,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListIdolsQuery はアイドル一覧取得クエリ
type ListIdolsQuery struct {
    // 検索条件
    Name        *string `form:"name"`         // 部分一致検索
    Nationality *string `form:"nationality"`  // 完全一致
    GroupID     *string `form:"group_id"`     // グループIDフィルター
    AgencyID    *string `form:"agency_id"`    // 事務所IDフィルター（後で実装）

    // 年齢範囲
    AgeMin      *int    `form:"age_min"`
    AgeMax      *int    `form:"age_max"`

    // 生年月日範囲
    BirthdateFrom *string `form:"birthdate_from"` // YYYY-MM-DD
    BirthdateTo   *string `form:"birthdate_to"`   // YYYY-MM-DD

    // タグ（将来実装）
    Tags        []string `form:"tags"`

    // ソート
    Sort        *string `form:"sort"`   // name, birthdate, created_at
    Order       *string `form:"order"`  // asc, desc

    // ページネーション
    Page        *int    `form:"page"`
    Limit       *int    `form:"limit"`
}

func (q *ListIdolsQuery) ApplyDefaults() {
    if q.Page == nil || *q.Page < 1 {
        defaultPage := 1
        q.Page = &defaultPage
    }
    if q.Limit == nil || *q.Limit < 1 {
        defaultLimit := 20
        q.Limit = &defaultLimit
    }
    if q.Limit != nil && *q.Limit > 100 {
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

// バリデーション
func (q *ListIdolsQuery) Validate() error {
    if q.Sort != nil {
        allowedSorts := []string{"name", "birthdate", "created_at"}
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

// contains はスライスに要素が含まれているかチェック
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
