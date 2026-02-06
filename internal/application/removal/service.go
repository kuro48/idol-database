package removal

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/removal"
)

// ApplicationService は削除申請のアプリケーションサービス
type ApplicationService struct {
	removalRepo removal.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(removalRepo removal.Repository) *ApplicationService {
	return &ApplicationService{
		removalRepo: removalRepo,
	}
}

// CreateRemovalRequest は新しい削除申請を作成する
func (s *ApplicationService) CreateRemovalRequest(ctx context.Context, input CreateInput) (*removal.RemovalRequest, error) {
	// ターゲットタイプの検証
	targetType, err := removal.NewTargetType(input.TargetType)
	if err != nil {
		return nil, fmt.Errorf("無効なターゲットタイプです: %w", err)
	}

	// 申請者情報の作成
	requester, err := removal.NewRequester(input.Requester)
	if err != nil {
		return nil, fmt.Errorf("無効な申請者タイプです: %w", err)
	}

	// 削除理由の作成
	reason, err := removal.NewRemovalReason(input.Reason)
	if err != nil {
		return nil, fmt.Errorf("削除理由が無効です: %w", err)
	}

	// 連絡先情報の作成
	contactInfo, err := removal.NewContactInfo(input.ContactInfo)
	if err != nil {
		return nil, fmt.Errorf("連絡先情報が無効です: %w", err)
	}

	// 証拠資料URLの作成（オプショナル）
	evidence, err := removal.NewEvidenceURL(input.Evidence)
	if err != nil {
		return nil, fmt.Errorf("証拠資料URLが無効です: %w", err)
	}

	// 詳細説明の作成
	description, err := removal.NewRemovalReason(input.Description)
	if err != nil {
		return nil, fmt.Errorf("詳細説明が無効です: %w", err)
	}

	// 削除申請エンティティの作成
	request := removal.NewRemovalRequest(
		input.TargetID,
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

	return request, nil
}

// GetRemovalRequest は削除申請を取得する
func (s *ApplicationService) GetRemovalRequest(ctx context.Context, id string) (*removal.RemovalRequest, error) {
	removalID, err := removal.NewRemovalID(id)
	if err != nil {
		return nil, fmt.Errorf("無効な削除申請IDです: %w", err)
	}

	request, err := s.removalRepo.FindByID(ctx, removalID)
	if err != nil {
		return nil, fmt.Errorf("削除申請の取得に失敗しました: %w", err)
	}

	return request, nil
}

// ListAllRemovalRequests は全ての削除申請を取得する
func (s *ApplicationService) ListAllRemovalRequests(ctx context.Context) ([]*removal.RemovalRequest, error) {
	requests, err := s.removalRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("削除申請一覧の取得に失敗しました: %w", err)
	}

	return requests, nil
}

// ListPendingRemovalRequests は保留中の削除申請を取得する
func (s *ApplicationService) ListPendingRemovalRequests(ctx context.Context) ([]*removal.RemovalRequest, error) {
	requests, err := s.removalRepo.FindPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("保留中の削除申請の取得に失敗しました: %w", err)
	}

	return requests, nil
}

// UpdateRemovalRequest は削除申請を更新する
func (s *ApplicationService) UpdateRemovalRequest(ctx context.Context, request *removal.RemovalRequest) error {
	if err := s.removalRepo.Update(ctx, request); err != nil {
		return fmt.Errorf("削除申請の更新に失敗しました: %w", err)
	}

	return nil
}
