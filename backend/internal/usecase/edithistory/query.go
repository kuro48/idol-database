package edithistory

import "errors"

// GetEditHistoryQuery は単一の編集履歴取得クエリ
type GetEditHistoryQuery struct {
	ID string
}

// FieldChangeDTO はフィールド変更のデータ転送オブジェクト
type FieldChangeDTO struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// EditHistoryDTO は編集履歴のデータ転送オブジェクト
type EditHistoryDTO struct {
	ID         string                    `json:"id"`
	EntityType string                    `json:"entity_type"`
	EntityID   string                    `json:"entity_id"`
	Action     string                    `json:"action"`
	Changes    map[string]FieldChangeDTO `json:"changes"`
	ChangedBy  string                    `json:"changed_by"`
	CreatedAt  string                    `json:"created_at"`
}

// ListEditHistoryQuery は編集履歴一覧クエリ
type ListEditHistoryQuery struct {
	EntityType *string `form:"entity_type"`
	EntityID   *string `form:"entity_id"`
	Action     *string `form:"action"`
	ChangedBy  *string `form:"changed_by"`
	Page       *int    `form:"page"`
	Limit      *int    `form:"limit"`
}

func (q *ListEditHistoryQuery) ApplyDefaults() {
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
}

func (q *ListEditHistoryQuery) Validate() error {
	if q.EntityType != nil {
		valid := []string{"idol", "group", "agency", "event", "release", "venue", "membership"}
		if !contains(valid, *q.EntityType) {
			return errors.New("無効なエンティティ種別です")
		}
	}
	if q.Action != nil {
		valid := []string{"create", "update", "delete", "restore"}
		if !contains(valid, *q.Action) {
			return errors.New("無効なアクションです")
		}
	}
	return nil
}

// EditHistoryListResult は編集履歴一覧のレスポンス
type EditHistoryListResult struct {
	Data  []*EditHistoryDTO `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
