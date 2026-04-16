package submission

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	domain "github.com/kuro48/idol-api/internal/domain/submission"
)

// Usecase は投稿審査のユースケース実装
type Usecase struct {
	submissionApp SubmissionAppPort
	emailNotifier EmailNotifier // nil の場合はメール通知をスキップ
}

// NewUsecase はユースケースを作成する
func NewUsecase(submissionApp SubmissionAppPort, emailNotifier EmailNotifier) *Usecase {
	return &Usecase{
		submissionApp: submissionApp,
		emailNotifier: emailNotifier,
	}
}

// CreateSubmission は新しい投稿審査を作成する
func (u *Usecase) CreateSubmission(ctx context.Context, cmd CreateSubmissionCommand) (*PublicSubmissionDTO, error) {
	// Payload を JSON 文字列化
	payloadJSON, err := json.Marshal(cmd.Payload)
	if err != nil {
		return nil, fmt.Errorf("ペイロードのJSON変換に失敗しました: %w", err)
	}

	sub, err := u.submissionApp.CreateSubmission(ctx, SubmissionCreateInput{
		TargetType:       cmd.TargetType,
		Payload:          string(payloadJSON),
		SourceURLs:       cmd.SourceURLs,
		ContributorEmail: cmd.ContributorEmail,
	})
	if err != nil {
		return nil, err
	}

	return toPublicDTO(sub), nil
}

// GetSubmissionPublic は投稿審査を公開情報のみで取得する（投稿者向け）
func (u *Usecase) GetSubmissionPublic(ctx context.Context, id string) (*PublicSubmissionDTO, error) {
	sub, err := u.submissionApp.GetSubmission(ctx, id)
	if err != nil {
		return nil, err
	}

	return toPublicDTO(sub), nil
}

// ListAllSubmissions は全ての投稿審査を取得する（管理者向け）
func (u *Usecase) ListAllSubmissions(ctx context.Context) ([]*SubmissionDTO, error) {
	submissions, err := u.submissionApp.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return toAdminDTOs(submissions), nil
}

// ListPendingSubmissions は審査待ちの投稿審査を取得する（管理者向け）
func (u *Usecase) ListPendingSubmissions(ctx context.Context) ([]*SubmissionDTO, error) {
	submissions, err := u.submissionApp.ListPending(ctx)
	if err != nil {
		return nil, err
	}

	return toAdminDTOs(submissions), nil
}

