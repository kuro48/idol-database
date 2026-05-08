package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/release"
)

// ReleaseHandler はリリースハンドラー
type ReleaseHandler struct {
	usecase release.ReleaseUseCase
}

// NewReleaseHandler はリリースハンドラーを作成する
func NewReleaseHandler(usecase release.ReleaseUseCase) *ReleaseHandler {
	return &ReleaseHandler{usecase: usecase}
}

// ArtistRefRequest はアーティスト参照リクエスト
type ArtistRefRequest struct {
	Kind string `json:"kind" binding:"required,oneof=idol group"`
	ID   string `json:"id" binding:"required"`
	Role string `json:"role"`
}

// TrackRequest は収録曲リクエスト
type TrackRequest struct {
	TrackNumber   int     `json:"track_number" binding:"required,min=1"`
	Title         string  `json:"title" binding:"required,min=1,max=200"`
	DurationSec   *int    `json:"duration_sec" binding:"omitempty,min=0"`
	ISRC          *string `json:"isrc" binding:"omitempty"`
	CoverImageURL *string `json:"cover_image_url" binding:"omitempty,url"`
}

// StreamingLinksRequest はストリーミングリンクリクエスト
type StreamingLinksRequest struct {
	Spotify      *string `json:"spotify" binding:"omitempty,url"`
	AppleMusic   *string `json:"apple_music" binding:"omitempty,url"`
	YouTubeMusic *string `json:"youtube_music" binding:"omitempty,url"`
	YouTube      *string `json:"youtube" binding:"omitempty,url"`
	LineMusic    *string `json:"line_music" binding:"omitempty,url"`
	AmazonMusic  *string `json:"amazon_music" binding:"omitempty,url"`
	Official     *string `json:"official" binding:"omitempty,url"`
}

// CreateReleaseRequest はリリース作成リクエスト
type CreateReleaseRequest struct {
	Title          string                 `json:"title" binding:"required,min=1,max=200"`
	ReleaseType    string                 `json:"release_type" binding:"required"`
	ReleaseDate    string                 `json:"release_date" binding:"required,datetime=2006-01-02"`
	Artists        []ArtistRefRequest     `json:"artists" binding:"required,min=1,dive"`
	Tracks         []TrackRequest         `json:"tracks" binding:"omitempty,dive"`
	StreamingLinks *StreamingLinksRequest `json:"streaming_links" binding:"omitempty"`
	CoverImageURL  *string                `json:"cover_image_url" binding:"omitempty,url"`
	Aliases        []string               `json:"aliases" binding:"omitempty"`
	TagIDs         []string               `json:"tag_ids" binding:"omitempty"`
}

// UpdateReleaseRequest はリリース更新リクエスト
type UpdateReleaseRequest struct {
	Title          *string                `json:"title" binding:"omitempty,min=1,max=200"`
	ReleaseType    *string                `json:"release_type" binding:"omitempty"`
	ReleaseDate    *string                `json:"release_date" binding:"omitempty,datetime=2006-01-02"`
	Artists        []ArtistRefRequest     `json:"artists" binding:"omitempty,min=1,dive"`
	Tracks         []TrackRequest         `json:"tracks" binding:"omitempty,dive"`
	StreamingLinks *StreamingLinksRequest `json:"streaming_links" binding:"omitempty"`
	CoverImageURL  *string                `json:"cover_image_url" binding:"omitempty,url"`
	Aliases        []string               `json:"aliases" binding:"omitempty"`
	TagIDs         []string               `json:"tag_ids" binding:"omitempty"`
}

// UpdateStreamingLinksRequest はストリーミングリンク更新リクエスト
type UpdateStreamingLinksRequest struct {
	Spotify      *string `json:"spotify" binding:"omitempty,url"`
	AppleMusic   *string `json:"apple_music" binding:"omitempty,url"`
	YouTubeMusic *string `json:"youtube_music" binding:"omitempty,url"`
	YouTube      *string `json:"youtube" binding:"omitempty,url"`
	LineMusic    *string `json:"line_music" binding:"omitempty,url"`
	AmazonMusic  *string `json:"amazon_music" binding:"omitempty,url"`
	Official     *string `json:"official" binding:"omitempty,url"`
}

