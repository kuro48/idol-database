package membership

import (
	"context"
	"fmt"

	domain "github.com/kuro48/idol-api/internal/domain/membership"
)

type Usecase struct {
	appService MembershipAppPort
}

func NewUsecase(appService MembershipAppPort) *Usecase {
	return &Usecase{appService: appService}
}

func (u *Usecase) CreateMembership(ctx context.Context, cmd CreateMembershipCommand) (*MembershipDTO, error) {
	m, err := u.appService.CreateMembership(ctx, MembershipCreateInput{
		IdolID:   cmd.IdolID,
		GroupID:  cmd.GroupID,
		Role:     cmd.Role,
		JoinedAt: cmd.JoinedAt,
	})
	if err != nil {
		return nil, err
	}
	dto := toDTO(m)
	return &dto, nil
}

func (u *Usecase) GetMembership(ctx context.Context, query GetMembershipQuery) (*MembershipDTO, error) {
	m, err := u.appService.GetMembership(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	dto := toDTO(m)
	return &dto, nil
}

func (u *Usecase) ListMemberships(ctx context.Context, query ListMembershipQuery) (*MembershipSearchResult, error) {
	query.Normalize()
	if err := query.Validate(); err != nil {
		return nil, err
	}

	var role *domain.Role
	if query.Role != nil {
		r, err := domain.NewRole(*query.Role)
		if err != nil {
			return nil, err
		}
		role = &r
	}

	criteria := domain.SearchCriteria{
		IdolID:   query.IdolID,
		GroupID:  query.GroupID,
		IsActive: query.IsActive,
		Role:     role,
		Sort:     *query.Sort,
		Order:    *query.Order,
		Offset:   (*query.Page - 1) * *query.Limit,
		Limit:    *query.Limit,
	}

	ms, err := u.appService.SearchMemberships(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("メンバーシップ一覧の取得エラー: %w", err)
	}

	total, err := u.appService.CountMemberships(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", err)
	}

	dtos := make([]*MembershipDTO, 0, len(ms))
	for _, m := range ms {
		dto := toDTO(m)
		dtos = append(dtos, &dto)
	}

	totalPages := int(total) / *query.Limit
	if int(total)%*query.Limit != 0 {
		totalPages++
	}

	return &MembershipSearchResult{
		Data: dtos,
		Meta: &PaginationMeta{
			Total:      total,
			Page:       *query.Page,
			PerPage:    *query.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (u *Usecase) ListByIdolID(ctx context.Context, idolID string) ([]*MembershipDTO, error) {
	ms, err := u.appService.ListByIdolID(ctx, idolID)
	if err != nil {
		return nil, err
	}
	return toDTOs(ms), nil
}

func (u *Usecase) ListByGroupID(ctx context.Context, groupID string) ([]*MembershipDTO, error) {
	ms, err := u.appService.ListByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return toDTOs(ms), nil
}

func (u *Usecase) UpdateMembership(ctx context.Context, cmd UpdateMembershipCommand) error {
	return u.appService.UpdateMembership(ctx, MembershipUpdateInput{
		ID:       cmd.ID,
		Role:     cmd.Role,
		JoinedAt: cmd.JoinedAt,
		LeftAt:   cmd.LeftAt,
	})
}

func (u *Usecase) DeleteMembership(ctx context.Context, cmd DeleteMembershipCommand) error {
	return u.appService.DeleteMembership(ctx, cmd.ID)
}

func toDTO(m *domain.Membership) MembershipDTO {
	dto := MembershipDTO{
		ID:        m.ID().Value(),
		IdolID:    m.IdolID(),
		GroupID:   m.GroupID(),
		Role:      m.Role().String(),
		IsActive:  m.IsActive(),
		CreatedAt: m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
	if m.JoinedAt() != nil {
		s := m.JoinedAt().Format("2006-01-02")
		dto.JoinedAt = &s
	}
	if m.LeftAt() != nil {
		s := m.LeftAt().Format("2006-01-02")
		dto.LeftAt = &s
	}
	return dto
}

func toDTOs(ms []*domain.Membership) []*MembershipDTO {
	dtos := make([]*MembershipDTO, 0, len(ms))
	for _, m := range ms {
		dto := toDTO(m)
		dtos = append(dtos, &dto)
	}
	return dtos
}
