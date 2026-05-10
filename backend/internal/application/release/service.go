package release

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/kuro48/idol-api/internal/domain/release"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
)

// WebhookPublisher はリリース変更イベントを通知する契約
type WebhookPublisher interface {
	Publish(ctx context.Context, event domainWebhook.EventType, payload interface{}) error
}

// ApplicationService はリリースアプリケーションサービス
type ApplicationService struct {
	repository    release.Repository
	domainService *release.DomainService
	publisher     WebhookPublisher
}

func NewApplicationService(repository release.Repository, publisher WebhookPublisher) *ApplicationService {
	return &ApplicationService{
		repository:    repository,
		domainService: release.NewDomainService(repository),
		publisher:     publisher,
	}
}

// CreateRelease はリリースを作成する
func (s *ApplicationService) CreateRelease(ctx context.Context, input CreateInput) (*release.Release, error) {
	title, err := release.NewReleaseTitle(input.Title)
	if err != nil {
		return nil, fmt.Errorf("タイトルエラー: %w", err)
	}

	releaseType, err := release.NewReleaseType(input.ReleaseType)
	if err != nil {
		return nil, fmt.Errorf("リリース種別エラー: %w", err)
	}

	releaseDate, err := release.NewReleaseDateFromString(input.ReleaseDate)
	if err != nil {
		return nil, fmt.Errorf("リリース日エラー: %w", err)
	}

	artists, err := buildArtistRefs(input.Artists)
	if err != nil {
		return nil, err
	}

	if err := s.domainService.CanCreate(ctx, title, artists, releaseDate); err != nil {
		return nil, err
	}

	r, err := release.NewRelease(title, releaseType, releaseDate, artists)
	if err != nil {
		return nil, fmt.Errorf("リリース生成エラー: %w", err)
	}

	if len(input.Tracks) > 0 {
		tracks, err := buildTracks(input.Tracks)
		if err != nil {
			return nil, err
		}
		if err := r.SetTracks(tracks); err != nil {
			return nil, fmt.Errorf("収録曲エラー: %w", err)
		}
	}

	if input.StreamingLinks != nil {
		links, err := buildStreamingLinks(input.StreamingLinks)
		if err != nil {
			return nil, err
		}
		r.UpdateStreamingLinks(links)
	}

	if input.CoverImageURL != nil {
		r.SetCoverImageURL(input.CoverImageURL)
	}
	if len(input.Aliases) > 0 {
		r.SetAliases(input.Aliases)
	}
	if len(input.TagIDs) > 0 {
		r.SetTags(input.TagIDs)
	}

	if err := s.repository.Save(ctx, r); err != nil {
		return nil, fmt.Errorf("リリースの保存エラー: %w", err)
	}

	s.publishWebhook(ctx, domainWebhook.EventReleaseCreated, releaseWebhookPayload(r))
	return r, nil
}

// GetRelease はリリースを取得する
func (s *ApplicationService) GetRelease(ctx context.Context, id string) (*release.Release, error) {
	rid, err := release.NewReleaseID(id)
	if err != nil {
		return nil, fmt.Errorf("IDエラー: %w", err)
	}
	r, err := s.repository.FindByID(ctx, rid)
	if err != nil {
		return nil, fmt.Errorf("リリース取得エラー: %w", err)
	}
	return r, nil
}

