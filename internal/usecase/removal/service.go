package removal

import (
	"context"
	"fmt"

	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	domain "github.com/kuro48/idol-api/internal/domain/removal"
)

// Usecase は削除申請のユースケース
type Usecase struct {
	removalApp *appRemoval.ApplicationService
	idolApp    *appIdol.ApplicationService
	groupApp   *appGroup.ApplicationService
}

// NewUsecase はユースケースを作成する
func NewUsecase(removalApp *appRemoval.ApplicationService, idolApp *appIdol.ApplicationService, groupApp *appGroup.ApplicationService) *Usecase {
	return &Usecase{
		removalApp: removalApp,
		idolApp:    idolApp,
		groupApp:   groupApp,
	}
}

// CreateRemovalRequest は削除申請を作成する
func (u *Usecase) CreateRemovalRequest(ctx context.Context, cmd CreateRemovalRequestCommand) (*RemovalRequestDTO, error) {
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

	request, err := u.removalApp.CreateRemovalRequest(ctx, appRemoval.CreateInput{
		TargetType:  cmd.TargetType,
		TargetID:    cmd.TargetID,
		Requester:   cmd.Requester,
		Reason:      cmd.Reason,
		ContactInfo: cmd.ContactInfo,
		Evidence:    cmd.Evidence,
		Description: cmd.Description,
	})
	if err != nil {
		return nil, err
	}

	dto := toDTO(request)
	return &dto, nil
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
	case "rejected":
		if err := request.Reject(); err != nil {
			return nil, fmt.Errorf("却下に失敗しました: %w", err)
		}
	default:
		return nil, fmt.Errorf("無効なステータスです: %s", cmd.Status)
	}

	// 更新を保存
	if err := u.removalApp.UpdateRemovalRequest(ctx, request); err != nil {
		return nil, err
	}

	dto := toDTO(request)
	return &dto, nil
}

// toDTO はエンティティをDTOに変換する
func toDTO(request *domain.RemovalRequest) RemovalRequestDTO {
	return RemovalRequestDTO{
		ID:          request.ID().Value(),
		TargetID:    request.TargetID(),
		TargetType:  string(request.TargetType()),
		Requester:   string(request.Requester().Type()),
		Reason:      request.Reason().Value(),
		ContactInfo: request.ContactInfo().Value(),
		Evidence:    request.Evidence().Value(),
		Description: request.Description().Value(),
		Status:      string(request.Status()),
		CreatedAt:   request.CreatedAt(),
		UpdatedAt:   request.UpdatedAt(),
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
