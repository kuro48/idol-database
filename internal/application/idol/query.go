package idol

// GetIdolQuery はアイドル取得クエリ
type GetIdolQuery struct {
	ID string
}

// ListIdolsQuery はアイドル一覧取得クエリ
type ListIdolsQuery struct {
	// 将来的にページネーションやフィルタを追加可能
	Limit  int
	Offset int
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
