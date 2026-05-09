package userprefs

import (
	"context"
	"fmt"

	domain "github.com/kuro48/idol-api/internal/domain/userprefs"
)

// Service はユーザー設定のアプリケーションサービス
type Service struct {
	repo domain.Repository
}

// New はサービスを作成する
func New(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// GetOrCreate は sub に対応するユーザー設定を返す。存在しない場合は新規作成する。
func (s *Service) GetOrCreate(ctx context.Context, sub string) (*domain.UserPrefs, error) {
	prefs, err := s.repo.FindBySub(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("ユーザー設定の取得に失敗しました: %w", err)
	}
	if prefs != nil {
		return prefs, nil
	}

	prefs, err = domain.New(sub)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Upsert(ctx, prefs); err != nil {
		return nil, fmt.Errorf("ユーザー設定の初期化に失敗しました: %w", err)
	}
	return prefs, nil
}

// UpdateOshiColor は推しメンカラーを更新して保存する
func (s *Service) UpdateOshiColor(ctx context.Context, sub, color string) (*domain.UserPrefs, error) {
	prefs, err := s.GetOrCreate(ctx, sub)
	if err != nil {
		return nil, err
	}

	if err := prefs.UpdateOshiColor(color); err != nil {
		return nil, err
	}

	if err := s.repo.Upsert(ctx, prefs); err != nil {
		return nil, fmt.Errorf("ユーザー設定の保存に失敗しました: %w", err)
	}
	return prefs, nil
}
