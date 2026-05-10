package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/agency"
)

// AgencyHandler は事務所ハンドラー
type AgencyHandler struct {
	usecase agency.AgencyUseCase
}

// NewAgencyHandler は事務所ハンドラーを作成する
func NewAgencyHandler(usecase agency.AgencyUseCase) *AgencyHandler {
	return &AgencyHandler{
		usecase: usecase,
	}
}

// CreateAgency は事務所を作成する
// @Summary      事務所作成
// @Description  新しい事務所を作成する
// @Tags         agencies
// @Accept       json
// @Produce      json
// @Param        agency body agency.CreateAgencyCommand true "事務所作成リクエスト"
// @Success      201 {object} agency.AgencyDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /agencies [post]
func (h *AgencyHandler) CreateAgency(c *gin.Context) {
	var cmd agency.CreateAgencyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	dto, err := h.usecase.CreateAgency(middleware.AuditContextFor(c), cmd)
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
// @Summary      事務所詳細取得
// @Description  IDを指定して事務所情報を取得する
// @Tags         agencies
// @Produce      json
// @Param        id path string true "事務所ID"
// @Param        include query []string false "関連データ読み込み"
// @Success      200 {object} agency.AgencyDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /agencies/{id} [get]
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
// @Summary      事務所一覧取得
// @Description  条件を指定して事務所一覧を取得する
// @Tags         agencies
// @Produce      json
// @Param        name query string false "名前（部分一致）"
// @Param        country query string false "国コード"
// @Param        sort query string false "ソート項目" Enums(name, founded_date, created_at) default(created_at)
// @Param        order query string false "ソート順" Enums(asc, desc) default(desc)
// @Param        page query int false "ページ番号" default(1)
// @Param        limit query int false "1ページあたりの件数" default(20)
// @Success      200 {object} agency.AgencySearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /agencies [get]
func (h *AgencyHandler) ListAgencies(c *gin.Context) {
	var query agency.ListAgenciesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	result, err := h.usecase.ListAgencies(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "事務所一覧の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateAgency は事務所を更新する
// @Summary      事務所更新
// @Description  既存の事務所を更新する
// @Tags         agencies
// @Accept       json
// @Produce      json
// @Param        id path string true "事務所ID"
// @Param        agency body agency.UpdateAgencyCommand true "事務所更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /agencies/{id} [put]
func (h *AgencyHandler) UpdateAgency(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var cmd agency.UpdateAgencyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd.ID = id

	err := h.usecase.UpdateAgency(middleware.AuditContextFor(c), cmd)
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
// @Summary      事務所削除
// @Description  既存の事務所を削除する
// @Tags         agencies
// @Param        id path string true "事務所ID"
// @Success      204
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /agencies/{id} [delete]
func (h *AgencyHandler) DeleteAgency(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
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
