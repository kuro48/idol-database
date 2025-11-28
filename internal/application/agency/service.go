package agency

import (
	"context"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/agency"
)

// ApplicationService は事務所アプリケーションサービス
type ApplicationService struct {
	repository    agency.Repository
	domainService *agency.DomainService
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repository agency.Repository) *ApplicationService {
	return &ApplicationService{
		repository:    repository,
		domainService: agency.NewDomainService(repository),
	}
}

// CreateAgency は事務所を作成する
func (s *ApplicationService) CreateAgency(ctx context.Context, cmd CreateAgencyCommand) (*AgencyDTO, error) {
	// 値オブジェクトの生成
	name, err := agency.NewAgencyName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("名前の生成エラー: %w", err)
	}

	country, err := agency.NewCountry(cmd.Country)
	if err != nil {
		return nil, fmt.Errorf("国の生成エラー: %w", err)
	}

	// ドメインサービスで重複チェック
	if err := s.domainService.CanCreate(ctx, name); err != nil {
		return nil, err
	}

	// IDを生成（MongoDBのObjectIDを生成）
	id, err := agency.NewAgencyID(generateID())
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	// エンティティの生成
	newAgency := agency.NewAgency(id, name, country)

	// オプションフィールドの設定
	if cmd.FoundedDate != nil {
		foundedDate, err := time.Parse("2006-01-02", *cmd.FoundedDate)
		if err == nil {
			newAgency.UpdateDetails(nil, cmd.NameEn, &foundedDate, cmd.OfficialWebsite, cmd.Description, cmd.LogoURL)
		}
	} else {
		newAgency.UpdateDetails(nil, cmd.NameEn, nil, cmd.OfficialWebsite, cmd.Description, cmd.LogoURL)
	}

	// 保存
	if err := s.repository.Save(ctx, newAgency); err != nil {
		return nil, fmt.Errorf("事務所の保存エラー: %w", err)
	}

	return s.toDTO(newAgency), nil
}

// GetAgency は事務所を取得する
func (s *ApplicationService) GetAgency(ctx context.Context, query GetAgencyQuery) (*AgencyDTO, error) {
	id, err := agency.NewAgencyID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundAgency, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("事務所の取得エラー: %w", err)
	}

	return s.toDTO(foundAgency), nil
}

// ListAgencies は事務所一覧を取得する
func (s *ApplicationService) ListAgencies(ctx context.Context, query ListAgenciesQuery) ([]*AgencyDTO, error) {
	agencies, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("事務所一覧の取得エラー: %w", err)
	}

	dtos := make([]*AgencyDTO, 0, len(agencies))
	for _, a := range agencies {
		dtos = append(dtos, s.toDTO(a))
	}

	return dtos, nil
}

// UpdateAgency は事務所を更新する
func (s *ApplicationService) UpdateAgency(ctx context.Context, cmd UpdateAgencyCommand) error {
	id, err := agency.NewAgencyID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingAgency, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("事務所の取得エラー: %w", err)
	}

	// 名前の更新と重複チェック
	var newName *agency.AgencyName
	if cmd.Name != nil {
		name, err := agency.NewAgencyName(*cmd.Name)
		if err != nil {
			return fmt.Errorf("名前の生成エラー: %w", err)
		}

		// 名前の重複チェック（自分自身は除外）
		isDuplicate, err := s.domainService.IsDuplicateName(ctx, name, &id)
		if err != nil {
			return err
		}
		if isDuplicate {
			return fmt.Errorf("同じ名前の事務所が既に存在します")
		}

		newName = &name
	}

	// 設立日のパース
	var foundedDate *time.Time
	if cmd.FoundedDate != nil {
		parsed, err := time.Parse("2006-01-02", *cmd.FoundedDate)
		if err == nil {
			foundedDate = &parsed
		}
	}

	// 更新
	existingAgency.UpdateDetails(newName, cmd.NameEn, foundedDate, cmd.OfficialWebsite, cmd.Description, cmd.LogoURL)

	// 保存
	if err := s.repository.Update(ctx, existingAgency); err != nil {
		return fmt.Errorf("事務所の更新エラー: %w", err)
	}

	return nil
}

// DeleteAgency は事務所を削除する
func (s *ApplicationService) DeleteAgency(ctx context.Context, cmd DeleteAgencyCommand) error {
	id, err := agency.NewAgencyID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("事務所の削除エラー: %w", err)
	}

	return nil
}

// toDTO はドメインモデルをDTOに変換する
func (s *ApplicationService) toDTO(a *agency.Agency) *AgencyDTO {
	var foundedDateStr *string
	if a.FoundedDate() != nil {
		str := a.FoundedDate().Format("2006-01-02")
		foundedDateStr = &str
	}

	return &AgencyDTO{
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

// generateID はIDを生成する（簡易実装）
func generateID() string {
	// 実際にはMongoDBのObjectIDを生成する
	// ここでは仮実装
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
