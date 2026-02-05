package event

import "errors"

// GetEventQuery はイベント取得クエリ
type GetEventQuery struct {
	ID string
}

// EventDTO はイベントのデータ転送オブジェクト
type EventDTO struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	EventType     string   `json:"event_type"`
	StartDateTime string   `json:"start_date_time"`
	EndDateTime   *string  `json:"end_date_time,omitempty"`
	VenueID       *string  `json:"venue_id,omitempty"`
	PerformerIDs  []string `json:"performer_ids"`
	TicketURL     *string  `json:"ticket_url,omitempty"`
	OfficialURL   *string  `json:"official_url,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Tags          []string `json:"tags"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ListEventsQuery はイベント一覧取得クエリ
type ListEventsQuery struct {
	// 検索条件
	EventType     *string `form:"event_type"`
	StartDateFrom *string `form:"start_date_from"` // YYYY-MM-DD
	StartDateTo   *string `form:"start_date_to"`   // YYYY-MM-DD
	VenueID       *string `form:"venue_id"`
	PerformerID   *string `form:"performer_id"`
	Tags          []string `form:"tags"`

	// ソート
	Sort  *string `form:"sort"`  // start_date_time, created_at
	Order *string `form:"order"` // asc, desc

	// ページネーション
	Page  *int `form:"page"`
	Limit *int `form:"limit"`
}

func (q *ListEventsQuery) ApplyDefaults() {
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
		defaultSort := "start_date_time"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "asc"
		q.Order = &defaultOrder
	}
}

// バリデーション
func (q *ListEventsQuery) Validate() error {
	if q.Sort != nil {
		allowedSorts := []string{"start_date_time", "created_at"}
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

// SearchResult は検索結果のレスポンス構造
type SearchResult struct {
	Data  []*EventDTO      `json:"data"`
	Meta  *PaginationMeta  `json:"meta"`
	Links *PaginationLinks `json:"links,omitempty"`
}

// PaginationMeta はページネーション情報
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginationLinks はページネーションリンク
type PaginationLinks struct {
	First string  `json:"first"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
	Last  string  `json:"last"`
}
