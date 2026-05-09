package apikey

// CreateKeyInput はAPIキー作成の入力
type CreateKeyInput struct {
	Email    string
	Name     string
	PlanType string // "free" | "developer" | "business"
}

// RevokeKeyInput はAPIキー無効化の入力
type RevokeKeyInput struct {
	ID string // MongoDB ObjectID hex
}

// UpdateOshiColorInput は推しメンカラー更新の入力
type UpdateOshiColorInput struct {
	ID        string // MongoDB ObjectID hex
	OshiColor string // CSS hex color (#RGB or #RRGGBB)、空文字はクリア
}
