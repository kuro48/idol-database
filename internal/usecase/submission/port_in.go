package submission

import "context"

// SubmissionUseCase は投稿審査のユースケース Input Port
type SubmissionUseCase interface {
	// CreateSubmission は新しい投稿審査を作成する
	CreateSubmission(ctx context.Context, cmd CreateSubmissionCommand) (*PublicSubmissionDTO, error)

	// GetSubmissionPublic は投稿審査を公開情報のみで取得する（投稿者向け）
	GetSubmissionPublic(ctx context.Context, id string) (*PublicSubmissionDTO, error)

	// ListAllSubmissions は全ての投稿審査を取得する（管理者向け）
	ListAllSubmissions(ctx context.Context) ([]*SubmissionDTO, error)

	// ListPendingSubmissions は審査待ちの投稿審査を取得する（管理者向け）
	ListPendingSubmissions(ctx context.Context) ([]*SubmissionDTO, error)

	// UpdateStatus は投稿審査のステータスを更新する（管理者向け）
	UpdateStatus(ctx context.Context, cmd UpdateStatusCommand) (*SubmissionDTO, error)

	// ReviseSubmission は差し戻し後の再投稿を行う（投稿者向け）
	ReviseSubmission(ctx context.Context, cmd ReviseSubmissionCommand) (*PublicSubmissionDTO, error)
}
