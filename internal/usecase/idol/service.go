package idol

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	domain "github.com/kuro48/idol-api/internal/domain/idol"
)

// Usecase はアイドルのユースケース
type Usecase struct {
	appService *appIdol.ApplicationService
	agencyApp  *appAgency.ApplicationService
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService *appIdol.ApplicationService, agencyApp *appAgency.ApplicationService) *Usecase {
	return &Usecase{appService: appService, agencyApp: agencyApp}
}

// CreateIdol はアイドルを作成する
func (u *Usecase) CreateIdol(ctx context.Context, cmd CreateIdolCommand) (*IdolDTO, error) {
	// AgencyIDが指定された場合は存在確認
	if cmd.AgencyID != nil {
		if _, err := u.agencyApp.GetAgency(ctx, *cmd.AgencyID); err != nil {
			return nil, fmt.Errorf("指定された事務所が見つかりません: %w", err)
		}
	}

	entity, err := u.appService.CreateIdol(ctx, appIdol.CreateInput{
		Name:      cmd.Name,
		Birthdate: cmd.Birthdate,
		AgencyID:  cmd.AgencyID,
	})
	if err != nil {
		return nil, err
	}

	dto := u.toDTO(entity)
	return dto, nil
}

// GetIdol はアイドルを取得する
func (u *Usecase) GetIdol(ctx context.Context, query GetIdolQuery) (*IdolDTO, error) {
	entity, err := u.appService.GetIdol(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	dto := u.toDTO(entity)

	// includeパラメータの処理
	if query.Include != nil {
		includes := strings.Split(*query.Include, ",")
		if err := u.loadIncludes(ctx, dto, includes); err != nil {
			return nil, fmt.Errorf("関連データの読み込みエラー: %w", err)
		}
	}

	return dto, nil
}

// ListIdols はアイドル一覧を取得する
func (u *Usecase) ListIdols(ctx context.Context, query ListIdolsQuery) ([]*IdolDTO, error) {
	idols, err := u.appService.ListIdols(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*IdolDTO, 0, len(idols))
	for _, i := range idols {
		dtos = append(dtos, u.toDTO(i))
	}

	return dtos, nil
}

// UpdateIdol はアイドルを更新する
func (u *Usecase) UpdateIdol(ctx context.Context, cmd UpdateIdolCommand) error {
	// AgencyIDが指定された場合は存在確認
	if cmd.AgencyID != nil {
		if _, err := u.agencyApp.GetAgency(ctx, *cmd.AgencyID); err != nil {
			return fmt.Errorf("指定された事務所が見つかりません: %w", err)
		}
	}

	return u.appService.UpdateIdol(ctx, appIdol.UpdateInput{
		ID:        cmd.ID,
		Name:      cmd.Name,
		Birthdate: cmd.Birthdate,
		AgencyID:  cmd.AgencyID,
	})
}

// DeleteIdol はアイドルを削除する
func (u *Usecase) DeleteIdol(ctx context.Context, cmd DeleteIdolCommand) error {
	return u.appService.DeleteIdol(ctx, cmd.ID)
}

// RestoreIdol はソフトデリートされたアイドルを復元する
func (u *Usecase) RestoreIdol(ctx context.Context, id string) error {
	return u.appService.RestoreIdol(ctx, id)
}

// UpdateSocialLinks はSNS/外部リンクを更新する
func (u *Usecase) UpdateSocialLinks(ctx context.Context, cmd UpdateSocialLinksCommand) error {
	return u.appService.UpdateSocialLinks(ctx, appIdol.UpdateSocialLinksInput{
		ID:              cmd.ID,
		Twitter:         cmd.Twitter,
		Instagram:       cmd.Instagram,
		TikTok:          cmd.TikTok,
		YouTube:         cmd.YouTube,
		Facebook:        cmd.Facebook,
		OfficialWebsite: cmd.OfficialWebsite,
		FanClub:         cmd.FanClub,
	})
}

// SearchIdols は条件を指定してアイドルを検索する
func (u *Usecase) SearchIdols(ctx context.Context, query ListIdolsQuery) (*SearchResult, error) {
	criteria := u.queryToCriteria(query)

	idols, total, err := u.appService.SearchIdols(ctx, criteria)
	if err != nil {
		return nil, err
	}

	dtos := make([]*IdolDTO, 0, len(idols))
	for _, i := range idols {
		dtos = append(dtos, u.toDTO(i))
	}

	// includeパラメータの処理
	if query.Include != nil {
		includes := strings.Split(*query.Include, ",")
		for _, dto := range dtos {
			if err := u.loadIncludes(ctx, dto, includes); err != nil {
				return nil, fmt.Errorf("関連データの読み込みエラー: %w", err)
			}
		}
	}

	// ページネーション情報を計算
	meta := u.calculatePaginationMeta(total, *query.Page, *query.Limit)

	// ページネーションリンクを生成
	links := u.generatePaginationLinks(query, meta.TotalPages)

	return &SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: links,
	}, nil
}

