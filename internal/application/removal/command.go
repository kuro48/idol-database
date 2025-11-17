package removal

// CreateRemovalRequestCommand は削除申請作成コマンド
type CreateRemovalRequestCommand struct {
	IdolID      string `json:"idol_id" binding:"required"`
	Requester   string `json:"requester" binding:"required"`
	Reason      string `json:"reason" binding:"required"`
	ContactInfo string `json:"contact_info" binding:"required,email"`
	Evidence    string `json:"evidence"`
	Description string `json:"description" binding:"required"`
}

// UpdateStatusCommand はステータス更新コマンド
type UpdateStatusCommand struct {
	ID     string `json:"id" binding:"required"`
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}
