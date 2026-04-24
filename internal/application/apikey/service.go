package apikey

import (
	"context"
	"fmt"

	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	"github.com/kuro48/idol-api/internal/domain/plan"
	"github.com/kuro48/idol-api/internal/shared/id"
)

// CreateKeyOutput はAPIキー作成の出力
// RawKey は生成直後の一度だけ返す値（DBには保存しない）
type CreateKeyOutput struct {
	RawKey string
	Key    *domainapikey.APIKey
}

// ApplicationService はAPIキー管理のアプリケーションサービス
type ApplicationService struct {
	repo domainapikey.Repository
}

// NewApplicationService はAPIキーアプリケーションサービスを作成する
func NewApplicationService(repo domainapikey.Repository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

// CreateKey は新しいAPIキーを作成する
// 生のキー文字列は出力に一度だけ含まれ、DBには保存されない
func (s *ApplicationService) CreateKey(ctx context.Context, input CreateKeyInput) (*CreateKeyOutput, error) {
	if !plan.IsValid(plan.Type(input.PlanType)) {
		return nil, fmt.Errorf("無効なプラン種別です: %s", input.PlanType)
	}

	rawKey, err := domainapikey.GenerateRawKey()
	if err != nil {
		return nil, fmt.Errorf("APIキーの生成に失敗しました: %w", err)
	}

	newID := id.Generate()
	key, err := domainapikey.New(newID, rawKey, input.Email, input.Name, plan.Type(input.PlanType))
	if err != nil {
		return nil, fmt.Errorf("APIキーエンティティの作成に失敗しました: %w", err)
	}

	if err := s.repo.Save(ctx, key); err != nil {
		return nil, fmt.Errorf("APIキーの保存に失敗しました: %w", err)
	}

	return &CreateKeyOutput{RawKey: rawKey, Key: key}, nil
}

// ListKeysByEmail はメールアドレスに紐づく全APIキーを返す
func (s *ApplicationService) ListKeysByEmail(ctx context.Context, email string) ([]*domainapikey.APIKey, error) {
	keys, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("APIキーの取得に失敗しました: %w", err)
	}
	return keys, nil
}

// RevokeKey はAPIキーを無効化する
func (s *ApplicationService) RevokeKey(ctx context.Context, input RevokeKeyInput) error {
	key, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("APIキーの取得に失敗しました: %w", err)
	}
	if key == nil {
		return fmt.Errorf("APIキーが見つかりません: %s", input.ID)
	}

	key.Deactivate()
	if err := s.repo.Update(ctx, key); err != nil {
		return fmt.Errorf("APIキーの無効化に失敗しました: %w", err)
	}
	return nil
}
