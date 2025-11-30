package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/idol"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// IdolHandler はDDD構造を使用したアイドルハンドラー
type IdolHandler struct {
	appService *idol.ApplicationService
}

// NewIdolHandler はDDDハンドラーを作成する
func NewIdolHandler(appService *idol.ApplicationService) *IdolHandler {
	return &IdolHandler{
		appService: appService,
	}
}

// CreateIdolRequest はアイドル作成リクエスト
type CreateIdolRequest struct {
	Name      string  `json:"name" binding:"required,min=1,max=100"`
	Birthdate string  `json:"birthdate" binding:"omitempty,datetime=2006-01-02"`
	AgencyID  *string `json:"agency_id" binding:"omitempty"`
}

// UpdateIdolRequest はアイドル更新リクエスト
type UpdateIdolRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=100"`
	Birthdate *string `json:"birthdate" binding:"omitempty,datetime=2006-01-02"`
	AgencyID  *string `json:"agency_id" binding:"omitempty"`
}

// CreateIdol はアイドルを作成する
func (h *IdolHandler) CreateIdol(c *gin.Context) {
	var req CreateIdolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := idol.CreateIdolCommand{
		Name:      req.Name,
		Birthdate: &req.Birthdate,
		AgencyID:  req.AgencyID,
	}

	dto, err := h.appService.CreateIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("アイドルの作成に失敗しました"))
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetIdol はアイドルを取得する
func (h *IdolHandler) GetIdol(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	query := idol.GetIdolQuery{ID: id}

	dto, err := h.appService.GetIdol(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, middleware.NewNotFoundError("アイドル"))
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListIdols はアイドル一覧を取得する（検索機能付き）
func (h *IdolHandler) ListIdols(c *gin.Context) {
	var query idol.ListIdolsQuery

	// クエリパラメータをバインド
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです: "+err.Error()))
		return
	}

	// デフォルト値を適用
	query.ApplyDefaults()

	// バリデーション
	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	// 検索実行
	result, err := h.appService.SearchIdols(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("検索に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateIdol はアイドルを更新する
func (h *IdolHandler) UpdateIdol(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	var req UpdateIdolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := idol.UpdateIdolCommand{
		ID:        id,
		Name:      req.Name,
		Birthdate: req.Birthdate,
		AgencyID:  req.AgencyID,
	}

	err := h.appService.UpdateIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("アイドルの更新に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "アイドルが更新されました"})
}

// DeleteIdol はアイドルを削除する
func (h *IdolHandler) DeleteIdol(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	cmd := idol.DeleteIdolCommand{ID: id}

	err := h.appService.DeleteIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("アイドルの削除に失敗しました"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateSocialLinks はSNS/外部リンクを更新する
func (h *IdolHandler) UpdateSocialLinks(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	var cmd idol.UpdateSocialLinksCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd.ID = id

	err := h.appService.UpdateSocialLinks(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SNSリンクが更新されました"})
}
