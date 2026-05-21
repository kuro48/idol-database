package release

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	appRelease "github.com/kuro48/idol-api/internal/application/release"
	domainRelease "github.com/kuro48/idol-api/internal/domain/release"
)

// Usecase はリリースのユースケース
type Usecase struct {
	appService ReleaseAppPort
	idolApp    IdolExistencePort
	groupApp   GroupExistencePort
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService ReleaseAppPort, idolApp IdolExistencePort, groupApp GroupExistencePort) *Usecase {
	return &Usecase{appService: appService, idolApp: idolApp, groupApp: groupApp}
}

// CreateRelease はリリースを作成する
func (u *Usecase) CreateRelease(ctx context.Context, cmd CreateReleaseCommand) (*ReleaseDTO, error) {
	if err := u.validateArtists(ctx, cmd.Artists); err != nil {
		return nil, err
	}
	if err := u.validateTrackParticipants(ctx, cmd.Tracks); err != nil {
		return nil, err
	}

	input := appRelease.CreateInput{
		Title:         cmd.Title,
		ReleaseType:   cmd.ReleaseType,
		ReleaseDate:   cmd.ReleaseDate,
		Artists:       toAppArtistRefs(cmd.Artists),
		Tracks:        toAppTracks(cmd.Tracks),
		CoverImageURL: cmd.CoverImageURL,
		Aliases:       cmd.Aliases,
		TagIDs:        cmd.TagIDs,
	}
	if cmd.StreamingLinks != nil {
		input.StreamingLinks = toAppStreamingLinks(cmd.StreamingLinks)
	}

	r, err := u.appService.CreateRelease(ctx, input)
	if err != nil {
		return nil, err
	}
	return u.toDTO(r), nil
}

// GetRelease はリリースを取得する
func (u *Usecase) GetRelease(ctx context.Context, id string) (*ReleaseDTO, error) {
	r, err := u.appService.GetRelease(ctx, id)
	if err != nil {
		return nil, err
	}
	return u.toDTO(r), nil
}

// SearchReleases は条件を指定してリリースを検索する
func (u *Usecase) SearchReleases(ctx context.Context, query ListReleasesQuery) (*SearchResult, error) {
	criteria := u.queryToCriteria(query)

	releases, total, err := u.appService.SearchReleases(ctx, criteria)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ReleaseDTO, 0, len(releases))
	for _, r := range releases {
		dtos = append(dtos, u.toDTO(r))
	}

	meta := u.calcMeta(total, *query.Page, *query.Limit)
	links := u.buildLinks(query, meta.TotalPages)

	return &SearchResult{Data: dtos, Meta: meta, Links: links}, nil
}

// UpdateRelease はリリースを更新する
func (u *Usecase) UpdateRelease(ctx context.Context, cmd UpdateReleaseCommand) error {
	if err := u.validateArtists(ctx, cmd.Artists); err != nil {
		return err
	}
	if err := u.validateTrackParticipants(ctx, cmd.Tracks); err != nil {
		return err
	}

	return u.appService.UpdateRelease(ctx, appRelease.UpdateInput{
		ID:             cmd.ID,
		Title:          cmd.Title,
		ReleaseType:    cmd.ReleaseType,
		ReleaseDate:    cmd.ReleaseDate,
		Artists:        toAppArtistRefs(cmd.Artists),
		Tracks:         toAppTracks(cmd.Tracks),
		StreamingLinks: toAppStreamingLinks(cmd.StreamingLinks),
		CoverImageURL:  cmd.CoverImageURL,
		Aliases:        cmd.Aliases,
		TagIDs:         cmd.TagIDs,
	})
}

// DeleteRelease はリリースを削除する
func (u *Usecase) DeleteRelease(ctx context.Context, cmd DeleteReleaseCommand) error {
	return u.appService.DeleteRelease(ctx, cmd.ID)
}

// RestoreRelease はリリースを復元する
func (u *Usecase) RestoreRelease(ctx context.Context, id string) error {
	return u.appService.RestoreRelease(ctx, id)
}

// UpdateStreamingLinks はストリーミングリンクを更新する
func (u *Usecase) UpdateStreamingLinks(ctx context.Context, cmd UpdateStreamingLinksCommand) error {
	return u.appService.UpdateStreamingLinks(ctx, appRelease.UpdateStreamingLinksInput{
		ID:    cmd.ID,
		Links: toAppStreamingLinksValue(cmd.Links),
	})
}