// UpdateRelease はリリースを更新する
func (s *ApplicationService) UpdateRelease(ctx context.Context, input UpdateInput) error {
	rid, err := release.NewReleaseID(input.ID)
	if err != nil {
		return fmt.Errorf("IDエラー: %w", err)
	}

	r, err := s.repository.FindByID(ctx, rid)
	if err != nil {
		return fmt.Errorf("リリース取得エラー: %w", err)
	}

	if input.Title != nil {
		t, err := release.NewReleaseTitle(*input.Title)
		if err != nil {
			return fmt.Errorf("タイトルエラー: %w", err)
		}
		if err := r.ChangeTitle(t); err != nil {
			return err
		}
	}
	if input.ReleaseType != nil {
		rt, err := release.NewReleaseType(*input.ReleaseType)
		if err != nil {
			return fmt.Errorf("リリース種別エラー: %w", err)
		}
		r.UpdateType(rt)
	}
	if input.ReleaseDate != nil {
		d, err := release.NewReleaseDateFromString(*input.ReleaseDate)
		if err != nil {
			return fmt.Errorf("リリース日エラー: %w", err)
		}
		r.UpdateReleaseDate(d)
	}
	if input.Artists != nil {
		artists, err := buildArtistRefs(input.Artists)
		if err != nil {
			return err
		}
		if err := r.SetArtists(artists); err != nil {
			return err
		}
	}
	if input.Tracks != nil {
		tracks, err := buildTracks(input.Tracks)
		if err != nil {
			return err
		}
		if err := r.SetTracks(tracks); err != nil {
			return fmt.Errorf("収録曲エラー: %w", err)
		}
	}
	if input.StreamingLinks != nil {
		links, err := buildStreamingLinks(input.StreamingLinks)
		if err != nil {
			return err
		}
		r.UpdateStreamingLinks(links)
	}
	if input.CoverImageURL != nil {
		r.SetCoverImageURL(input.CoverImageURL)
	}
	if input.Aliases != nil {
		r.SetAliases(input.Aliases)
	}
	if input.TagIDs != nil {
		r.SetTags(input.TagIDs)
	}

	if err := s.repository.Update(ctx, r); err != nil {
		return fmt.Errorf("リリース更新エラー: %w", err)
	}

	s.publishWebhook(ctx, domainWebhook.EventReleaseUpdated, releaseWebhookPayload(r))
	return nil
}

// DeleteRelease はリリースをソフトデリートする
func (s *ApplicationService) DeleteRelease(ctx context.Context, id string) error {
	rid, err := release.NewReleaseID(id)
	if err != nil {
		return fmt.Errorf("IDエラー: %w", err)
	}
	if err := s.repository.Delete(ctx, rid); err != nil {
		return fmt.Errorf("リリース削除エラー: %w", err)
	}
	s.publishWebhook(ctx, domainWebhook.EventReleaseDeleted, map[string]interface{}{"id": id})
	return nil
}

// RestoreRelease はソフトデリートされたリリースを復元する
func (s *ApplicationService) RestoreRelease(ctx context.Context, id string) error {
	rid, err := release.NewReleaseID(id)
	if err != nil {
		return fmt.Errorf("IDエラー: %w", err)
	}
	if err := s.repository.Restore(ctx, rid); err != nil {
		return fmt.Errorf("リリース復元エラー: %w", err)
	}
	return nil
}

// SearchReleases は条件を指定してリリースを検索する（並行処理版）
func (s *ApplicationService) SearchReleases(ctx context.Context, criteria release.SearchCriteria) ([]*release.Release, int64, error) {
	var releases []*release.Release
	var total int64
	var errSearch, errCount error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		releases, errSearch = s.repository.Search(ctx, criteria)
	}()
	go func() {
		defer wg.Done()
		total, errCount = s.repository.Count(ctx, criteria)
	}()

	wg.Wait()

	if errSearch != nil {
		return nil, 0, fmt.Errorf("検索エラー: %w", errSearch)
	}
	if errCount != nil {
		return nil, 0, fmt.Errorf("件数取得エラー: %w", errCount)
	}

	return releases, total, nil
}

// UpdateStreamingLinks はストリーミングリンクを更新する
func (s *ApplicationService) UpdateStreamingLinks(ctx context.Context, input UpdateStreamingLinksInput) error {
	rid, err := release.NewReleaseID(input.ID)
	if err != nil {
		return fmt.Errorf("IDエラー: %w", err)
	}
	r, err := s.repository.FindByID(ctx, rid)
	if err != nil {
		return fmt.Errorf("リリース取得エラー: %w", err)
	}

	links, err := buildStreamingLinks(&input.Links)
	if err != nil {
		return err
	}
	r.UpdateStreamingLinks(links)

	if err := s.repository.Update(ctx, r); err != nil {
		return fmt.Errorf("リリース更新エラー: %w", err)
	}
	return nil
}

