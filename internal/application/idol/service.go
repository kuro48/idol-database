package idol

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kuro48/idol-api/internal/domain/idol"
)

// ApplicationService はアイドルアプリケーションサービス
type ApplicationService struct {
	repository    idol.Repository
	domainService *idol.DomainService
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repository idol.Repository) *ApplicationService {
	return &ApplicationService{
		repository:    repository,
		domainService: idol.NewDomainService(repository),
	}
}

// CreateIdol はアイドルを作成する
func (s *ApplicationService) CreateIdol(ctx context.Context, cmd CreateIdolCommand) (*IdolDTO, error) {
	// 値オブジェクトの生成
	name, err := idol.NewIdolName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("名前の生成エラー: %w", err)
	}

	// ドメインサービスで重複チェック
	if err := s.domainService.CanCreate(ctx, name); err != nil {
		return nil, err
	}

	var birthdate *idol.Birthdate
	if cmd.Birthdate != nil {
		bd, err := idol.NewBirthdateFromString(*cmd.Birthdate)
		if err != nil {
			return nil, fmt.Errorf("生年月日の生成エラー: %w", err)
		}
		birthdate = &bd
	}

	// エンティティの生成
	newIdol, err := idol.NewIdol(name, birthdate)
	if err != nil {
		return nil, fmt.Errorf("アイドルの生成エラー: %w", err)
	}

	// 保存
	if err := s.repository.Save(ctx, newIdol); err != nil {
		return nil, fmt.Errorf("アイドルの保存エラー: %w", err)
	}

	return s.toDTO(newIdol), nil
}

// GetIdol はアイドルを取得する
func (s *ApplicationService) GetIdol(ctx context.Context, query GetIdolQuery) (*IdolDTO, error) {
	id, err := idol.NewIdolID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundIdol, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("アイドルの取得エラー: %w", err)
	}

	return s.toDTO(foundIdol), nil
}

// ListIdols はアイドル一覧を取得する
func (s *ApplicationService) ListIdols(ctx context.Context, query ListIdolsQuery) ([]*IdolDTO, error) {
	idols, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("アイドル一覧の取得エラー: %w", err)
	}

	dtos := make([]*IdolDTO, 0, len(idols))
	for _, i := range idols {
		dtos = append(dtos, s.toDTO(i))
	}

	return dtos, nil
}

// UpdateIdol はアイドルを更新する
func (s *ApplicationService) UpdateIdol(ctx context.Context, cmd UpdateIdolCommand) error {
	id, err := idol.NewIdolID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingIdol, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("アイドルの取得エラー: %w", err)
	}

	// 各フィールドの更新
	if cmd.Name != nil {
		name, err := idol.NewIdolName(*cmd.Name)
		if err != nil {
			return fmt.Errorf("名前の生成エラー: %w", err)
		}

		// 名前の重複チェック（自分自身は除外）
		isDuplicate, err := s.domainService.IsDuplicateName(ctx, name, &id)
		if err != nil {
			return err
		}
		if isDuplicate {
			return fmt.Errorf("同じ名前のアイドルが既に存在します")
		}

		if err := existingIdol.ChangeName(name); err != nil {
			return err
		}
	}

	if cmd.Birthdate != nil {
		bd, err := idol.NewBirthdateFromString(*cmd.Birthdate)
		if err != nil {
			return fmt.Errorf("生年月日の生成エラー: %w", err)
		}
		existingIdol.UpdateBirthdate(&bd)
	}

	// 更新の保存
	if err := s.repository.Update(ctx, existingIdol); err != nil {
		return fmt.Errorf("アイドルの更新エラー: %w", err)
	}

	return nil
}

// DeleteIdol はアイドルを削除する
func (s *ApplicationService) DeleteIdol(ctx context.Context, cmd DeleteIdolCommand) error {
	id, err := idol.NewIdolID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("アイドルの削除エラー: %w", err)
	}

	return nil
}

// SearchIdols は条件を指定してアイドルを検索する
func (s *ApplicationService) SearchIdols(ctx context.Context, query ListIdolsQuery) (*SearchResult, error) {
	// SearchCriteriaに変換
	criteria := s.queryToCriteria(query)

	// 総件数を取得
	total, err := s.repository.Count(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", err)
	}

	// 検索実行
	idols, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("検索エラー: %w", err)
	}

	// DTOに変換
	dtos := make([]*IdolDTO, 0, len(idols))
	for _, i := range idols {
		dtos = append(dtos, s.toDTO(i))
	}

	// ページネーション情報を計算
	meta := s.calculatePaginationMeta(total, *query.Page, *query.Limit)

	return &SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: nil, // リンクは後で実装
	}, nil
}

// queryToCriteria はListIdolsQueryをSearchCriteriaに変換
func (s *ApplicationService) queryToCriteria(query ListIdolsQuery) idol.SearchCriteria {
	criteria := idol.SearchCriteria{
		Name:        query.Name,
		Nationality: query.Nationality,
		GroupID:     query.GroupID,
		AgeMin:      query.AgeMin,
		AgeMax:      query.AgeMax,
		Sort:        *query.Sort,
		Order:       *query.Order,
		Offset:      (*query.Page - 1) * *query.Limit,
		Limit:       *query.Limit,
	}

	// 生年月日範囲の変換（YYYY-MM-DDからtime.Timeへ）
	if query.BirthdateFrom != nil {
		if t, err := time.Parse("2006-01-02", *query.BirthdateFrom); err == nil {
			criteria.BirthdateFrom = &t
		}
	}
	if query.BirthdateTo != nil {
		if t, err := time.Parse("2006-01-02", *query.BirthdateTo); err == nil {
			criteria.BirthdateTo = &t
		}
	}

	return criteria
}

// calculatePaginationMeta はページネーション情報を計算
func (s *ApplicationService) calculatePaginationMeta(total int64, page, perPage int) *PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationMeta{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// toDTO はドメインモデルをDTOに変換する
func (s *ApplicationService) toDTO(i *idol.Idol) *IdolDTO {
	var birthdateStr string
	if i.Birthdate() != nil {
		birthdateStr = i.Birthdate().String()
	}

	var age *int
	if ageValue, err := i.Age(); err == nil {
		age = &ageValue
	}

	return &IdolDTO{
		ID:          i.ID().Value(),
		Name:        i.Name().Value(),
		Birthdate:   birthdateStr,
		Age:         age,
		CreatedAt:   i.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   i.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