// UpdateExternalIDs は外部IDマッピングを更新する
func (u *Usecase) UpdateExternalIDs(ctx context.Context, cmd UpdateExternalIDsCommand) error {
	return u.appService.UpdateExternalIDs(ctx, appRelease.UpdateExternalIDsInput{
		ID:          cmd.ID,
		ExternalIDs: cmd.ExternalIDs,
	})
}

// validateArtists はアーティスト参照の存在確認を行う
func (u *Usecase) validateArtists(ctx context.Context, artists []ArtistRefCommand) error {
	for _, a := range artists {
		switch a.Kind {
		case "idol":
			if err := u.idolApp.GetIdol(ctx, a.ID); err != nil {
				return fmt.Errorf("アイドルID '%s' が見つかりません: %w", a.ID, err)
			}
		case "group":
			if err := u.groupApp.GetGroup(ctx, a.ID); err != nil {
				return fmt.Errorf("グループID '%s' が見つかりません: %w", a.ID, err)
			}
		}
	}
	return nil
}

// validateTrackParticipants は楽曲参加アイドルの存在確認を行う。
func (u *Usecase) validateTrackParticipants(ctx context.Context, tracks []TrackCommand) error {
	for _, track := range tracks {
		for _, participant := range track.Participants {
			if err := u.idolApp.GetIdol(ctx, participant.IdolID); err != nil {
				return fmt.Errorf("楽曲 '%s' のアイドルID '%s' が見つかりません: %w", track.Title, participant.IdolID, err)
			}
		}
	}
	return nil
}

func (u *Usecase) queryToCriteria(query ListReleasesQuery) domainRelease.SearchCriteria {
	criteria := domainRelease.SearchCriteria{
		Title:    query.Title,
		ArtistID: query.ArtistID,
		Sort:     *query.Sort,
		Order:    *query.Order,
		Offset:   (*query.Page - 1) * *query.Limit,
		Limit:    *query.Limit,
	}

	if query.ReleaseType != nil {
		releaseType := domainRelease.ReleaseType(*query.ReleaseType)
		criteria.ReleaseType = &releaseType
	}
	if query.ArtistKind != nil {
		artistKind := domainRelease.ArtistKind(*query.ArtistKind)
		criteria.ArtistKind = &artistKind
	}
	if query.ReleaseDateFrom != nil {
		if t, err := time.Parse("2006-01-02", *query.ReleaseDateFrom); err == nil {
			criteria.ReleaseDateFrom = &t
		}
	}
	if query.ReleaseDateTo != nil {
		if t, err := time.Parse("2006-01-02", *query.ReleaseDateTo); err == nil {
			criteria.ReleaseDateTo = &t
		}
	}

	return criteria
}

