package removal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	domain "github.com/kuro48/idol-api/internal/domain/removal"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
)

// Usecase は削除申請のユースケース
type Usecase struct {
	removalApp RemovalAppPort
	idolApp    RemovalIdolPort
	groupApp   RemovalGroupPort
	notifier   RemovalNotifier
	publisher  RemovalWebhookPublisher
}

const removalSLAWindow = 72 * time.Hour

// NewUsecase はユースケースを作成する
func NewUsecase(removalApp RemovalAppPort, idolApp RemovalIdolPort, groupApp RemovalGroupPort, notifier RemovalNotifier, publisher RemovalWebhookPublisher) *Usecase {
	return &Usecase{
		removalApp: removalApp,
		idolApp:    idolApp,
		groupApp:   groupApp,
		notifier:   notifier,
		publisher:  publisher,
	}
}

// CreateRemovalRequest は削除申請を作成する
func (u *Usecase) CreateRemovalRequest(ctx context.Context, cmd CreateRemovalRequestCommand) (*CreateRemovalRequestResult, error) {
	// ターゲットタイプの検証
	targetType, err := domain.NewTargetType(cmd.TargetType)
	if err != nil {
		return nil, fmt.Errorf("無効なターゲットタイプです: %w", err)
	}

	// ターゲットの存在確認
	switch targetType {
	case domain.TargetTypeIdol:
		if _, err := u.idolApp.GetIdol(ctx, cmd.TargetID); err != nil {
			return nil, fmt.Errorf("指定されたアイドルが見つかりません: %w", err)
		}
	case domain.TargetTypeGroup:
		if _, err := u.groupApp.GetGroup(ctx, cmd.TargetID); err != nil {
			return nil, fmt.Errorf("指定されたグループが見つかりません: %w", err)
		}
	}

	result, err := u.removalApp.CreateRemovalRequest(ctx, RemovalCreateInput{
		TargetType:  cmd.TargetType,
		TargetID:    cmd.TargetID,
		Requester:   cmd.RequesterType,
		Reason:      cmd.Reason,
		ContactInfo: cmd.ContactInfo,
		Evidence:    cmd.Evidence,
		Description: cmd.Description,
	})
	if err != nil {
		return nil, err
	}

	dto := toDTO(result.Request)
	u.notifyReceived(ctx, result.Request, result.AccessToken)
	return &CreateRemovalRequestResult{
		RemovalRequest: &dto,
		AccessToken:    result.AccessToken,
	}, nil
}

// GetRemovalRequest は削除申請を取得する
func (u *Usecase) GetRemovalRequest(ctx context.Context, id string) (*RemovalRequestDTO, error) {
	request, err := u.removalApp.GetRemovalRequest(ctx, id)
	if err != nil {
		return nil, err
	}

	dto := toDTO(request)
	return &dto, nil
}

// GetRemovalRequestPublic は削除申請を公開情報のみで取得する（contact_info等の機微情報を除外）
func (u *Usecase) GetRemovalRequestPublic(ctx context.Context, id string, accessToken string) (*PublicRemovalRequestDTO, error) {
	request, err := u.removalApp.GetRemovalRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if !request.VerifyAccessToken(accessToken) {
		return nil, fmt.Errorf("削除申請のアクセストークンが無効です")
	}

	dto := &PublicRemovalRequestDTO{
		ID:            request.ID().Value(),
		TargetID:      request.TargetID(),
		TargetType:    string(request.TargetType()),
		RequesterType: string(request.Requester().Type()),
		Reason:        request.Reason().Value(),
		Evidence:      request.Evidence().Value(),
		Description:   request.Description().Value(),
		Status:        string(request.Status()),
		CreatedAt:     request.CreatedAt(),
		UpdatedAt:     request.UpdatedAt(),
	}
	return dto, nil
}

// ListAllRemovalRequests は全ての削除申請を取得する
func (u *Usecase) ListAllRemovalRequests(ctx context.Context) ([]*RemovalRequestDTO, error) {
	requests, err := u.removalApp.ListAllRemovalRequests(ctx)
	if err != nil {
		return nil, err
	}

	return toDTOs(requests), nil
}

// ListPendingRemovalRequests は保留中の削除申請を取得する
func (u *Usecase) ListPendingRemovalRequests(ctx context.Context) ([]*RemovalRequestDTO, error) {
	requests, err := u.removalApp.ListPendingRemovalRequests(ctx)
	if err != nil {
		return nil, err
	}

	return toDTOs(requests), nil
}

