package event

// PerformerInput はパフォーマー入力データ
type PerformerInput struct {
	PerformerID   string
	BillingStatus string
}

// CreateInput はイベント作成の入力
type CreateInput struct {
	Title         string
	EventType     string
	StartDateTime string
	EndDateTime   *string
	VenueID       *string
	Performers    []PerformerInput
	TicketURL     *string
	OfficialURL   *string
	Description   *string
	Tags          []string
}

// UpdateInput はイベント更新の入力
type UpdateInput struct {
	ID            string
	Title         *string
	StartDateTime *string
	EndDateTime   *string
	VenueID       *string
	TicketURL     *string
	OfficialURL   *string
	Description   *string
}

// AddPerformerInput はパフォーマー追加の入力
type AddPerformerInput struct {
	EventID       string
	PerformerID   string
	BillingStatus string
}

// RemovePerformerInput はパフォーマー削除の入力
type RemovePerformerInput struct {
	EventID     string
	PerformerID string
}