func (u *Usecase) calcMeta(total int64, page, perPage int) *PaginationMeta {
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

func (u *Usecase) buildLinks(query ListReleasesQuery, totalPages int) *PaginationLinks {
	const baseURL = "/api/v1/releases"
	build := func(page int) string {
		p := url.Values{}
		p.Set("page", strconv.Itoa(page))
		p.Set("limit", strconv.Itoa(*query.Limit))
		if query.Title != nil {
			p.Set("title", *query.Title)
		}
		if query.ReleaseType != nil {
			p.Set("release_type", *query.ReleaseType)
		}
		if query.ArtistID != nil {
			p.Set("artist_id", *query.ArtistID)
		}
		if query.ArtistKind != nil {
			p.Set("artist_kind", *query.ArtistKind)
		}
		if query.ReleaseDateFrom != nil {
			p.Set("release_date_from", *query.ReleaseDateFrom)
		}
		if query.ReleaseDateTo != nil {
			p.Set("release_date_to", *query.ReleaseDateTo)
		}
		if query.Sort != nil {
			p.Set("sort", *query.Sort)
		}
		if query.Order != nil {
			p.Set("order", *query.Order)
		}
		return baseURL + "?" + p.Encode()
	}

	links := &PaginationLinks{First: build(1), Last: build(totalPages)}
	if *query.Page < totalPages {
		v := build(*query.Page + 1)
		links.Next = &v
	}
	if *query.Page > 1 {
		v := build(*query.Page - 1)
		links.Prev = &v
	}
	return links
}

func (u *Usecase) toDTO(r *domainRelease.Release) *ReleaseDTO {
	artists := make([]ArtistRefDTO, 0, len(r.Artists()))
	for _, a := range r.Artists() {
		artists = append(artists, ArtistRefDTO{
			Kind: string(a.Kind()),
			ID:   a.ID(),
			Role: a.Role(),
		})
	}

	tracks := make([]TrackDTO, 0, len(r.Tracks()))
	for _, t := range r.Tracks() {
		participants := make([]TrackParticipantDTO, 0, len(t.Participants()))
		for _, p := range t.Participants() {
			participants = append(participants, TrackParticipantDTO{
				IdolID:   p.IdolID(),
				Status:   p.Status().Value(),
				Position: p.Position(),
			})
		}
		tracks = append(tracks, TrackDTO{
			TrackNumber:   t.TrackNumber(),
			Title:         t.Title(),
			DurationSec:   t.DurationSec(),
			ISRC:          t.ISRC(),
			CoverImageURL: t.CoverImageURL(),
			Participants:  participants,
		})
	}

	var linksDTO *StreamingLinksDTO
	if sl := r.StreamingLinks(); sl != nil {
		linksDTO = &StreamingLinksDTO{
			Spotify:      sl.Spotify(),
			AppleMusic:   sl.AppleMusic(),
			YouTubeMusic: sl.YouTubeMusic(),
			YouTube:      sl.YouTube(),
			LineMusic:    sl.LineMusic(),
			AmazonMusic:  sl.AmazonMusic(),
			Official:     sl.Official(),
		}
	}

	var extIDs map[string]string
	if ids := r.ExternalIDs(); !ids.IsEmpty() {
		raw := ids.All()
		extIDs = make(map[string]string, len(raw))
		for k, v := range raw {
			extIDs[string(k)] = v
		}
	}

	return &ReleaseDTO{
		ID:             r.ID().Value(),
		Title:          r.Title().Value(),
		ReleaseType:    r.ReleaseType().Value(),
		ReleaseDate:    r.ReleaseDate().String(),
		Artists:        artists,
		Tracks:         tracks,
		StreamingLinks: linksDTO,
		CoverImageURL:  r.CoverImageURL(),
		Aliases:        r.Aliases(),
		TagIDs:         r.TagIDs(),
		ExternalIDs:    extIDs,
		CreatedAt:      r.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      r.UpdatedAt().Format(time.RFC3339),
	}
}

func toAppArtistRefs(cmds []ArtistRefCommand) []appRelease.ArtistRefInput {
	if cmds == nil {
		return nil
	}
	refs := make([]appRelease.ArtistRefInput, 0, len(cmds))
	for _, c := range cmds {
		refs = append(refs, appRelease.ArtistRefInput{Kind: c.Kind, ID: c.ID, Role: c.Role})
	}
	return refs
}

func toAppTracks(cmds []TrackCommand) []appRelease.TrackInput {
	if cmds == nil {
		return nil
	}
	tracks := make([]appRelease.TrackInput, 0, len(cmds))
	for _, c := range cmds {
		tracks = append(tracks, appRelease.TrackInput{
			TrackNumber:   c.TrackNumber,
			Title:         c.Title,
			DurationSec:   c.DurationSec,
			ISRC:          c.ISRC,
			CoverImageURL: c.CoverImageURL,
			Participants:  toAppTrackParticipants(c.Participants),
		})
	}
	return tracks
}

func toAppTrackParticipants(cmds []TrackParticipantCommand) []appRelease.TrackParticipantInput {
	if cmds == nil {
		return nil
	}
	participants := make([]appRelease.TrackParticipantInput, 0, len(cmds))
	for _, c := range cmds {
		participants = append(participants, appRelease.TrackParticipantInput{
			IdolID:   c.IdolID,
			Status:   c.Status,
			Position: c.Position,
		})
	}
	return participants
}

func toAppStreamingLinks(cmd *StreamingLinksCommand) *appRelease.StreamingLinksInput {
	if cmd == nil {
		return nil
	}
	return &appRelease.StreamingLinksInput{
		Spotify:      cmd.Spotify,
		AppleMusic:   cmd.AppleMusic,
		YouTubeMusic: cmd.YouTubeMusic,
		YouTube:      cmd.YouTube,
		LineMusic:    cmd.LineMusic,
		AmazonMusic:  cmd.AmazonMusic,
		Official:     cmd.Official,
	}
}

func toAppStreamingLinksValue(cmd StreamingLinksCommand) appRelease.StreamingLinksInput {
	return appRelease.StreamingLinksInput{
		Spotify:      cmd.Spotify,
		AppleMusic:   cmd.AppleMusic,
		YouTubeMusic: cmd.YouTubeMusic,
		YouTube:      cmd.YouTube,
		LineMusic:    cmd.LineMusic,
		AmazonMusic:  cmd.AmazonMusic,
		Official:     cmd.Official,
	}
}
