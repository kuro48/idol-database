package membership

import (
	"context"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/membership"
	"github.com/kuro48/idol-api/internal/shared/audit"
)

type ApplicationService struct {
	repository membership.Repository
}

func NewApplicationService(repo membership.Repository) *ApplicationService {
	return &ApplicationService{repository: repo}
}

func (s *ApplicationService) CreateMembership(ctx context.Context, input CreateInput) (*membership.Membership, error) {
	role, err := membership.NewRole(input.Role)
	if err != nil {
		return nil, fmt.Errorf("ロールの生成エラー: %w", err)
	}

	var joinedAt *time.Time
	if input.JoinedAt != nil {
		t, err := time.Parse("2006-01-02", *input.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("加入日の形式が不正です: %w", err)
		}
		joinedAt = &t
	}

	m, err := membership.NewMembership(input.IdolID, input.GroupID, role, joinedAt)
	if err != nil {
		return nil, err
	}

	_ = audit.ActorFrom(ctx) // propagated to repository layer

	if err := s.repository.Save(ctx, m); err != nil {
		return nil, fmt.Errorf("メンバーシップの保存エラー: %w", err)
	}

	return m, nil
}

func (s *ApplicationService) GetMembership(ctx context.Context, id string) (*membership.Membership, error) {
	mid, err := membership.NewMembershipID(id)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	m, err := s.repository.FindByID(ctx, mid)
	if err != nil {
		return nil, fmt.Errorf("メンバーシップの取得エラー: %w", err)
	}

	return m, nil
}

func (s *ApplicationService) ListByIdolID(ctx context.Context, idolID string) ([]*membership.Membership, error) {
	ms, err := s.repository.FindByIdolID(ctx, idolID)
	if err != nil {
		return nil, fmt.Errorf("アイドルのメンバーシップ取得エラー: %w", err)
	}
	return ms, nil
}

func (s *ApplicationService) ListByGroupID(ctx context.Context, groupID string) ([]*membership.Membership, error) {
	ms, err := s.repository.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("グループのメンバーシップ取得エラー: %w", err)
	}
	return ms, nil
}

func (s *ApplicationService) SearchMemberships(ctx context.Context, criteria membership.SearchCriteria) ([]*membership.Membership, error) {
	ms, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("メンバーシップ検索エラー: %w", err)
	}
	return ms, nil
}

func (s *ApplicationService) CountMemberships(ctx context.Context, criteria membership.SearchCriteria) (int64, error) {
	return s.repository.Count(ctx, criteria)
}

func (s *ApplicationService) UpdateMembership(ctx context.Context, input UpdateInput) error {
	mid, err := membership.NewMembershipID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	m, err := s.repository.FindByID(ctx, mid)
	if err != nil {
		return fmt.Errorf("メンバーシップの取得エラー: %w", err)
	}

	if input.Role != nil {
		role, err := membership.NewRole(*input.Role)
		if err != nil {
			return fmt.Errorf("ロールの生成エラー: %w", err)
		}
		if err := m.UpdateRole(role); err != nil {
			return err
		}
	}

	if input.JoinedAt != nil {
		t, err := time.Parse("2006-01-02", *input.JoinedAt)
		if err != nil {
			return fmt.Errorf("加入日の形式が不正です: %w", err)
		}
		m.UpdateJoinedAt(&t)
	}

	if input.LeftAt != nil {
		if *input.LeftAt == "" {
			m.ClearLeftAt()
		} else {
			t, err := time.Parse("2006-01-02", *input.LeftAt)
			if err != nil {
				return fmt.Errorf("脱退日の形式が不正です: %w", err)
			}
			if err := m.Leave(t); err != nil {
				return err
			}
		}
	}

	if err := s.repository.Update(ctx, m); err != nil {
		return fmt.Errorf("メンバーシップの更新エラー: %w", err)
	}

	return nil
}

func (s *ApplicationService) DeleteMembership(ctx context.Context, id string) error {
	mid, err := membership.NewMembershipID(id)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, mid); err != nil {
		return fmt.Errorf("メンバーシップの削除エラー: %w", err)
	}

	return nil
}
