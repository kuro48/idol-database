package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	domain "github.com/kuro48/idol-api/internal/domain/submission"
)

// Usecase は投稿審査のユースケース実装
type Usecase struct {
	submissionApp SubmissionAppPort
	targetPort    SubmissionTargetPort
	emailNotifier EmailNotifier // nil の場合はメール通知をスキップ
}

// NewUsecase はユースケースを作成する
func NewUsecase(submissionApp SubmissionAppPort, targetPort SubmissionTargetPort, emailNotifier EmailNotifier) *Usecase {
	return &Usecase{
		submissionApp: submissionApp,
		targetPort:    targetPort,
		emailNotifier: emailNotifier,
	}
}

// CreateSubmission は新しい投稿審査を作成する
func (u *Usecase) CreateSubmission(ctx context.Context, cmd CreateSubmissionCommand) (*CreateSubmissionResult, error) {
	// Payload を JSON 文字列化
	payloadJSON, err := json.Marshal(cmd.Payload)
	if err != nil {
		return nil, fmt.Errorf("ペイロードのJSON変換に失敗しました: %w", err)
	}
	if err := validateSubmissionPayload(cmd.TargetType, payloadJSON); err != nil {
		return nil, err
	}

	result, err := u.submissionApp.CreateSubmission(ctx, SubmissionCreateInput{
		TargetType:            cmd.TargetType,
		Payload:               string(payloadJSON),
		SourceURLs:            cmd.SourceURLs,
		ContributorEmail:      cmd.ContributorEmail,
		ContributorIdentityID: cmd.ContributorIdentityID,
	})
	if err != nil {
		return nil, err
	}

	return &CreateSubmissionResult{
		Submission:  toPublicDTO(result.Submission),
		AccessToken: result.AccessToken,
	}, nil
}

// GetSubmissionPublic は投稿審査を公開情報のみで取得する（投稿者向け）
func (u *Usecase) GetSubmissionPublic(ctx context.Context, id string, accessToken string) (*PublicSubmissionDTO, error) {
	sub, err := u.submissionApp.GetSubmission(ctx, id)
	if err != nil {
		return nil, err
	}
	if !sub.VerifyAccessToken(accessToken) {
		return nil, fmt.Errorf("投稿審査のアクセストークンが無効です")
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

// ListMySubmissions は認証済み本人の投稿審査一覧を取得する
func (u *Usecase) ListMySubmissions(ctx context.Context, subjectID string) ([]*PublicSubmissionDTO, error) {
	submissions, err := u.submissionApp.FindByContributorIdentityID(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*PublicSubmissionDTO, 0, len(submissions))
	for _, sub := range submissions {
		dtos = append(dtos, toPublicDTO(sub))
	}
	return dtos, nil
}

// UpdateStatus は投稿審査のステータスを更新する（管理者向け）
func (u *Usecase) UpdateStatus(ctx context.Context, cmd UpdateStatusCommand) (*SubmissionDTO, error) {
	sub, err := u.submissionApp.GetSubmission(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	switch cmd.Status {
	case "approved":
		if !sub.IsPending() {
			return nil, fmt.Errorf("承認に失敗しました: %w", domain.NewDomainError("承認できるのは審査待ちの投稿のみです"))
		}
		if err := u.applyApprovedSubmission(ctx, sub); err != nil {
			return nil, fmt.Errorf("承認対象データの反映に失敗しました: %w", err)
		}
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

func (u *Usecase) applyApprovedSubmission(ctx context.Context, sub *domain.Submission) error {
	if u.targetPort == nil {
		return fmt.Errorf("承認対象データの作成先が未設定です")
	}

	switch sub.TargetType() {
	case domain.SubmissionTypeIdol:
		input, err := decodeSubmissionPayload[IdolCreateInput](sub.Payload())
		if err != nil {
			return err
		}
		return u.targetPort.CreateIdol(ctx, input)
	case domain.SubmissionTypeGroup:
		input, err := decodeSubmissionPayload[GroupCreateInput](sub.Payload())
		if err != nil {
			return err
		}
		return u.targetPort.CreateGroup(ctx, input)
	case domain.SubmissionTypeAgency:
		input, err := decodeSubmissionPayload[AgencyCreateInput](sub.Payload())
		if err != nil {
			return err
		}
		return u.targetPort.CreateAgency(ctx, input)
	case domain.SubmissionTypeEvent:
		input, err := decodeSubmissionPayload[EventCreateInput](sub.Payload())
		if err != nil {
			return err
		}
		return u.targetPort.CreateEvent(ctx, input)
	default:
		return fmt.Errorf("未対応の投稿タイプです: %s", sub.TargetType())
	}
}

func decodeSubmissionPayload[T any](payload string) (T, error) {
	var input T
	if err := json.Unmarshal([]byte(payload), &input); err != nil {
		return input, fmt.Errorf("投稿ペイロードの解析に失敗しました: %w", err)
	}
	return input, nil
}

func validateSubmissionPayload(targetType string, payload []byte) error {
	switch targetType {
	case "idol":
		var input IdolCreateInput
		if err := decodeStrictPayload(payload, &input); err != nil {
			return err
		}
		if input.Name == "" {
			return fmt.Errorf("idol 投稿ペイロードには name が必須です")
		}
	case "group":
		var input GroupCreateInput
		if err := decodeStrictPayload(payload, &input); err != nil {
			return err
		}
		if input.Name == "" {
			return fmt.Errorf("group 投稿ペイロードには name が必須です")
		}
	case "agency":
		var input AgencyCreateInput
		if err := decodeStrictPayload(payload, &input); err != nil {
			return err
		}
		if input.Name == "" || input.Country == "" {
			return fmt.Errorf("agency 投稿ペイロードには name と country が必須です")
		}
	case "event":
		var input EventCreateInput
		if err := decodeStrictPayload(payload, &input); err != nil {
			return err
		}
		if input.Title == "" || input.EventType == "" || input.StartDateTime == "" {
			return fmt.Errorf("event 投稿ペイロードには title, event_type, start_date_time が必須です")
		}
	default:
		return fmt.Errorf("未対応の投稿タイプです: %s", targetType)
	}
	return nil
}

func decodeStrictPayload(payload []byte, dest interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("投稿ペイロードの形式が不正です: %w", err)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("投稿ペイロードに複数JSON値を含めることはできません")
	}
	return nil
}

// ReviseSubmission は差し戻し後の再投稿を行う（投稿者向け）
func (u *Usecase) ReviseSubmission(ctx context.Context, cmd ReviseSubmissionCommand) (*PublicSubmissionDTO, error) {
	sub, err := u.submissionApp.GetSubmission(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if !sub.VerifyAccessToken(cmd.AccessToken) {
		return nil, fmt.Errorf("投稿審査のアクセストークンが無効です")
	}

	// Payload を JSON 文字列化
	payloadJSON, err := json.Marshal(cmd.Payload)
	if err != nil {
		return nil, fmt.Errorf("ペイロードのJSON変換に失敗しました: %w", err)
	}
	if err := validateSubmissionPayload(string(sub.TargetType()), payloadJSON); err != nil {
		return nil, err
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
		sub.ContributorIdentityID(),
		sub.AccessTokenHash(),
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
