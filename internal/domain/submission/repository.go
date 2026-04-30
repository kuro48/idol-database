package submission

import "context"

// Repository は投稿審査リポジトリのインターフェース
type Repository interface {
	// Save は新しい投稿審査を保存する
	Save(ctx context.Context, submission *Submission) error

	// FindByID はIDで投稿審査を取得する
	FindByID(ctx context.Context, id SubmissionID) (*Submission, error)

	// FindAll は全ての投稿審査を取得する
	FindAll(ctx context.Context) ([]*Submission, error)

	// FindPending は審査待ちの投稿審査を取得する
	FindPending(ctx context.Context) ([]*Submission, error)

	// FindByContributorEmail は投稿者メールアドレスで投稿審査を取得する
	FindByContributorEmail(ctx context.Context, email string) ([]*Submission, error)

	// Update は投稿審査を更新する
	Update(ctx context.Context, submission *Submission) error
}
