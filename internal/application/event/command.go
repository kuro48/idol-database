package event

// CreateEventCommand はイベント作成コマンド
type CreateEventCommand struct {
	Title         string
	EventType     string
	StartDateTime string   // RFC3339形式
	EndDateTime   *string  // RFC3339形式（オプション）
	VenueID       *string
	PerformerIDs  []string
	TicketURL     *string
	OfficialURL   *string
	Description   *string
	Tags          []string
}

// UpdateEventCommand はイベント更新コマンド
type UpdateEventCommand struct {
	ID            string
	Title         *string
	StartDateTime *string // RFC3339形式
	EndDateTime   *string // RFC3339形式
	VenueID       *string
	TicketURL     *string
	OfficialURL   *string
	Description   *string
}

// DeleteEventCommand はイベント削除コマンド
type DeleteEventCommand struct {
	ID string
}

// AddPerformerCommand はパフォーマー追加コマンド
type AddPerformerCommand struct {
	EventID     string
	PerformerID string
}

// RemovePerformerCommand はパフォーマー削除コマンド
type RemovePerformerCommand struct {
	EventID     string
	PerformerID string
}
