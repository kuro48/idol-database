package idol

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kuro48/idol-api/internal/domain/agency"
	"github.com/kuro48/idol-api/internal/domain/idol"
)

// ApplicationService はアイドルアプリケーションサービス
type ApplicationService struct {
	repository       idol.Repository
	domainService    *idol.DomainService
	agencyRepository agency.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repository idol.Repository, agencyRepository agency.Repository) *ApplicationService {
	return &ApplicationService{
		repository:       repository,
		domainService:    idol.NewDomainService(repository),
		agencyRepository: agencyRepository,
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

	// 事務所IDの設定
	if cmd.AgencyID != nil {
		newIdol.UpdateAgency(cmd.AgencyID)
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

	dto := s.toDTO(foundIdol)

	// includeパラメータの処理
	if query.Include != nil {
		includes := strings.Split(*query.Include, ",")
		if err := s.loadIncludes(ctx, dto, includes); err != nil {
			return nil, fmt.Errorf("関連データの読み込みエラー: %w", err)
		}
	}

	return dto, nil
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

	if cmd.AgencyID != nil {
		existingIdol.UpdateAgency(cmd.AgencyID)
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

// SearchIdols は条件を指定してアイドルを検索する（並行処理版）
func (s *ApplicationService) SearchIdols(ctx context.Context, query ListIdolsQuery) (*SearchResult, error) {
	// SearchCriteriaに変換
	criteria := s.queryToCriteria(query)

	// 並行処理: データ取得と件数取得を同時実行
	var idols []*idol.Idol
	var total int64
	var errSearch, errCount error

	var wg sync.WaitGroup
	wg.Add(2)

	// データ取得
	go func() {
		defer wg.Done()
		idols, errSearch = s.repository.Search(ctx, criteria)
	}()

	// 総件数取得
	go func() {
		defer wg.Done()
		total, errCount = s.repository.Count(ctx, criteria)
	}()

	wg.Wait()

	// エラーチェック
	if errSearch != nil {
		return nil, fmt.Errorf("検索エラー: %w", errSearch)
	}
	if errCount != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", errCount)
	}

	// DTOに変換
	dtos := make([]*IdolDTO, 0, len(idols))
	for _, i := range idols {
		dtos = append(dtos, s.toDTO(i))
	}

	// includeパラメータの処理
	if query.Include != nil {
		includes := strings.Split(*query.Include, ",")
		for _, dto := range dtos {
			if err := s.loadIncludes(ctx, dto, includes); err != nil {
				return nil, fmt.Errorf("関連データの読み込みエラー: %w", err)
			}
		}
	}

	// ページネーション情報を計算
	meta := s.calculatePaginationMeta(total, *query.Page, *query.Limit)

	// ページネーションリンクを生成
	links := s.generatePaginationLinks(query, meta.TotalPages)

	return &SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: links,
	}, nil
}

// queryToCriteria はListIdolsQueryをSearchCriteriaに変換
func (s *ApplicationService) queryToCriteria(query ListIdolsQuery) idol.SearchCriteria {
	criteria := idol.SearchCriteria{
		Name:        query.Name,
		Nationality: query.Nationality,
		GroupID:     query.GroupID,
		AgencyID:    query.AgencyID,
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

// generatePaginationLinks はページネーションリンクを生成
func (s *ApplicationService) generatePaginationLinks(query ListIdolsQuery, totalPages int) *PaginationLinks {
	baseURL := "/api/v1/idols"

	// クエリパラメータを構築
	buildURL := func(page int) string {
		params := url.Values{}
		params.Set("page", strconv.Itoa(page))
		params.Set("limit", strconv.Itoa(*query.Limit))

		if query.Name != nil {
			params.Set("name", *query.Name)
		}
		if query.Nationality != nil {
			params.Set("nationality", *query.Nationality)
		}
		if query.GroupID != nil {
			params.Set("group_id", *query.GroupID)
		}
		if query.AgencyID != nil {
			params.Set("agency_id", *query.AgencyID)
		}
		if query.Include != nil {
			params.Set("include", *query.Include)
		}
		if query.AgeMin != nil {
			params.Set("age_min", strconv.Itoa(*query.AgeMin))
		}
		if query.AgeMax != nil {
			params.Set("age_max", strconv.Itoa(*query.AgeMax))
		}
		if query.BirthdateFrom != nil {
			params.Set("birthdate_from", *query.BirthdateFrom)
		}
		if query.BirthdateTo != nil {
			params.Set("birthdate_to", *query.BirthdateTo)
		}
		if query.Sort != nil {
			params.Set("sort", *query.Sort)
		}
		if query.Order != nil {
			params.Set("order", *query.Order)
		}

		return baseURL + "?" + params.Encode()
	}

	links := &PaginationLinks{
		First: buildURL(1),
		Last:  buildURL(totalPages),
	}

	// 次ページリンク
	if *query.Page < totalPages {
		next := buildURL(*query.Page + 1)
		links.Next = &next
	}

	// 前ページリンク
	if *query.Page > 1 {
		prev := buildURL(*query.Page - 1)
		links.Prev = &prev
	}

	return links
}

// loadIncludes は関連データを読み込んでDTOに展開する
func (s *ApplicationService) loadIncludes(ctx context.Context, dto *IdolDTO, includes []string) error {
	for _, include := range includes {
		switch strings.TrimSpace(include) {
		case "agency":
			if dto.AgencyID != nil {
				agencyID, err := agency.NewAgencyID(*dto.AgencyID)
				if err != nil {
					return fmt.Errorf("事務所IDの生成エラー: %w", err)
				}
				foundAgency, err := s.agencyRepository.FindByID(ctx, agencyID)
				if err != nil {
					// 事務所が見つからない場合はnilのまま（エラーにしない）
					continue
				}
				// Agencyを簡易DTOに変換して格納
				dto.Agency = map[string]interface{}{
					"id":               foundAgency.ID().Value(),
					"name":             foundAgency.Name().Value(),
					"name_en":          foundAgency.NameEn(),
					"country":          foundAgency.Country().Value(),
					"official_website": foundAgency.OfficialWebsite(),
					"logo_url":         foundAgency.LogoURL(),
				}
			}
		// 将来的に他のinclude対象（groups, events等）を追加可能
		}
	}
	return nil
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
		AgencyID:    i.AgencyID(),
		CreatedAt:   i.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   i.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
