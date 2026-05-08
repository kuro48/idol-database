package release

import (
	"context"
	"fmt"
	"time"
)

// DomainService はリリースドメインのドメインサービス
type DomainService struct {
	repository Repository
}

func NewDomainService(repository Repository) *DomainService {
	return &DomainService{repository: repository}
}

// CanCreate は同一アーティスト・タイトル・日付の重複がないか検証する
func (s *DomainService) CanCreate(ctx context.Context, title ReleaseTitle, artists []ArtistRef, date ReleaseDate) error {
	from := date.Value()
	to := date.Value()
	criteria := SearchCriteria{
		Title:           strPtr(title.Value()),
		ReleaseDateFrom: &from,
		ReleaseDateTo:   &to,
		Limit:           100,
	}

	existing, err := s.repository.Search(ctx, criteria)
	if err != nil {
		return fmt.Errorf("重複チェックエラー: %w", err)
	}

	for _, r := range existing {
		if r.Title().Value() == title.Value() && sameArtists(r.Artists(), artists) {
			return NewDomainError("同じタイトル・アーティスト・日付のリリースが既に存在します")
		}
	}
	return nil
}

// sameArtists は2つのアーティスト参照リストが同一かを判定する
func sameArtists(a, b []ArtistRef) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, ref := range a {
		counts[string(ref.Kind())+":"+ref.ID()]++
	}
	for _, ref := range b {
		key := string(ref.Kind()) + ":" + ref.ID()
		counts[key]--
		if counts[key] < 0 {
			return false
		}
	}
	return true
}

func strPtr(s string) *string     { return &s }
func timePtr(t time.Time) *time.Time { return &t }
