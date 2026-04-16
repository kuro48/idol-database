package submission

import "time"

// PublicSubmissionDTO は投稿者向け公開用DTO（メール等の機微情報を除外）
type PublicSubmissionDTO struct {
	ID           string    `json:"id"`
	TargetType   string    `json:"target_type"`
	Payload      string    `json:"payload"`
	SourceURLs   []string  `json:"source_urls"`
	Status       string    `json:"status"`
	RevisionNote string    `json:"revision_note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SubmissionDTO は管理者向けDTO（全フィールド）
type SubmissionDTO struct {
	ID               string     `json:"id"`
	TargetType       string     `json:"target_type"`
	Payload          string     `json:"payload"`
	SourceURLs       []string   `json:"source_urls"`
	ContributorEmail string     `json:"contributor_email"`
	SnsUserID        string     `json:"sns_user_id,omitempty"`
	Status           string     `json:"status"`
	RevisionNote     string     `json:"revision_note,omitempty"`
	ReviewedBy       string     `json:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
