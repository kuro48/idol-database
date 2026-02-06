package tag

// CreateTagCommand はタグ作成コマンド
type CreateTagCommand struct {
	Name        string `json:"name" binding:"required,max=50"`
	Category    string `json:"category" binding:"required,oneof=genre region style other"`
	Description string `json:"description" binding:"max=200"`
}

// UpdateTagCommand はタグ更新コマンド
type UpdateTagCommand struct {
	ID          string
	Name        string `json:"name" binding:"required,max=50"`
	Category    string `json:"category" binding:"required,oneof=genre region style other"`
	Description string `json:"description" binding:"max=200"`
}
