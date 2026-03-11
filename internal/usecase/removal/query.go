package removal

import "time"

// RemovalRequestDTO は削除申請のデータ転送オブジェクト（管理者用）
type RemovalRequestDTO struct {
	ID            string    `json:"id"`
	TargetID      string    `json:"target_id"`
	TargetType    string    `json:"target_type"`
	RequesterType string    `json:"requester_type"`
	Reason        string    `json:"reason"`
	ContactInfo   string    `json:"contact_info"`
	Evidence      string    `json:"evidence,omitempty"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PublicRemovalRequestDTO は公開用削除申請DTO（contact_info等の機微情報を除外）
type PublicRemovalRequestDTO struct {
	ID            string    `json:"id"`
	TargetID      string    `json:"target_id"`
	TargetType    string    `json:"target_type"`
	RequesterType string    `json:"requester_type"`
	Reason        string    `json:"reason"`
	Evidence      string    `json:"evidence,omitempty"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