// UpdateStatus は投稿審査のステータスを更新する（管理者向け）
func (u *Usecase) UpdateStatus(ctx context.Context, cmd UpdateStatusCommand) (*SubmissionDTO, error) {
	sub, err := u.submissionApp.GetSubmission(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	switch cmd.Status {
	case "approved":
		if err := sub.Approve(cmd.ReviewedBy); err != nil {
			return nil, fmt.Errorf("承認に失敗しました: %w", err)
		}
	case "rejected":
		if err := sub.Reject(cmd.ReviewedBy); err != nil {
			return nil, fmt.Errorf("却下に失敗しました: %w", err)
		}
	case "needs_revision":
		if err := sub.RequestRevision(cmd.ReviewedBy, cmd.RevisionNote); err != nil {
			return nil, fmt.Errorf("差し戻しに失敗しました: %w", err)
		}
	default:
		return nil, fmt.Errorf("無効なステータスです: %s", cmd.Status)
	}

	if err := u.submissionApp.UpdateSubmission(ctx, sub); err != nil {
		return nil, fmt.Errorf("ステータス更新の保存に失敗しました: %w", err)
	}

	// メール通知（失敗してもレスポンスに影響させない）
	if u.emailNotifier != nil {
		notification := StatusNotification{
			To:           sub.ContributorEmail().Value(),
			SubmissionID: sub.ID().Value(),
			TargetType:   string(sub.TargetType()),
			Status:       cmd.Status,
			RevisionNote: cmd.RevisionNote,
		}
		if err := u.emailNotifier.NotifyStatusChanged(ctx, notification); err != nil {
			slog.Error("メール通知に失敗しました",
				"error", err,
				"submission_id", sub.ID().Value(),
				"status", cmd.Status,
			)
		}
	}

	return toAdminDTO(sub), nil
}

// ReviseSubmission は差し戻し後の再投稿を行う（投稿者向け）
func (u *Usecase) ReviseSubmission(ctx context.Context, cmd ReviseSubmissionCommand) (*PublicSubmissionDTO, error) {
	sub, err := u.submissionApp.GetSubmission(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	// Payload を JSON 文字列化
	payloadJSON, err := json.Marshal(cmd.Payload)
	if err != nil {
		return nil, fmt.Errorf("ペイロードのJSON変換に失敗しました: %w", err)
	}

	// ステータスを needs_revision から pending に戻す
	if err := sub.Resubmit(); err != nil {
		return nil, fmt.Errorf("再投稿に失敗しました: %w", err)
	}

	// 再投稿時のペイロード・SourceURLs を更新した新しいエンティティを構築
	sourceURLObjs := make([]domain.SourceURL, 0, len(cmd.SourceURLs))
	for _, rawURL := range cmd.SourceURLs {
		srcURL, err := domain.NewSourceURL(rawURL)
		if err != nil {
			return nil, fmt.Errorf("無効な参照元URLです: %w", err)
		}
		sourceURLObjs = append(sourceURLObjs, srcURL)
	}

	updated := domain.Reconstruct(
		sub.ID(),
		sub.TargetType(),
		string(payloadJSON),
		sourceURLObjs,
		sub.ContributorEmail(),
		sub.SnsUserID(),
		sub.Status(),
		sub.RevisionNote(),
		sub.ReviewedBy(),
		sub.ReviewedAt(),
		sub.CreatedAt(),
		sub.UpdatedAt(),
	)

	if err := u.submissionApp.UpdateSubmission(ctx, updated); err != nil {
		return nil, fmt.Errorf("再投稿の保存に失敗しました: %w", err)
	}

	return toPublicDTO(updated), nil
}

// toPublicDTO はエンティティを投稿者向けDTOに変換する
func toPublicDTO(sub *domain.Submission) *PublicSubmissionDTO {
	sourceURLs := make([]string, 0, len(sub.SourceURLs()))
	for _, u := range sub.SourceURLs() {
		sourceURLs = append(sourceURLs, u.Value())
	}

	return &PublicSubmissionDTO{
		ID:           sub.ID().Value(),
		TargetType:   string(sub.TargetType()),
		Payload:      sub.Payload(),
		SourceURLs:   sourceURLs,
		Status:       string(sub.Status()),
		RevisionNote: sub.RevisionNote(),
		CreatedAt:    sub.CreatedAt(),
		UpdatedAt:    sub.UpdatedAt(),
	}
}

// toAdminDTO はエンティティを管理者向けDTOに変換する
func toAdminDTO(sub *domain.Submission) *SubmissionDTO {
	sourceURLs := make([]string, 0, len(sub.SourceURLs()))
	for _, u := range sub.SourceURLs() {
		sourceURLs = append(sourceURLs, u.Value())
	}

	return &SubmissionDTO{
		ID:               sub.ID().Value(),
		TargetType:       string(sub.TargetType()),
		Payload:          sub.Payload(),
		SourceURLs:       sourceURLs,
		ContributorEmail: sub.ContributorEmail().Value(),
		SnsUserID:        sub.SnsUserID(),
		Status:           string(sub.Status()),
		RevisionNote:     sub.RevisionNote(),
		ReviewedBy:       sub.ReviewedBy(),
		ReviewedAt:       sub.ReviewedAt(),
		CreatedAt:        sub.CreatedAt(),
		UpdatedAt:        sub.UpdatedAt(),
	}
}

// toAdminDTOs は複数のエンティティを管理者向けDTOに変換する
func toAdminDTOs(submissions []*domain.Submission) []*SubmissionDTO {
	dtos := make([]*SubmissionDTO, len(submissions))
	for i, sub := range submissions {
		dtos[i] = toAdminDTO(sub)
	}
	return dtos
}
