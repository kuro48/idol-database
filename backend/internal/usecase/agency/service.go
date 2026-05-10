package agency

import (
	"context"
	"fmt"

	domain "github.com/kuro48/idol-api/internal/domain/agency"
)

// Usecase は事務所のユースケース
type Usecase struct {
	appService AgencyAppPort
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService AgencyAppPort) *Usecase {
	return &Usecase{appService: appService}
}

// CreateAgency は事務所を作成する
func (u *Usecase) CreateAgency(ctx context.Context, cmd CreateAgencyCommand) (*AgencyDTO, error) {
	entity, err := u.appService.CreateAgency(ctx, AgencyCreateInput{
		Name:            cmd.Name,
		NameEn:          cmd.NameEn,
		FoundedDate:     cmd.FoundedDate,
		Country:         cmd.Country,
		OfficialWebsite: cmd.OfficialWebsite,
		Description:     cmd.Description,
		LogoURL:         cmd.LogoURL,
	})
	if err != nil {
		return nil, err
	}

	dto := toDTO(entity)
	return &dto, nil
}

// GetAgency は事務所を取得する
func (u *Usecase) GetAgency(ctx context.Context, query GetAgencyQuery) (*AgencyDTO, error) {
	entity, err := u.appService.GetAgency(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	dto := toDTO(entity)
	return &dto, nil
}

// ListAgencies は事務所一覧を取得する（ページネーション付き）
func (u *Usecase) ListAgencies(ctx context.Context, query ListAgenciesQuery) (*AgencySearchResult, error) {
	query.Normalize()

	result, err := u.appService.ListAgenciesWithPagination(ctx, domain.SearchOptions{
		Name:    query.Name,
		Country: query.Country,
		Sort:    *query.Sort,
		Order:   *query.Order,
		Page:    *query.Page,
		Limit:   *query.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("事務所一覧の取得エラー: %w", err)
	}

	dtos := make([]*AgencyDTO, 0, len(result.Agencies))
	for _, a := range result.Agencies {
		dto := toDTO(a)
		dtos = append(dtos, &dto)
	}

	totalPages := int(result.Total) / *query.Limit
	if int(result.Total)%*query.Limit != 0 {
		totalPages++
	}

	return &AgencySearchResult{
		Data: dtos,
		Meta: PaginationMeta{
			Total:      result.Total,
			Page:       *query.Page,
			PerPage:    *query.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

// UpdateAgency は事務所を更新する
func (u *Usecase) UpdateAgency(ctx context.Context, cmd UpdateAgencyCommand) error {
	return u.appService.UpdateAgency(ctx, AgencyUpdateInput{
		ID:              cmd.ID,
		Name:            cmd.Name,
		NameEn:          cmd.NameEn,
		FoundedDate:     cmd.FoundedDate,
		OfficialWebsite: cmd.OfficialWebsite,
		Description:     cmd.Description,
		LogoURL:         cmd.LogoURL,
	})
}

// DeleteAgency は事務所を削除する
func (u *Usecase) DeleteAgency(ctx context.Context, cmd DeleteAgencyCommand) error {
	return u.appService.DeleteAgency(ctx, cmd.ID)
}

func toDTO(a *domain.Agency) AgencyDTO {
	var foundedDateStr *string
	if a.FoundedDate() != nil {
		str := a.FoundedDate().Format("2006-01-02")
		foundedDateStr = &str
	}

	return AgencyDTO{
		ID:              a.ID().Value(),
		Name:            a.Name().Value(),
		NameEn:          a.NameEn(),
		FoundedDate:     foundedDateStr,
		Country:         a.Country().Value(),
		OfficialWebsite: a.OfficialWebsite(),
		Description:     a.Description(),
		LogoURL:         a.LogoURL(),
		CreatedAt:       a.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       a.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
