package agency

import (
	"context"

	app "github.com/kuro48/idol-api/internal/application/agency"
	domain "github.com/kuro48/idol-api/internal/domain/agency"
)

// Usecase は事務所のユースケース
type Usecase struct {
	appService *app.ApplicationService
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService *app.ApplicationService) *Usecase {
	return &Usecase{appService: appService}
}

// CreateAgency は事務所を作成する
func (u *Usecase) CreateAgency(ctx context.Context, cmd CreateAgencyCommand) (*AgencyDTO, error) {
	entity, err := u.appService.CreateAgency(ctx, app.CreateInput{
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

// ListAgencies は事務所一覧を取得する
func (u *Usecase) ListAgencies(ctx context.Context, query ListAgenciesQuery) ([]*AgencyDTO, error) {
	agencies, err := u.appService.ListAgencies(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*AgencyDTO, 0, len(agencies))
	for _, a := range agencies {
		dto := toDTO(a)
		dtos = append(dtos, &dto)
	}

	return dtos, nil
}

// UpdateAgency は事務所を更新する
func (u *Usecase) UpdateAgency(ctx context.Context, cmd UpdateAgencyCommand) error {
	return u.appService.UpdateAgency(ctx, app.UpdateInput{
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
