package removal

import "time"

// RemovalRequestDTO は削除申請のデータ転送オブジェクト
type RemovalRequestDTO struct {
	ID          string    `json:"id"`
	IdolID      string    `json:"idol_id"`
	Requester   string    `json:"requester"`
	Reason      string    `json:"reason"`
	ContactInfo string    `json:"contact_info"`
	Evidence    string    `json:"evidence,omitempty"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
