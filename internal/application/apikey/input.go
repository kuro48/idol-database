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