// UpdateExternalIDs は外部IDマッピングを更新する
func (s *ApplicationService) UpdateExternalIDs(ctx context.Context, input UpdateExternalIDsInput) error {
	rid, err := release.NewReleaseID(input.ID)
	if err != nil {
		return fmt.Errorf("IDエラー: %w", err)
	}
	r, err := s.repository.FindByID(ctx, rid)
	if err != nil {
		return fmt.Errorf("リリース取得エラー: %w", err)
	}

	extIDs := r.ExternalIDs()
	for k, v := range input.ExternalIDs {
		kind := release.ReleaseExternalIDKind(k)

		if v != "" {
			existing, err := s.repository.FindByExternalID(ctx, kind, v)
			if err != nil {
				return fmt.Errorf("外部ID重複チェックエラー: %w", err)
			}
			if existing != nil && existing.ID().Value() != input.ID {
				return fmt.Errorf("外部ID '%s' の値 '%s' は既に別のリリースに登録されています", k, v)
			}
		}

		if err := extIDs.Set(kind, v); err != nil {
			return fmt.Errorf("外部IDの設定エラー (%s): %w", k, err)
		}
	}

	r.UpdateExternalIDs(extIDs)
	if err := s.repository.Update(ctx, r); err != nil {
		return fmt.Errorf("リリース更新エラー: %w", err)
	}
	return nil
}

func (s *ApplicationService) publishWebhook(ctx context.Context, event domainWebhook.EventType, payload interface{}) {
	if s.publisher == nil {
		return
	}
	if err := s.publisher.Publish(ctx, event, payload); err != nil {
		slog.Error("リリースWebhook配信キュー投入に失敗しました", "event", event, "error", err)
	}
}

func releaseWebhookPayload(r *release.Release) map[string]interface{} {
	return map[string]interface{}{
		"id":           r.ID().Value(),
		"title":        r.Title().Value(),
		"release_type": r.ReleaseType().Value(),
		"release_date": r.ReleaseDate().String(),
	}
}

func buildArtistRefs(inputs []ArtistRefInput) ([]release.ArtistRef, error) {
	refs := make([]release.ArtistRef, 0, len(inputs))
	for _, a := range inputs {
		ref, err := release.NewArtistRef(release.ArtistKind(a.Kind), a.ID, a.Role)
		if err != nil {
			return nil, fmt.Errorf("アーティスト参照エラー: %w", err)
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func buildTracks(inputs []TrackInput) ([]release.Track, error) {
	tracks := make([]release.Track, 0, len(inputs))
	for _, t := range inputs {
		participants, err := buildTrackParticipants(t.Participants)
		if err != nil {
			return nil, fmt.Errorf("楽曲参加情報エラー (track %d): %w", t.TrackNumber, err)
		}
		track, err := release.NewTrack(t.TrackNumber, t.Title, t.DurationSec, t.ISRC, t.CoverImageURL, participants)
		if err != nil {
			return nil, fmt.Errorf("楽曲エラー (track %d): %w", t.TrackNumber, err)
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

func buildTrackParticipants(inputs []TrackParticipantInput) ([]release.TrackParticipant, error) {
	if inputs == nil {
		return nil, nil
	}
	participants := make([]release.TrackParticipant, 0, len(inputs))
	for _, input := range inputs {
		status, err := release.NewParticipationStatus(input.Status)
		if err != nil {
			return nil, err
		}
		participant, err := release.NewTrackParticipant(input.IdolID, status, input.Position)
		if err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}
	return participants, nil
}

func buildStreamingLinks(input *StreamingLinksInput) (*release.StreamingLinks, error) {
	links := release.NewStreamingLinks()
	if input == nil {
		return links, nil
	}
	type setter struct {
		val  *string
		fn   func(string) error
		name string
	}
	setters := []setter{
		{input.Spotify, links.SetSpotify, "Spotify"},
		{input.AppleMusic, links.SetAppleMusic, "Apple Music"},
		{input.YouTubeMusic, links.SetYouTubeMusic, "YouTube Music"},
		{input.YouTube, links.SetYouTube, "YouTube"},
		{input.LineMusic, links.SetLineMusic, "LINE Music"},
		{input.AmazonMusic, links.SetAmazonMusic, "Amazon Music"},
		{input.Official, links.SetOfficial, "公式サイト"},
	}
	for _, s := range setters {
		if s.val != nil {
			if err := s.fn(*s.val); err != nil {
				return nil, fmt.Errorf("%s URLエラー: %w", s.name, err)
			}
		}
	}
	return links, nil
}
