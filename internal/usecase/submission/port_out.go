package submission

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/submission"
)

// SubmissionAppPort は Submission Usecase が application サービスに要求する契約
type SubmissionAppPort interface {
	// CreateSubmission は新しい投稿審査を作成する
	CreateSubmission(ctx context.Context, input SubmissionCreateInput) (*domain.Submission, error)

	// GetSubmission はIDで投稿審査を取得する
	GetSubmission(ctx context.Context, id string) (*domain.Submission, error)

	// ListAll は全ての投稿審査を取得する
	ListAll(ctx context.Context) ([]*domain.Submission, error)

	// ListPending は審査待ちの投稿審査を取得する
	ListPending(ctx context.Context) ([]*domain.Submission, error)

	// UpdateSubmission は投稿審査を更新する
	UpdateSubmission(ctx context.Context, submission *domain.Submission) error
}

// SubmissionCreateInput は投稿審査作成の入力
type SubmissionCreateInput struct {
	TargetType       string
	Payload          string
	SourceURLs       []string
	ContributorEmail string
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
