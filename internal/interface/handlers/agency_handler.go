package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/agency"
)

// AgencyHandler は事務所ハンドラー
type AgencyHandler struct {
	usecase *agency.Usecase
}

// NewAgencyHandler は事務所ハンドラーを作成する
func NewAgencyHandler(usecase *agency.Usecase) *AgencyHandler {
	return &AgencyHandler{
		usecase: usecase,
	}
}

// CreateAgency は事務所を作成する
func (h *AgencyHandler) CreateAgency(c *gin.Context) {
	var cmd agency.CreateAgencyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	dto, err := h.usecase.CreateAgency(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "事務所",
			Message:  "事務所の作成に失敗しました",
		})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetAgency は事務所を取得する
func (h *AgencyHandler) GetAgency(c *gin.Context) {
	var query agency.GetAgencyQuery
	if err := c.ShouldBindUri(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	dto, err := h.usecase.GetAgency(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "事務所"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListAgencies は事務所一覧を取得する
func (h *AgencyHandler) ListAgencies(c *gin.Context) {
	var query agency.ListAgenciesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	dtos, err := h.usecase.ListAgencies(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "事務所一覧の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dtos)
}

// UpdateAgency は事務所を更新する
func (h *AgencyHandler) UpdateAgency(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	var cmd agency.UpdateAgencyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd.ID = id

	err := h.usecase.UpdateAgency(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "事務所",
			Message:  "事務所の更新に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "事務所が更新されました"})
}

// DeleteAgency は事務所を削除する
func (h *AgencyHandler) DeleteAgency(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	cmd := agency.DeleteAgencyCommand{ID: id}

	err := h.usecase.DeleteAgency(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "事務所",
			Message:  "事務所の削除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
