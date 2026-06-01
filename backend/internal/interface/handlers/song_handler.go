package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/song"
)

type SongHandler struct {
	usecase song.SongUseCase
}

func NewSongHandler(uc song.SongUseCase) *SongHandler {
	return &SongHandler{usecase: uc}
}

// CreateSong は楽曲を作成する
// @Summary      楽曲作成
// @Tags         songs
// @Accept       json
// @Produce      json
// @Param        song body song.CreateSongCommand true "楽曲作成リクエスト"
// @Success      201 {object} song.SongDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /songs [post]
func (h *SongHandler) CreateSong(c *gin.Context) {
	var cmd song.CreateSongCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	dto, err := h.usecase.CreateSong(middleware.AuditContextFor(c), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "楽曲", Message: "楽曲の作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetSong は楽曲を取得する
// @Summary      楽曲詳細取得
// @Tags         songs
// @Produce      json
// @Param        id path string true "楽曲ID"
// @Success      200 {object} song.SongDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /songs/{id} [get]
func (h *SongHandler) GetSong(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dto, err := h.usecase.GetSong(c.Request.Context(), song.GetSongQuery{ID: id})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "楽曲"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListSongs は楽曲一覧を取得する
// @Summary      楽曲一覧取得
// @Tags         songs
// @Produce      json
// @Param        title query string false "タイトル（部分一致）"
// @Param        isrc  query string false "ISRC"
// @Success      200 {object} song.SongSearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /songs [get]
func (h *SongHandler) ListSongs(c *gin.Context) {
	var query song.ListSongQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	result, err := h.usecase.ListSongs(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "楽曲一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateSong は楽曲を更新する
// @Summary      楽曲更新
// @Tags         songs
// @Accept       json
// @Produce      json
// @Param        id   path string true "楽曲ID"
// @Param        song body song.UpdateSongCommand true "楽曲更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /songs/{id} [put]
func (h *SongHandler) UpdateSong(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var cmd song.UpdateSongCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}
	cmd.ID = id

	if err := h.usecase.UpdateSong(middleware.AuditContextFor(c), cmd); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "楽曲", Message: "楽曲の更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "楽曲が更新されました"})
}

// DeleteSong は楽曲を削除する
// @Summary      楽曲削除
// @Tags         songs
// @Param        id path string true "楽曲ID"
// @Success      204
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /songs/{id} [delete]
func (h *SongHandler) DeleteSong(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	if err := h.usecase.DeleteSong(c.Request.Context(), song.DeleteSongCommand{ID: id}); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "楽曲", Message: "楽曲の削除に失敗しました"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
