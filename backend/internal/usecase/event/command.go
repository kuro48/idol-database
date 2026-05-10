package event

// CreateEventCommand はイベント作成コマンド
type CreateEventCommand struct {
	Title         string   `json:"title" binding:"required"`
	EventType     string   `json:"event_type" binding:"required"`
	StartDateTime string   `json:"start_date_time" binding:"required"` // RFC3339形式
	EndDateTime   *string  `json:"end_date_time,omitempty"`            // RFC3339形式（オプション）
	VenueID       *string  `json:"venue_id,omitempty"`
	PerformerIDs  []string `json:"performer_ids,omitempty"`
	TicketURL     *string  `json:"ticket_url,omitempty"`
	OfficialURL   *string  `json:"official_url,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// UpdateEventCommand はイベント更新コマンド
type UpdateEventCommand struct {
	ID            string  `json:"-"`
	Title         *string `json:"title,omitempty"`
	StartDateTime *string `json:"start_date_time,omitempty"` // RFC3339形式
	EndDateTime   *string `json:"end_date_time,omitempty"`   // RFC3339形式
	VenueID       *string `json:"venue_id,omitempty"`
	TicketURL     *string `json:"ticket_url,omitempty"`
	OfficialURL   *string `json:"official_url,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// DeleteEventCommand はイベント削除コマンド
type DeleteEventCommand struct {
	ID string `json:"-"`
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