// UpdateStatus はステータスを更新する
func (u *Usecase) UpdateStatus(ctx context.Context, cmd UpdateStatusCommand) (*RemovalRequestDTO, error) {
	// 削除申請の取得
	request, err := u.removalApp.GetRemovalRequest(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	switch cmd.Status {
	case "approved":
		if err := request.Approve(); err != nil {
			return nil, fmt.Errorf("承認に失敗しました: %w", err)
		}

		// 先にステータス更新を保存してから対象データを削除
		// これにより、削除が失敗してもステータスは正しく更新される
		if err := u.removalApp.UpdateRemovalRequest(ctx, request); err != nil {
			return nil, fmt.Errorf("ステータス更新の保存に失敗しました: %w", err)
		}

		// 承認時は対象データを物理削除
		switch request.TargetType() {
		case domain.TargetTypeIdol:
			if err := u.idolApp.DeleteIdol(ctx, request.TargetID()); err != nil {
				return nil, fmt.Errorf("アイドルの削除に失敗しました: %w", err)
			}
		case domain.TargetTypeGroup:
			if err := u.groupApp.DeleteGroup(ctx, request.TargetID()); err != nil {
				return nil, fmt.Errorf("グループの削除に失敗しました: %w", err)
			}
		}

		u.publishWebhook(ctx, domainWebhook.EventRemovalApproved, map[string]interface{}{
			"id":          request.ID().Value(),
			"target_id":   request.TargetID(),
			"target_type": string(request.TargetType()),
			"status":      string(request.Status()),
		})
	case "rejected":
		if err := request.Reject(); err != nil {
			return nil, fmt.Errorf("却下に失敗しました: %w", err)
		}

		// 却下時はステータス更新のみ保存
		if err := u.removalApp.UpdateRemovalRequest(ctx, request); err != nil {
			return nil, fmt.Errorf("ステータス更新の保存に失敗しました: %w", err)
		}
	default:
		return nil, fmt.Errorf("無効なステータスです: %s", cmd.Status)
	}

	dto := toDTO(request)
	u.notifyResolved(ctx, request)
	return &dto, nil
}

// toDTO はエンティティをDTOに変換する
func toDTO(request *domain.RemovalRequest) RemovalRequestDTO {
	slaDueAt := request.CreatedAt().Add(removalSLAWindow)
	slaOverdue := request.IsPending() && time.Now().After(slaDueAt)

	return RemovalRequestDTO{
		ID:            request.ID().Value(),
		TargetID:      request.TargetID(),
		TargetType:    string(request.TargetType()),
		RequesterType: string(request.Requester().Type()),
		Reason:        request.Reason().Value(),
		ContactInfo:   request.ContactInfo().Value(),
		Evidence:      request.Evidence().Value(),
		Description:   request.Description().Value(),
		Status:        string(request.Status()),
		CreatedAt:     request.CreatedAt(),
		UpdatedAt:     request.UpdatedAt(),
		SLADueAt:      slaDueAt,
		SLAOverdue:    slaOverdue,
	}
}

// toDTOs は複数のエンティティをDTOに変換する
func toDTOs(requests []*domain.RemovalRequest) []*RemovalRequestDTO {
	dtos := make([]*RemovalRequestDTO, len(requests))
	for i, request := range requests {
		dto := toDTO(request)
		dtos[i] = &dto
	}
	return dtos
}

func (u *Usecase) publishWebhook(ctx context.Context, event domainWebhook.EventType, payload interface{}) {
	if u.publisher == nil {
		return
	}
	if err := u.publisher.Publish(ctx, event, payload); err != nil {
		slog.Error("削除申請Webhook配信キュー投入に失敗しました", "event", event, "error", err)
	}
}

func (u *Usecase) notifyReceived(ctx context.Context, request *domain.RemovalRequest, accessToken string) {
	if u.notifier == nil {
		return
	}
	if err := u.notifier.NotifyReceived(ctx, ReceivedNotification{
		To:          request.ContactInfo().Value(),
		RequestID:   request.ID().Value(),
		TargetType:  string(request.TargetType()),
		AccessToken: accessToken,
		CreatedAt:   request.CreatedAt(),
	}); err != nil {
		slog.Error("削除申請受付通知に失敗しました", "request_id", request.ID().Value(), "error", err)
	}
}

func (u *Usecase) notifyResolved(ctx context.Context, request *domain.RemovalRequest) {
	if u.notifier == nil {
		return
	}
	if err := u.notifier.NotifyResolved(ctx, ResolvedNotification{
		To:         request.ContactInfo().Value(),
		RequestID:  request.ID().Value(),
		TargetType: string(request.TargetType()),
		Status:     string(request.Status()),
		UpdatedAt:  request.UpdatedAt(),
	}); err != nil {
		slog.Error("削除申請完了通知に失敗しました", "request_id", request.ID().Value(), "status", request.Status(), "error", err)
	}
}
