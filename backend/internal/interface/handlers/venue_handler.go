package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/venue"
)

type VenueHandler struct {
	usecase venue.VenueUseCase
}

func NewVenueHandler(uc venue.VenueUseCase) *VenueHandler {
	return &VenueHandler{usecase: uc}
}

// CreateVenue は会場を作成する
// @Summary      会場作成
// @Tags         venues
// @Accept       json
// @Produce      json
// @Param        venue body venue.CreateVenueCommand true "会場作成リクエスト"
// @Success      201 {object} venue.VenueDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /venues [post]
func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var cmd venue.CreateVenueCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	dto, err := h.usecase.CreateVenue(middleware.AuditContextFor(c), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "会場", Message: "会場の作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetVenue は会場を取得する
// @Summary      会場詳細取得
// @Tags         venues
// @Produce      json
// @Param        id path string true "会場ID"
// @Success      200 {object} venue.VenueDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /venues/{id} [get]
func (h *VenueHandler) GetVenue(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dto, err := h.usecase.GetVenue(c.Request.Context(), venue.GetVenueQuery{ID: id})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "会場"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListVenues は会場一覧を取得する
// @Summary      会場一覧取得
// @Tags         venues
// @Produce      json
// @Param        name       query string false "会場名（部分一致）"
// @Param        prefecture query string false "都道府県"
// @Success      200 {object} venue.VenueSearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /venues [get]
func (h *VenueHandler) ListVenues(c *gin.Context) {
	var query venue.ListVenueQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	result, err := h.usecase.ListVenues(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "会場一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateVenue は会場を更新する
// @Summary      会場更新
// @Tags         venues
// @Accept       json
// @Produce      json
// @Param        id    path string true "会場ID"
// @Param        venue body venue.UpdateVenueCommand true "会場更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /venues/{id} [put]
func (h *VenueHandler) UpdateVenue(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var cmd venue.UpdateVenueCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}
	cmd.ID = id

	if err := h.usecase.UpdateVenue(middleware.AuditContextFor(c), cmd); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "会場", Message: "会場の更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "会場が更新されました"})
}

// DeleteVenue は会場を削除する
// @Summary      会場削除
// @Tags         venues
// @Param        id path string true "会場ID"
// @Success      204
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /venues/{id} [delete]
func (h *VenueHandler) DeleteVenue(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	if err := h.usecase.DeleteVenue(c.Request.Context(), venue.DeleteVenueCommand{ID: id}); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "会場", Message: "会場の削除に失敗しました"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
