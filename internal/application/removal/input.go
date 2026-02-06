package removal

// CreateInput は削除申請作成の入力
type CreateInput struct {
	TargetType  string
	TargetID    string
	Requester   string
	Reason      string
	ContactInfo string
	Evidence    string
	Description string
}