// queryToCriteria はListIdolsQueryをSearchCriteriaに変換
func (u *Usecase) queryToCriteria(query ListIdolsQuery) domain.SearchCriteria {
	criteria := domain.SearchCriteria{
		Name:     query.Name,
		AgencyID: query.AgencyID,
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
func (u *Usecase) calculatePaginationMeta(total int64, page, perPage int) *PaginationMeta {
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
func (u *Usecase) generatePaginationLinks(query ListIdolsQuery, totalPages int) *PaginationLinks {
	baseURL := "/api/v1/idols"

	// クエリパラメータを構築
	buildURL := func(page int) string {
		params := url.Values{}
		params.Set("page", strconv.Itoa(page))
		params.Set("limit", strconv.Itoa(*query.Limit))

		if query.Name != nil {
			params.Set("name", *query.Name)
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
func (u *Usecase) loadIncludes(ctx context.Context, dto *IdolDTO, includes []string) error {
	for _, include := range includes {
		switch strings.TrimSpace(include) {
		case "agency":
			if dto.AgencyID != nil {
				foundAgency, err := u.agencyApp.GetAgency(ctx, *dto.AgencyID)
				if err != nil {
					// 事務所が見つからない場合はnilのまま（エラーにしない）
					continue
				}
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
func (u *Usecase) toDTO(i *domain.Idol) *IdolDTO {
	var birthdateStr string
	if i.Birthdate() != nil {
		birthdateStr = i.Birthdate().String()
	}

	var age *int
	if ageValue, err := i.Age(); err == nil {
		age = &ageValue
	}

	var socialLinksMap interface{}
	if i.SocialLinks() != nil {
		socialLinksMap = map[string]interface{}{
			"twitter":          i.SocialLinks().Twitter(),
			"instagram":        i.SocialLinks().Instagram(),
			"tiktok":           i.SocialLinks().TikTok(),
			"youtube":          i.SocialLinks().YouTube(),
			"facebook":         i.SocialLinks().Facebook(),
			"official_website": i.SocialLinks().Official(),
			"fan_club":         i.SocialLinks().FanClub(),
		}
	}

	// ExternalIDs の変換
	var externalIDsMap map[string]string
	if extIDs := i.ExternalIDs(); !extIDs.IsEmpty() {
		rawIDs := extIDs.All()
		externalIDsMap = make(map[string]string, len(rawIDs))
		for k, v := range rawIDs {
			externalIDsMap[string(k)] = v
		}
	}

	return &IdolDTO{
		ID:          i.ID().Value(),
		Name:        i.Name().Value(),
		Birthdate:   birthdateStr,
		Age:         age,
		AgencyID:    i.AgencyID(),
		SocialLinks: socialLinksMap,
		ExternalIDs: externalIDsMap,
		CreatedAt:   i.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   i.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// DuplicateCandidateDTO は重複候補のDTO
type DuplicateCandidateDTO struct {
	Idol   *IdolDTO `json:"idol"`
	Reason string   `json:"reason"`
	Score  int      `json:"score"`
}

// FindDuplicateCandidates は指定したアイドルIDの重複候補を返す
func (u *Usecase) FindDuplicateCandidates(ctx context.Context, id string) ([]*DuplicateCandidateDTO, error) {
	candidates, err := u.appService.FindDuplicateCandidates(ctx, id)
	if err != nil {
		return nil, err
	}

	dtos := make([]*DuplicateCandidateDTO, 0, len(candidates))
	for _, c := range candidates {
		dtos = append(dtos, &DuplicateCandidateDTO{
			Idol:   u.toDTO(c.Idol),
			Reason: c.Reason,
			Score:  c.Score,
		})
	}

	return dtos, nil
}

// GetExternalIDs は外部IDマッピングを取得する
func (u *Usecase) GetExternalIDs(ctx context.Context, id string) (map[string]string, error) {
	entity, err := u.appService.GetIdol(ctx, id)
	if err != nil {
		return nil, err
	}
	extIDs := entity.ExternalIDs()
	if extIDs.IsEmpty() {
		return map[string]string{}, nil
	}
	rawIDs := extIDs.All()
	result := make(map[string]string, len(rawIDs))
	for k, v := range rawIDs {
		result[string(k)] = v
	}
	return result, nil
}

// UpdateExternalIDs は外部IDマッピングを更新する
func (u *Usecase) UpdateExternalIDs(ctx context.Context, cmd UpdateExternalIDsCommand) error {
	return u.appService.UpdateExternalIDs(ctx, appIdol.UpdateExternalIDsInput{
		ID:          cmd.ID,
		ExternalIDs: cmd.ExternalIDs,
	})
}

// BulkResult はバルク処理の結果
type BulkResult struct {
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	Errors       []BulkError   `json:"errors,omitempty"`
	Created      []*IdolDTO    `json:"created,omitempty"`
}

// BulkError はバルク処理の個別エラー
type BulkError struct {
	Index   int    `json:"index"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message"`
}

// MaxBulkCreateSize はバルク作成の最大バッチサイズ
const MaxBulkCreateSize = 100

// BulkCreateIdols は複数のアイドルを一括作成する
// エラーが発生しても他のアイドルの処理を続ける（partial success）
func (u *Usecase) BulkCreateIdols(ctx context.Context, cmds []CreateIdolCommand) (*BulkResult, error) {
	if len(cmds) > MaxBulkCreateSize {
		return nil, fmt.Errorf("バッチサイズ %d は上限 %d を超えています", len(cmds), MaxBulkCreateSize)
	}

	result := &BulkResult{
		Errors:  make([]BulkError, 0),
		Created: make([]*IdolDTO, 0),
	}

	for i, cmd := range cmds {
		dto, err := u.CreateIdol(ctx, cmd)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, BulkError{
				Index:   i,
				Name:    cmd.Name,
				Message: err.Error(),
			})
			continue
		}
		result.SuccessCount++
		result.Created = append(result.Created, dto)
	}

	return result, nil
}
