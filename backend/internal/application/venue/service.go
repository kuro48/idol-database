package venue

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/venue"
	"github.com/kuro48/idol-api/internal/shared/audit"
)

// ApplicationService は会場に関するアプリケーションサービス
type ApplicationService struct {
	repository venue.Repository
}

func NewApplicationService(repo venue.Repository) *ApplicationService {
	return &ApplicationService{repository: repo}
}

func (s *ApplicationService) CreateVenue(ctx context.Context, input CreateInput) (*venue.Venue, error) {
	v, err := venue.NewVenue(input.Name)
	if err != nil {
		return nil, err
	}

	if input.NameEn != nil {
		v.UpdateNameEn(input.NameEn)
	}
	if input.Prefecture != nil {
		v.UpdatePrefecture(input.Prefecture)
	}
	if input.City != nil {
		v.UpdateCity(input.City)
	}
	if input.Address != nil {
		v.UpdateAddress(input.Address)
	}
	if input.Capacity != nil {
		v.UpdateCapacity(input.Capacity)
	}
	if input.OfficialURL != nil {
		v.UpdateOfficialURL(input.OfficialURL)
	}

	_ = audit.ActorFrom(ctx) // 監査情報はリポジトリ層で使用

	if err := s.repository.Save(ctx, v); err != nil {
		return nil, fmt.Errorf("会場の保存エラー: %w", err)
	}

	return v, nil
}

func (s *ApplicationService) GetVenue(ctx context.Context, id string) (*venue.Venue, error) {
	vid, err := venue.NewVenueID(id)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	v, err := s.repository.FindByID(ctx, vid)
	if err != nil {
		return nil, fmt.Errorf("会場の取得エラー: %w", err)
	}

	return v, nil
}

func (s *ApplicationService) SearchVenues(ctx context.Context, criteria venue.SearchCriteria) ([]*venue.Venue, error) {
	vs, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("会場検索エラー: %w", err)
	}
	return vs, nil
}

func (s *ApplicationService) CountVenues(ctx context.Context, criteria venue.SearchCriteria) (int64, error) {
	return s.repository.Count(ctx, criteria)
}

func (s *ApplicationService) UpdateVenue(ctx context.Context, input UpdateInput) error {
	vid, err := venue.NewVenueID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	v, err := s.repository.FindByID(ctx, vid)
	if err != nil {
		return fmt.Errorf("会場の取得エラー: %w", err)
	}

	if input.Name != nil {
		if err := v.UpdateName(*input.Name); err != nil {
			return err
		}
	}
	if input.NameEn != nil {
		v.UpdateNameEn(input.NameEn)
	}
	if input.Prefecture != nil {
		v.UpdatePrefecture(input.Prefecture)
	}
	if input.City != nil {
		v.UpdateCity(input.City)
	}
	if input.Address != nil {
		v.UpdateAddress(input.Address)
	}
	if input.Capacity != nil {
		v.UpdateCapacity(input.Capacity)
	}
	if input.OfficialURL != nil {
		v.UpdateOfficialURL(input.OfficialURL)
	}

	if err := s.repository.Update(ctx, v); err != nil {
		return fmt.Errorf("会場の更新エラー: %w", err)
	}

	return nil
}

func (s *ApplicationService) DeleteVenue(ctx context.Context, id string) error {
	vid, err := venue.NewVenueID(id)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, vid); err != nil {
		return fmt.Errorf("会場の削除エラー: %w", err)
	}

	return nil
}
