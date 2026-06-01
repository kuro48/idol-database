package song

import (
	"context"
	"fmt"

	"github.com/kuro48/idol-api/internal/domain/song"
	"github.com/kuro48/idol-api/internal/shared/audit"
)

type ApplicationService struct {
	repository song.Repository
}

func NewApplicationService(repo song.Repository) *ApplicationService {
	return &ApplicationService{repository: repo}
}

func (s *ApplicationService) CreateSong(ctx context.Context, input CreateInput) (*song.Song, error) {
	m, err := song.NewSong(
		input.Title,
		input.TitleKana,
		input.DurationSec,
		input.ISRC,
		input.CoverImageURL,
		input.Composers,
		input.Lyricists,
		input.Arrangers,
	)
	if err != nil {
		return nil, err
	}

	_ = audit.ActorFrom(ctx)

	if err := s.repository.Save(ctx, m); err != nil {
		return nil, fmt.Errorf("楽曲の保存エラー: %w", err)
	}

	return m, nil
}

func (s *ApplicationService) GetSong(ctx context.Context, id string) (*song.Song, error) {
	sid, err := song.NewSongID(id)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	m, err := s.repository.FindByID(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("楽曲の取得エラー: %w", err)
	}

	return m, nil
}

func (s *ApplicationService) SearchSongs(ctx context.Context, criteria song.SearchCriteria) ([]*song.Song, error) {
	ms, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("楽曲検索エラー: %w", err)
	}
	return ms, nil
}

func (s *ApplicationService) CountSongs(ctx context.Context, criteria song.SearchCriteria) (int64, error) {
	return s.repository.Count(ctx, criteria)
}

func (s *ApplicationService) UpdateSong(ctx context.Context, input UpdateInput) error {
	sid, err := song.NewSongID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	m, err := s.repository.FindByID(ctx, sid)
	if err != nil {
		return fmt.Errorf("楽曲の取得エラー: %w", err)
	}

	if err := m.Update(
		input.Title,
		input.TitleKana,
		input.DurationSec,
		input.ISRC,
		input.CoverImageURL,
		input.Composers,
		input.Lyricists,
		input.Arrangers,
	); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, m); err != nil {
		return fmt.Errorf("楽曲の更新エラー: %w", err)
	}

	return nil
}

func (s *ApplicationService) DeleteSong(ctx context.Context, id string) error {
	sid, err := song.NewSongID(id)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, sid); err != nil {
		return fmt.Errorf("楽曲の削除エラー: %w", err)
	}

	return nil
}
