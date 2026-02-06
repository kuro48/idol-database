package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/idol"
)

// IdolHandler はDDD構造を使用したアイドルハンドラー
type IdolHandler struct {
	usecase *idol.Usecase
}

// NewIdolHandler はDDDハンドラーを作成する
func NewIdolHandler(usecase *idol.Usecase) *IdolHandler {
	return &IdolHandler{
		usecase: usecase,
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
// @Summary      アイドル作成
// @Description  新しいアイドルを作成する
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        idol body CreateIdolRequest true "アイドル作成リクエスト"
// @Success      201 {object} idol.IdolDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /idols [post]
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

	dto, err := h.usecase.CreateIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("アイドルの作成に失敗しました"))
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetIdol はアイドルを取得する
// @Summary      アイドル詳細取得
// @Description  IDを指定してアイドル情報を取得する
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        id path string true "アイドルID"
// @Param        include query string false "関連データ読み込み (カンマ区切り: agency,groups)"
// @Success      200 {object} idol.IdolDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /idols/{id} [get]
func (h *IdolHandler) GetIdol(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	query := idol.GetIdolQuery{ID: id}

	dto, err := h.usecase.GetIdol(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, middleware.NewNotFoundError("アイドル"))
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListIdols はアイドル一覧を取得する（検索機能付き）
// @Summary      アイドル一覧取得
// @Description  条件を指定してアイドル一覧を取得（検索・フィルタリング・ページネーション対応）
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        name query string false "名前（部分一致）"
// @Param        nationality query string false "国籍（完全一致）"
// @Param        group_id query string false "グループID"
// @Param        agency_id query string false "事務所ID"
// @Param        age_min query int false "最小年齢"
// @Param        age_max query int false "最大年齢"
// @Param        birthdate_from query string false "生年月日FROM (YYYY-MM-DD)"
// @Param        birthdate_to query string false "生年月日TO (YYYY-MM-DD)"
// @Param        include query string false "関連データ読み込み (カンマ区切り: agency,groups)"
// @Param        sort query string false "ソート項目" Enums(name, birthdate, created_at) default(created_at)
// @Param        order query string false "ソート順" Enums(asc, desc) default(desc)
// @Param        page query int false "ページ番号" default(1)
// @Param        limit query int false "1ページあたりの件数" default(20)
// @Success      200 {object} idol.SearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /idols [get]
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
	result, err := h.usecase.SearchIdols(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("検索に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateIdol はアイドルを更新する
// @Summary      アイドル更新
// @Description  IDを指定してアイドル情報を更新する
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        id path string true "アイドルID"
// @Param        idol body UpdateIdolRequest true "アイドル更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /idols/{id} [put]
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

	err := h.usecase.UpdateIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("アイドルの更新に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "アイドルが更新されました"})
}

// DeleteIdol はアイドルを削除する
// @Summary      アイドル削除
// @Description  IDを指定してアイドルを削除する
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        id path string true "アイドルID"
// @Success      204
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /idols/{id} [delete]
func (h *IdolHandler) DeleteIdol(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	cmd := idol.DeleteIdolCommand{ID: id}

	err := h.usecase.DeleteIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("アイドルの削除に失敗しました"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateSocialLinks はSNS/外部リンクを更新する
// @Summary      SNS/外部リンク更新
// @Description  アイドルのSNS/外部リンク情報を更新する
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        id path string true "アイドルID"
// @Param        social_links body idol.UpdateSocialLinksCommand true "SNS/外部リンク更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /idols/{id}/social-links [put]
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

	err := h.usecase.UpdateSocialLinks(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SNSリンクが更新されました"})
}
