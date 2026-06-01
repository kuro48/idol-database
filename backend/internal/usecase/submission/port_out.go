package submission

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/submission"
)

// SubmissionAppPort は Submission Usecase が application サービスに要求する契約
type SubmissionAppPort interface {
	// CreateSubmission は新しい投稿審査を作成する
	CreateSubmission(ctx context.Context, input SubmissionCreateInput) (*SubmissionCreateResult, error)

	// GetSubmission はIDで投稿審査を取得する
	GetSubmission(ctx context.Context, id string) (*domain.Submission, error)

	// ListAll は全ての投稿審査を取得する
	ListAll(ctx context.Context) ([]*domain.Submission, error)

	// ListPending は審査待ちの投稿審査を取得する
	ListPending(ctx context.Context) ([]*domain.Submission, error)

	// UpdateSubmission は投稿審査を更新する
	UpdateSubmission(ctx context.Context, submission *domain.Submission) error

	// FindByContributorIdentityID は投稿者 identity ID で投稿審査を取得する
	FindByContributorIdentityID(ctx context.Context, identityID string) ([]*domain.Submission, error)
}

// SubmissionCreateInput は投稿審査作成の入力
type SubmissionCreateInput struct {
	TargetType            string
	Payload               string
	SourceURLs            []string
	ContributorEmail      string
	ContributorIdentityID string
}

// SubmissionCreateResult はアプリケーション層の作成結果
type SubmissionCreateResult struct {
	Submission  *domain.Submission
	AccessToken string
}

// SubmissionTargetPort は承認済み投稿を本体データへ反映する Output Port
type SubmissionTargetPort interface {
	CreateIdol(ctx context.Context, input IdolCreateInput) error
	CreateGroup(ctx context.Context, input GroupCreateInput) error
	CreateAgency(ctx context.Context, input AgencyCreateInput) error
	CreateEvent(ctx context.Context, input EventCreateInput) error
}

// IdolCreateInput は承認済み idol 投稿の作成入力
type IdolCreateInput struct {
	Name      string   `json:"name"`
	Birthdate *string  `json:"birthdate,omitempty"`
	AgencyID  *string  `json:"agency_id,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
}

// GroupCreateInput は承認済み group 投稿の作成入力
type GroupCreateInput struct {
	Name          string  `json:"name"`
	FormationDate *string `json:"formation_date,omitempty"`
	DisbandDate   *string `json:"disband_date,omitempty"`
}

// AgencyCreateInput は承認済み agency 投稿の作成入力
type AgencyCreateInput struct {
	Name            string  `json:"name"`
	NameEn          *string `json:"name_en,omitempty"`
	FoundedDate     *string `json:"founded_date,omitempty"`
	Country         string  `json:"country"`
	OfficialWebsite *string `json:"official_website,omitempty"`
	Description     *string `json:"description,omitempty"`
	LogoURL         *string `json:"logo_url,omitempty"`
}

// EventCreateInput は承認済み event 投稿の作成入力
// EventPerformerInput はイベントパフォーマーの入力（投稿経由）
type EventPerformerInput struct {
	PerformerID   string `json:"performer_id"`
	BillingStatus string `json:"billing_status,omitempty"`
}

type EventCreateInput struct {
	Title         string                `json:"title"`
	EventType     string                `json:"event_type"`
	StartDateTime string                `json:"start_date_time"`
	EndDateTime   *string               `json:"end_date_time,omitempty"`
	VenueID       *string               `json:"venue_id,omitempty"`
	Performers    []EventPerformerInput `json:"performers,omitempty"`
	TicketURL     *string               `json:"ticket_url,omitempty"`
	OfficialURL   *string               `json:"official_url,omitempty"`
	Description   *string               `json:"description,omitempty"`
	Tags          []string              `json:"tags,omitempty"`
}

// EmailNotifier は審査結果メール送信の Output Port
type EmailNotifier interface {
	// NotifyStatusChanged は投稿審査のステータス変更を投稿者にメール通知する
	NotifyStatusChanged(ctx context.Context, notification StatusNotification) error
}

// StatusNotification はメール通知に必要な情報
type StatusNotification struct {
	To           string // 投稿者メールアドレス
	SubmissionID string
	TargetType   string // idol / group / agency / event
	Status       string // approved / rejected / needs_revision
	RevisionNote string // needs_revision 時のみ使用
}