// UpdateReleaseExternalIDsRequest はリリース外部ID更新リクエスト
type UpdateReleaseExternalIDsRequest struct {
	ExternalIDs map[string]string `json:"external_ids" binding:"required"`
}

// CreateRelease はリリースを作成する
// @Summary      リリース作成
// @Description  新しいリリース（シングル・アルバム等）を作成する
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        release body CreateReleaseRequest true "リリース作成リクエスト"
// @Success      201 {object} release.ReleaseDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases [post]
func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	var req CreateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := release.CreateReleaseCommand{
		Title:         req.Title,
		ReleaseType:   req.ReleaseType,
		ReleaseDate:   req.ReleaseDate,
		Artists:       toArtistRefCommands(req.Artists),
		Tracks:        toTrackCommands(req.Tracks),
		CoverImageURL: req.CoverImageURL,
		Aliases:       req.Aliases,
		TagIDs:        req.TagIDs,
	}
	if req.StreamingLinks != nil {
		cmd.StreamingLinks = toStreamingLinksCommand(req.StreamingLinks)
	}

	dto, err := h.usecase.CreateRelease(middleware.AuditContextFor(c), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "リリースの作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetRelease はリリースを取得する
// @Summary      リリース詳細取得
// @Description  IDを指定してリリース情報を取得する
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        id path string true "リリースID"
// @Success      200 {object} release.ReleaseDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases/{id} [get]
func (h *ReleaseHandler) GetRelease(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dto, err := h.usecase.GetRelease(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, middleware.NewNotFoundError("リリース"))
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListReleases はリリース一覧を取得する
// @Summary      リリース一覧取得
// @Description  条件を指定してリリース一覧を取得（検索・フィルタリング・ページネーション対応）
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        title query string false "タイトル（部分一致）"
// @Param        release_type query string false "リリース種別" Enums(single, album, ep, mini_album, digital_single, compilation)
// @Param        artist_id query string false "アーティストID"
// @Param        artist_kind query string false "アーティスト種別" Enums(idol, group)
// @Param        release_date_from query string false "リリース日FROM (YYYY-MM-DD)"
// @Param        release_date_to query string false "リリース日TO (YYYY-MM-DD)"
// @Param        sort query string false "ソート項目" Enums(release_date, title, created_at) default(release_date)
// @Param        order query string false "ソート順" Enums(asc, desc) default(desc)
// @Param        page query int false "ページ番号" default(1)
// @Param        limit query int false "1ページあたりの件数" default(20)
// @Success      200 {object} release.SearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases [get]
func (h *ReleaseHandler) ListReleases(c *gin.Context) {
	var query release.ListReleasesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです: "+err.Error()))
		return
	}

	query.ApplyDefaults()

	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.usecase.SearchReleases(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "検索に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateRelease はリリースを更新する
// @Summary      リリース更新
// @Description  IDを指定してリリース情報を更新する
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        id path string true "リリースID"
// @Param        release body UpdateReleaseRequest true "リリース更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases/{id} [put]
func (h *ReleaseHandler) UpdateRelease(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req UpdateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := release.UpdateReleaseCommand{
		ID:            id,
		Title:         req.Title,
		ReleaseType:   req.ReleaseType,
		ReleaseDate:   req.ReleaseDate,
		Artists:       toArtistRefCommands(req.Artists),
		Tracks:        toTrackCommands(req.Tracks),
		CoverImageURL: req.CoverImageURL,
		Aliases:       req.Aliases,
		TagIDs:        req.TagIDs,
	}
	if req.StreamingLinks != nil {
		cmd.StreamingLinks = toStreamingLinksCommand(req.StreamingLinks)
	}

	if err := h.usecase.UpdateRelease(middleware.AuditContextFor(c), cmd); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "リリースの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "リリースが更新されました"})
}

// DeleteRelease はリリースを削除する
// @Summary      リリース削除
// @Description  IDを指定してリリースをソフトデリートする
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        id path string true "リリースID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases/{id} [delete]
func (h *ReleaseHandler) DeleteRelease(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	if err := h.usecase.DeleteRelease(middleware.AuditContextFor(c), release.DeleteReleaseCommand{ID: id}); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "リリースの削除に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "リリースが削除されました"})
}

// RestoreRelease はリリースを復元する
// @Summary      リリース復元
// @Description  ソフトデリートされたリリースを復元する
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        id path string true "リリースID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases/{id}/restore [put]
func (h *ReleaseHandler) RestoreRelease(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	if err := h.usecase.RestoreRelease(middleware.AuditContextFor(c), id); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "リリースの復元に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "リリースが復元されました"})
}

// UpdateStreamingLinks はストリーミングリンクを更新する
// @Summary      ストリーミングリンク更新
// @Description  リリースのストリーミングサービスリンクを更新する
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        id path string true "リリースID"
// @Param        links body UpdateStreamingLinksRequest true "ストリーミングリンク更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases/{id}/streaming-links [put]
func (h *ReleaseHandler) UpdateStreamingLinks(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req UpdateStreamingLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := release.UpdateStreamingLinksCommand{
		ID: id,
		Links: release.StreamingLinksCommand{
			Spotify:      req.Spotify,
			AppleMusic:   req.AppleMusic,
			YouTubeMusic: req.YouTubeMusic,
			YouTube:      req.YouTube,
			LineMusic:    req.LineMusic,
			AmazonMusic:  req.AmazonMusic,
			Official:     req.Official,
		},
	}

	if err := h.usecase.UpdateStreamingLinks(middleware.AuditContextFor(c), cmd); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "ストリーミングリンクの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ストリーミングリンクが更新されました"})
}

// UpdateExternalIDs は外部IDマッピングを更新する
// @Summary      外部ID更新
// @Description  リリースの外部サービスIDマッピングを更新する
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        id path string true "リリースID"
// @Param        external_ids body UpdateReleaseExternalIDsRequest true "外部ID更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /releases/{id}/external-ids [put]
func (h *ReleaseHandler) UpdateExternalIDs(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req UpdateReleaseExternalIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	if err := h.usecase.UpdateExternalIDs(middleware.AuditContextFor(c), release.UpdateExternalIDsCommand{
		ID:          id,
		ExternalIDs: req.ExternalIDs,
	}); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "リリース", Message: "外部IDの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "外部IDが更新されました"})
}

func toArtistRefCommands(reqs []ArtistRefRequest) []release.ArtistRefCommand {
	if reqs == nil {
		return nil
	}
	cmds := make([]release.ArtistRefCommand, 0, len(reqs))
	for _, r := range reqs {
		cmds = append(cmds, release.ArtistRefCommand{Kind: r.Kind, ID: r.ID, Role: r.Role})
	}
	return cmds
}

func toTrackCommands(reqs []TrackRequest) []release.TrackCommand {
	if reqs == nil {
		return nil
	}
	cmds := make([]release.TrackCommand, 0, len(reqs))
	for _, r := range reqs {
		cmds = append(cmds, release.TrackCommand{
			TrackNumber:   r.TrackNumber,
			Title:         r.Title,
			DurationSec:   r.DurationSec,
			ISRC:          r.ISRC,
			CoverImageURL: r.CoverImageURL,
		})
	}
	return cmds
}

func toStreamingLinksCommand(req *StreamingLinksRequest) *release.StreamingLinksCommand {
	if req == nil {
		return nil
	}
	return &release.StreamingLinksCommand{
		Spotify:      req.Spotify,
		AppleMusic:   req.AppleMusic,
		YouTubeMusic: req.YouTubeMusic,
		YouTube:      req.YouTube,
		LineMusic:    req.LineMusic,
		AmazonMusic:  req.AmazonMusic,
		Official:     req.Official,
	}
}
