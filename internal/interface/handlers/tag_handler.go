package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/tag"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// TagHandler はタグのハンドラー
type TagHandler struct {
	appService *tag.ApplicationService
}

// NewTagHandler はタグハンドラーを作成する
func NewTagHandler(appService *tag.ApplicationService) *TagHandler {
	return &TagHandler{
		appService: appService,
	}
}

// CreateTag はタグを作成する
// @Summary      タグ作成
// @Description  新しいタグを作成する
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        tag body tag.CreateTagCommand true "タグ作成リクエスト"
// @Success      201 {object} tag.TagDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var cmd tag.CreateTagCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	dto, err := h.appService.CreateTag(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetTag はタグを取得する
// @Summary      タグ詳細取得
// @Description  IDを指定してタグ情報を取得する
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id path string true "タグID"
// @Success      200 {object} tag.TagDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /tags/{id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	dto, err := h.appService.GetTag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, middleware.NewNotFoundError("タグが見つかりません"))
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListTags はタグ一覧を取得する（検索機能付き）
// @Summary      タグ一覧取得
// @Description  条件を指定してタグ一覧を取得（検索・フィルタリング・ページネーション対応）
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        name query string false "タグ名（部分一致）"
// @Param        category query string false "カテゴリ" Enums(genre, region, style, other)
// @Param        page query int false "ページ番号" default(1)
// @Param        limit query int false "1ページあたりの件数" default(20)
// @Success      200 {object} tag.SearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	// クエリパラメータの取得
	name := c.Query("name")
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// クエリオブジェクト構築
	query := tag.SearchQuery{
		Page:  page,
		Limit: limit,
	}

	if name != "" {
		query.Name = &name
	}
	if category != "" {
		query.Category = &category
	}

	// 検索実行
	baseURL := "/api/v1/tags"
	result, err := h.appService.SearchTags(c.Request.Context(), query, baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateTag はタグを更新する
// @Summary      タグ更新
// @Description  IDを指定してタグ情報を更新する
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id path string true "タグID"
// @Param        tag body tag.UpdateTagCommand true "タグ更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	var cmd tag.UpdateTagCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd.ID = id

	err := h.appService.UpdateTag(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "タグが更新されました"})
}

// DeleteTag はタグを削除する
// @Summary      タグ削除
// @Description  IDを指定してタグを削除する
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id path string true "タグID"
// @Success      204
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	err := h.appService.DeleteTag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
