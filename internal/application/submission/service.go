package submission

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/submission"
)

// ApplicationService は投稿審査のアプリケーションサービス
type ApplicationService struct {
	submissionRepo submission.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(submissionRepo submission.Repository) *ApplicationService {
	return &ApplicationService{
		submissionRepo: submissionRepo,
	}
}

// CreateSubmission は新しい投稿審査を作成する
func (s *ApplicationService) CreateSubmission(ctx context.Context, input CreateInput) (*submission.Submission, error) {
	// 投稿タイプのバリデーション
	targetType, err := submission.NewSubmissionType(input.TargetType)
	if err != nil {
		return nil, fmt.Errorf("無効な投稿タイプです: %w", err)
	}

	// 参照元URLの変換
	sourceURLs := make([]submission.SourceURL, 0, len(input.SourceURLs))
	for _, rawURL := range input.SourceURLs {
		srcURL, err := submission.NewSourceURL(rawURL)
		if err != nil {
			return nil, fmt.Errorf("無効な参照元URLです: %w", err)
		}
		sourceURLs = append(sourceURLs, srcURL)
	}

	// 投稿者メールアドレスのバリデーション
	contributorEmail, err := submission.NewContributorEmail(input.ContributorEmail)
	if err != nil {
		return nil, fmt.Errorf("無効な投稿者メールアドレスです: %w", err)
	}

	// 投稿審査エンティティの作成
	sub := submission.NewSubmission(
		targetType,
		input.Payload,
		sourceURLs,
		contributorEmail,
	)

	// 保存
	if err := s.submissionRepo.Save(ctx, sub); err != nil {
		return nil, fmt.Errorf("投稿審査の保存に失敗しました: %w", err)
	}

	return sub, nil
}

// GetSubmission は投稿審査を取得する
func (s *ApplicationService) GetSubmission(ctx context.Context, id string) (*submission.Submission, error) {
	submissionID, err := submission.NewSubmissionID(id)
	if err != nil {
		return nil, fmt.Errorf("無効な投稿審査IDです: %w", err)
	}

	sub, err := s.submissionRepo.FindByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("投稿審査の取得に失敗しました: %w", err)
	}

	return sub, nil
}

// ListAll は全ての投稿審査を取得する
func (s *ApplicationService) ListAll(ctx context.Context) ([]*submission.Submission, error) {
	submissions, err := s.submissionRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("投稿審査一覧の取得に失敗しました: %w", err)
	}

	return submissions, nil
}

// ListPending は審査待ちの投稿審査を取得する
func (s *ApplicationService) ListPending(ctx context.Context) ([]*submission.Submission, error) {
	submissions, err := s.submissionRepo.FindPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("審査待ち投稿審査の取得に失敗しました: %w", err)
	}

	return submissions, nil
}

// UpdateSubmission は投稿審査を更新する
func (s *ApplicationService) UpdateSubmission(ctx context.Context, sub *submission.Submission) error {
	if err := s.submissionRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("投稿審査の更新に失敗しました: %w", err)
	}

	return nil
}

// FindByContributorEmail は投稿者メールアドレスで投稿審査を取得する
func (s *ApplicationService) FindByContributorEmail(ctx context.Context, email string) ([]*submission.Submission, error) {
	submissions, err := s.submissionRepo.FindByContributorEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("投稿者メールアドレスによる投稿審査の取得に失敗しました: %w", err)
	}

	return submissions, nil
}
