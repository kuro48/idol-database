package song

import (
	"context"
	"fmt"

	domain "github.com/kuro48/idol-api/internal/domain/song"
)

type Usecase struct {
	appService SongAppPort
}

func NewUsecase(appService SongAppPort) *Usecase {
	return &Usecase{appService: appService}
}

func (u *Usecase) CreateSong(ctx context.Context, cmd CreateSongCommand) (*SongDTO, error) {
	m, err := u.appService.CreateSong(ctx, SongCreateInput{
		Title:         cmd.Title,
		TitleKana:     cmd.TitleKana,
		DurationSec:   cmd.DurationSec,
		ISRC:          cmd.ISRC,
		CoverImageURL: cmd.CoverImageURL,
		Composers:     cmd.Composers,
		Lyricists:     cmd.Lyricists,
		Arrangers:     cmd.Arrangers,
	})
	if err != nil {
		return nil, err
	}
	dto := toDTO(m)
	return &dto, nil
}

func (u *Usecase) GetSong(ctx context.Context, query GetSongQuery) (*SongDTO, error) {
	m, err := u.appService.GetSong(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	dto := toDTO(m)
	return &dto, nil
}

func (u *Usecase) ListSongs(ctx context.Context, query ListSongQuery) (*SongSearchResult, error) {
	query.Normalize()
	if err := query.Validate(); err != nil {
		return nil, err
	}

	criteria := domain.SearchCriteria{
		Title:  query.Title,
		ISRC:   query.ISRC,
		Sort:   *query.Sort,
		Order:  *query.Order,
		Offset: (*query.Page - 1) * *query.Limit,
		Limit:  *query.Limit,
	}

	ms, err := u.appService.SearchSongs(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("楽曲一覧の取得エラー: %w", err)
	}

	total, err := u.appService.CountSongs(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", err)
	}

	dtos := make([]*SongDTO, 0, len(ms))
	for _, m := range ms {
		dto := toDTO(m)
		dtos = append(dtos, &dto)
	}

	totalPages := int(total) / *query.Limit
	if int(total)%*query.Limit != 0 {
		totalPages++
	}

	return &SongSearchResult{
		Data: dtos,
		Meta: &SongPagination{
			Total:      total,
			Page:       *query.Page,
			PerPage:    *query.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (u *Usecase) UpdateSong(ctx context.Context, cmd UpdateSongCommand) error {
	return u.appService.UpdateSong(ctx, SongUpdateInput{
		ID:            cmd.ID,
		Title:         cmd.Title,
		TitleKana:     cmd.TitleKana,
		DurationSec:   cmd.DurationSec,
		ISRC:          cmd.ISRC,
		CoverImageURL: cmd.CoverImageURL,
		Composers:     cmd.Composers,
		Lyricists:     cmd.Lyricists,
		Arrangers:     cmd.Arrangers,
	})
}

func (u *Usecase) DeleteSong(ctx context.Context, cmd DeleteSongCommand) error {
	return u.appService.DeleteSong(ctx, cmd.ID)
}

func toDTO(m *domain.Song) SongDTO {
	return SongDTO{
		ID:            m.ID().Value(),
		Title:         m.Title(),
		TitleKana:     m.TitleKana(),
		DurationSec:   m.DurationSec(),
		ISRC:          m.ISRC(),
		CoverImageURL: m.CoverImageURL(),
		Composers:     m.Composers(),
		Lyricists:     m.Lyricists(),
		Arrangers:     m.Arrangers(),
		CreatedAt:     m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
