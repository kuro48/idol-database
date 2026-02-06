package idol

import (
	"context"
	"fmt"
	"sync"

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
func (s *ApplicationService) CreateIdol(ctx context.Context, input CreateInput) (*idol.Idol, error) {
	// 値オブジェクトの生成
	name, err := idol.NewIdolName(input.Name)
	if err != nil {
		return nil, fmt.Errorf("名前の生成エラー: %w", err)
	}

	// ドメインサービスで重複チェック
	if err := s.domainService.CanCreate(ctx, name); err != nil {
		return nil, err
	}

	var birthdate *idol.Birthdate
	if input.Birthdate != nil {
		bd, err := idol.NewBirthdateFromString(*input.Birthdate)
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
	if input.AgencyID != nil {
		newIdol.UpdateAgency(input.AgencyID)
	}

	// 保存
	if err := s.repository.Save(ctx, newIdol); err != nil {
		return nil, fmt.Errorf("アイドルの保存エラー: %w", err)
	}

	return newIdol, nil
}

// GetIdol はアイドルを取得する
func (s *ApplicationService) GetIdol(ctx context.Context, id string) (*idol.Idol, error) {
	idolID, err := idol.NewIdolID(id)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundIdol, err := s.repository.FindByID(ctx, idolID)
	if err != nil {
		return nil, fmt.Errorf("アイドルの取得エラー: %w", err)
	}

	return foundIdol, nil
}

// ListIdols はアイドル一覧を取得する
func (s *ApplicationService) ListIdols(ctx context.Context) ([]*idol.Idol, error) {
	idols, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("アイドル一覧の取得エラー: %w", err)
	}

	return idols, nil
}

// UpdateIdol はアイドルを更新する
func (s *ApplicationService) UpdateIdol(ctx context.Context, input UpdateInput) error {
	id, err := idol.NewIdolID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingIdol, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("アイドルの取得エラー: %w", err)
	}

	// 各フィールドの更新
	if input.Name != nil {
		name, err := idol.NewIdolName(*input.Name)
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

	if input.Birthdate != nil {
		bd, err := idol.NewBirthdateFromString(*input.Birthdate)
		if err != nil {
			return fmt.Errorf("生年月日の生成エラー: %w", err)
		}
		existingIdol.UpdateBirthdate(&bd)
	}

	if input.AgencyID != nil {
		existingIdol.UpdateAgency(input.AgencyID)
	}

	// 更新の保存
	if err := s.repository.Update(ctx, existingIdol); err != nil {
		return fmt.Errorf("アイドルの更新エラー: %w", err)
	}

	return nil
}

// DeleteIdol はアイドルを削除する
func (s *ApplicationService) DeleteIdol(ctx context.Context, id string) error {
	idolID, err := idol.NewIdolID(id)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, idolID); err != nil {
		return fmt.Errorf("アイドルの削除エラー: %w", err)
	}

	return nil
}

// UpdateSocialLinks はSNS/外部リンクを更新する
func (s *ApplicationService) UpdateSocialLinks(ctx context.Context, input UpdateSocialLinksInput) error {
	id, err := idol.NewIdolID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingIdol, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("アイドルの取得エラー: %w", err)
	}

	// SocialLinksの作成と設定
	links := idol.NewSocialLinks()

	if input.Twitter != nil && *input.Twitter != "" {
		if err := links.SetTwitter(*input.Twitter); err != nil {
			return fmt.Errorf("Twitter URLエラー: %w", err)
		}
	}

	if input.Instagram != nil && *input.Instagram != "" {
		if err := links.SetInstagram(*input.Instagram); err != nil {
			return fmt.Errorf("Instagram URLエラー: %w", err)
		}
	}

	if input.TikTok != nil && *input.TikTok != "" {
		if err := links.SetTikTok(*input.TikTok); err != nil {
			return fmt.Errorf("TikTok URLエラー: %w", err)
		}
	}

	if input.YouTube != nil && *input.YouTube != "" {
		if err := links.SetYouTube(*input.YouTube); err != nil {
			return fmt.Errorf("YouTube URLエラー: %w", err)
		}
	}

	if input.Facebook != nil && *input.Facebook != "" {
		if err := links.SetFacebook(*input.Facebook); err != nil {
			return fmt.Errorf("Facebook URLエラー: %w", err)
		}
	}

	if input.OfficialWebsite != nil && *input.OfficialWebsite != "" {
		if err := links.SetOfficial(*input.OfficialWebsite); err != nil {
			return fmt.Errorf("公式サイトURLエラー: %w", err)
		}
	}

	if input.FanClub != nil && *input.FanClub != "" {
		if err := links.SetFanClub(*input.FanClub); err != nil {
			return fmt.Errorf("ファンクラブURLエラー: %w", err)
		}
	}

	existingIdol.UpdateSocialLinks(links)

	if err := s.repository.Update(ctx, existingIdol); err != nil {
		return fmt.Errorf("アイドルの更新エラー: %w", err)
	}

	return nil
}

// SearchIdols は条件を指定してアイドルを検索する（並行処理版）
func (s *ApplicationService) SearchIdols(ctx context.Context, criteria idol.SearchCriteria) ([]*idol.Idol, int64, error) {
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
		return nil, 0, fmt.Errorf("検索エラー: %w", errSearch)
	}
	if errCount != nil {
		return nil, 0, fmt.Errorf("件数取得エラー: %w", errCount)
	}

	return idols, total, nil
}
