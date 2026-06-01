package venue

import (
	"context"
	"fmt"

	domain "github.com/kuro48/idol-api/internal/domain/venue"
)

// Usecase は会場ユースケースの実装
type Usecase struct {
	appService VenueAppPort
}

func NewUsecase(appService VenueAppPort) *Usecase {
	return &Usecase{appService: appService}
}

func (u *Usecase) CreateVenue(ctx context.Context, cmd CreateVenueCommand) (*VenueDTO, error) {
	v, err := u.appService.CreateVenue(ctx, VenueCreateInput{
		Name:        cmd.Name,
		NameEn:      cmd.NameEn,
		Prefecture:  cmd.Prefecture,
		City:        cmd.City,
		Address:     cmd.Address,
		Capacity:    cmd.Capacity,
		OfficialURL: cmd.OfficialURL,
	})
	if err != nil {
		return nil, err
	}
	dto := toDTO(v)
	return &dto, nil
}

func (u *Usecase) GetVenue(ctx context.Context, query GetVenueQuery) (*VenueDTO, error) {
	v, err := u.appService.GetVenue(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	dto := toDTO(v)
	return &dto, nil
}

func (u *Usecase) ListVenues(ctx context.Context, query ListVenueQuery) (*VenueSearchResult, error) {
	query.Normalize()
	if err := query.Validate(); err != nil {
		return nil, err
	}

	criteria := domain.SearchCriteria{
		Name:       query.Name,
		Prefecture: query.Prefecture,
		Sort:       *query.Sort,
		Order:      *query.Order,
		Offset:     (*query.Page - 1) * *query.Limit,
		Limit:      *query.Limit,
	}

	vs, err := u.appService.SearchVenues(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("会場一覧の取得エラー: %w", err)
	}

	total, err := u.appService.CountVenues(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", err)
	}

	dtos := make([]*VenueDTO, 0, len(vs))
	for _, v := range vs {
		dto := toDTO(v)
		dtos = append(dtos, &dto)
	}

	totalPages := int(total) / *query.Limit
	if int(total)%*query.Limit != 0 {
		totalPages++
	}

	return &VenueSearchResult{
		Data: dtos,
		Meta: &PaginationMeta{
			Total:      total,
			Page:       *query.Page,
			PerPage:    *query.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (u *Usecase) UpdateVenue(ctx context.Context, cmd UpdateVenueCommand) error {
	return u.appService.UpdateVenue(ctx, VenueUpdateInput{
		ID:          cmd.ID,
		Name:        cmd.Name,
		NameEn:      cmd.NameEn,
		Prefecture:  cmd.Prefecture,
		City:        cmd.City,
		Address:     cmd.Address,
		Capacity:    cmd.Capacity,
		OfficialURL: cmd.OfficialURL,
	})
}

func (u *Usecase) DeleteVenue(ctx context.Context, cmd DeleteVenueCommand) error {
	return u.appService.DeleteVenue(ctx, cmd.ID)
}

func toDTO(v *domain.Venue) VenueDTO {
	return VenueDTO{
		ID:          v.ID().Value(),
		Name:        v.Name(),
		NameEn:      v.NameEn(),
		Prefecture:  v.Prefecture(),
		City:        v.City(),
		Address:     v.Address(),
		Capacity:    v.Capacity(),
		OfficialURL: v.OfficialURL(),
		CreatedAt:   v.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   v.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
