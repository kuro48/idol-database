package song

import (
	"errors"
	"time"

	"github.com/kuro48/idol-api/internal/shared/source"
)

// Song は楽曲を表すドメインモデル
type Song struct {
	id            SongID
	title         string
	titleKana     *string
	durationSec   *int
	isrc          *string
	coverImageURL *string
	composers     []string
	lyricists     []string
	arrangers     []string
	sources       []source.Source
	createdAt     time.Time
	updatedAt     time.Time
}

func NewSong(title string, titleKana *string, durationSec *int, isrc *string, coverImageURL *string, composers, lyricists, arrangers []string) (*Song, error) {
	if title == "" {
		return nil, errors.New("楽曲タイトルは必須です")
	}
	if len([]rune(title)) > 200 {
		return nil, errors.New("楽曲タイトルは200文字以内である必要があります")
	}
	if durationSec != nil && *durationSec < 0 {
		return nil, errors.New("再生時間は0以上である必要があります")
	}
	now := time.Now()
	return &Song{
		title:         title,
		titleKana:     titleKana,
		durationSec:   durationSec,
		isrc:          isrc,
		coverImageURL: coverImageURL,
		composers:     composers,
		lyricists:     lyricists,
		arrangers:     arrangers,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func Reconstruct(
	id SongID,
	title string,
	titleKana *string,
	durationSec *int,
	isrc *string,
	coverImageURL *string,
	composers, lyricists, arrangers []string,
	sources []source.Source,
	createdAt, updatedAt time.Time,
) *Song {
	return &Song{
		id:            id,
		title:         title,
		titleKana:     titleKana,
		durationSec:   durationSec,
		isrc:          isrc,
		coverImageURL: coverImageURL,
		composers:     composers,
		lyricists:     lyricists,
		arrangers:     arrangers,
		sources:       sources,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (s *Song) ID() SongID             { return s.id }
func (s *Song) Title() string          { return s.title }
func (s *Song) TitleKana() *string     { return s.titleKana }
func (s *Song) DurationSec() *int      { return s.durationSec }
func (s *Song) ISRC() *string          { return s.isrc }
func (s *Song) CoverImageURL() *string { return s.coverImageURL }
func (s *Song) CreatedAt() time.Time   { return s.createdAt }
func (s *Song) UpdatedAt() time.Time   { return s.updatedAt }

func (s *Song) Composers() []string {
	if s.composers == nil {
		return []string{}
	}
	return s.composers
}

func (s *Song) Lyricists() []string {
	if s.lyricists == nil {
		return []string{}
	}
	return s.lyricists
}

func (s *Song) Arrangers() []string {
	if s.arrangers == nil {
		return []string{}
	}
	return s.arrangers
}

func (s *Song) Sources() []source.Source {
	if s.sources == nil {
		return []source.Source{}
	}
	return s.sources
}

func (s *Song) SetID(id SongID) {
	s.id = id
}

func (s *Song) Update(title string, titleKana *string, durationSec *int, isrc *string, coverImageURL *string, composers, lyricists, arrangers []string) error {
	if title == "" {
		return errors.New("楽曲タイトルは必須です")
	}
	if len([]rune(title)) > 200 {
		return errors.New("楽曲タイトルは200文字以内である必要があります")
	}
	if durationSec != nil && *durationSec < 0 {
		return errors.New("再生時間は0以上である必要があります")
	}
	s.title = title
	s.titleKana = titleKana
	s.durationSec = durationSec
	s.isrc = isrc
	s.coverImageURL = coverImageURL
	s.composers = composers
	s.lyricists = lyricists
	s.arrangers = arrangers
	s.updatedAt = time.Now()
	return nil
}

func (s *Song) SetSources(sources []source.Source) {
	s.sources = sources
	s.updatedAt = time.Now()
}
