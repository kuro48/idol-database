package removal

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/group"
	"github.com/kuro48/idol-api/internal/domain/idol"
	"github.com/kuro48/idol-api/internal/domain/removal"
)

// ApplicationService は削除申請のアプリケーションサービス
type ApplicationService struct {
	removalRepo removal.Repository
	idolRepo    idol.Repository
	groupRepo   group.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(
	removalRepo removal.Repository,
	idolRepo idol.Repository,
	groupRepo group.Repository,
) *ApplicationService {
	return &ApplicationService{
		removalRepo: removalRepo,
		idolRepo:    idolRepo,
		groupRepo:   groupRepo,
	}
}

// CreateRemovalRequest は新しい削除申請を作成する
func (s *ApplicationService) CreateRemovalRequest(ctx context.Context, cmd CreateRemovalRequestCommand) (*RemovalRequestDTO, error) {
	// ターゲットタイプの検証
	targetType, err := removal.NewTargetType(cmd.TargetType)
	if err != nil {
		return nil, fmt.Errorf("無効なターゲットタイプです: %w", err)
	}

	// ターゲットの存在確認
	switch targetType {
	case removal.TargetTypeIdol:
		idolID, err := idol.NewIdolID(cmd.TargetID)
		if err != nil {
			return nil, fmt.Errorf("無効なアイドルIDです: %w", err)
		}
		_, err = s.idolRepo.FindByID(ctx, idolID)
		if err != nil {
			return nil, fmt.Errorf("指定されたアイドルが見つかりません: %w", err)
		}
	case removal.TargetTypeGroup:
		groupID, err := group.NewGroupID(cmd.TargetID)
		if err != nil {
			return nil, fmt.Errorf("無効なグループIDです: %w", err)
		}
		_, err = s.groupRepo.FindByID(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("指定されたグループが見つかりません: %w", err)
		}
	}

	// 申請者情報の作成
	requester, err := removal.NewRequester(cmd.Requester)
	if err != nil {
		return nil, fmt.Errorf("無効な申請者タイプです: %w", err)
	}

	// 削除理由の作成
	reason, err := removal.NewRemovalReason(cmd.Reason)
	if err != nil {
		return nil, fmt.Errorf("削除理由が無効です: %w", err)
	}

	// 連絡先情報の作成
	contactInfo, err := removal.NewContactInfo(cmd.ContactInfo)
	if err != nil {
		return nil, fmt.Errorf("連絡先情報が無効です: %w", err)
	}

	// 証拠資料URLの作成（オプショナル）
	evidence, err := removal.NewEvidenceURL(cmd.Evidence)
	if err != nil {
		return nil, fmt.Errorf("証拠資料URLが無効です: %w", err)
	}

	// 詳細説明の作成
	description, err := removal.NewRemovalReason(cmd.Description)
	if err != nil {
		return nil, fmt.Errorf("詳細説明が無効です: %w", err)
	}

	// 削除申請エンティティの作成
	request := removal.NewRemovalRequest(
		cmd.TargetID,
		targetType,
		requester,
		reason,
		contactInfo,
		evidence,
		description,
	)

	// 保存
	if err := s.removalRepo.Save(ctx, request); err != nil {
		return nil, fmt.Errorf("削除申請の保存に失敗しました: %w", err)
	}

	// DTOに変換して返す
	return toDTO(request), nil
}

// GetRemovalRequest は削除申請を取得する
func (s *ApplicationService) GetRemovalRequest(ctx context.Context, id string) (*RemovalRequestDTO, error) {
	removalID, err := removal.NewRemovalID(id)
	if err != nil {
		return nil, fmt.Errorf("無効な削除申請IDです: %w", err)
	}

	request, err := s.removalRepo.FindByID(ctx, removalID)
	if err != nil {
		return nil, fmt.Errorf("削除申請の取得に失敗しました: %w", err)
	}

	return toDTO(request), nil
}

// ListAllRemovalRequests は全ての削除申請を取得する
func (s *ApplicationService) ListAllRemovalRequests(ctx context.Context) ([]*RemovalRequestDTO, error) {
	requests, err := s.removalRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("削除申請一覧の取得に失敗しました: %w", err)
	}

	return toDTOs(requests), nil
}

// ListPendingRemovalRequests は保留中の削除申請を取得する
func (s *ApplicationService) ListPendingRemovalRequests(ctx context.Context) ([]*RemovalRequestDTO, error) {
	requests, err := s.removalRepo.FindPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("保留中の削除申請の取得に失敗しました: %w", err)
	}

	return toDTOs(requests), nil
}

// UpdateStatus はステータスを更新する
func (s *ApplicationService) UpdateStatus(ctx context.Context, cmd UpdateStatusCommand) (*RemovalRequestDTO, error) {
	// IDの検証
	removalID, err := removal.NewRemovalID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("無効な削除申請IDです: %w", err)
	}

	// 削除申請の取得
	request, err := s.removalRepo.FindByID(ctx, removalID)
	if err != nil {
		return nil, fmt.Errorf("削除申請の取得に失敗しました: %w", err)
	}

	// ステータス更新
	switch cmd.Status {
	case "approved":
		if err := request.Approve(); err != nil {
			return nil, fmt.Errorf("承認に失敗しました: %w", err)
		}

		// 承認時は対象データを物理削除
		switch request.TargetType() {
		case removal.TargetTypeIdol:
			idolID, err := idol.NewIdolID(request.TargetID())
			if err != nil {
				return nil, fmt.Errorf("アイドルIDの変換に失敗しました: %w", err)
			}
			if err := s.idolRepo.Delete(ctx, idolID); err != nil {
				return nil, fmt.Errorf("アイドルの削除に失敗しました: %w", err)
			}
		case removal.TargetTypeGroup:
			groupID, err := group.NewGroupID(request.TargetID())
			if err != nil {
				return nil, fmt.Errorf("グループIDの変換に失敗しました: %w", err)
			}
			if err := s.groupRepo.Delete(ctx, groupID); err != nil {
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
	if err := s.removalRepo.Update(ctx, request); err != nil {
		return nil, fmt.Errorf("ステータスの更新に失敗しました: %w", err)
	}

	return toDTO(request), nil
}

// toDTO はエンティティをDTOに変換する
func toDTO(request *removal.RemovalRequest) *RemovalRequestDTO {
	return &RemovalRequestDTO{
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
func toDTOs(requests []*removal.RemovalRequest) []*RemovalRequestDTO {
	dtos := make([]*RemovalRequestDTO, len(requests))
	for i, request := range requests {
		dtos[i] = toDTO(request)
	}
	return dtos
}
