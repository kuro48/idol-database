package edithistory

// RecordInput は編集履歴記録の入力
type RecordInput struct {
	EntityType string
	EntityID   string
	Action     string
	Changes    map[string]FieldChangeInput
	ChangedBy  string
}

// FieldChangeInput はフィールド変更の入力
type FieldChangeInput struct {
	Before interface{}
	After  interface{}
}
