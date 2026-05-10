package tag

// CreateInput はタグ作成の入力
type CreateInput struct {
	Name        string
	Category    string
	Description string
}

// UpdateInput はタグ更新の入力
type UpdateInput struct {
	ID          string
	Name        string
	Category    string
	Description string
}
