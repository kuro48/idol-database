package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/group"
)

type GroupHandler struct {
	usecase group.GroupUseCase
}

func NewGroupHandler(usecase group.GroupUseCase) *GroupHandler {
	return &GroupHandler{
		usecase: usecase,
	}
}

type CreateGroupRequest struct {
	Name          string  `json:"name" binding:"required,min=1,max=100"`
	FormationDate *string `json:"formation_date" binding:"omitempty,datetime=2006-01-02"`
	DisbandDate   *string `json:"disband_date" binding:"omitempty,datetime=2006-01-02"`
}

type UpdateGroupRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=100"`
	FormationDate *string `json:"formation_date" binding:"omitempty,datetime=2006-01-02"`
	DisbandDate   *string `json:"disband_date" binding:"omitempty,datetime=2006-01-02"`
}

// CreateGroup はグループを作成する
// @Summary      グループ作成
// @Description  新しいグループを作成する
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        group body CreateGroupRequest true "グループ作成リクエスト"
// @Success      201 {object} group.GroupDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := group.CreateGroupCommand{
		Name:          req.Name,
		FormationDate: req.FormationDate,
		DisbandDate:   req.DisbandDate,
	}

	dto, err := h.usecase.CreateGroup(middleware.AuditContextFor(c), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "グループ",
			Message:  "グループの作成に失敗しました",
		})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetGroup はグループを取得する
// @Summary      グループ詳細取得
// @Description  IDを指定してグループ情報を取得する
// @Tags         groups
// @Produce      json
// @Param        id path string true "グループID"
// @Success      200 {object} group.GroupDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /groups/{id} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	query := group.GetGroupQuery{ID: id}

	dto, err := h.usecase.GetGroup(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "グループ"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListGroup はグループ一覧を取得する
// @Summary      グループ一覧取得
// @Description  条件を指定してグループ一覧を取得する
// @Tags         groups
// @Produce      json
// @Param        name query string false "名前（部分一致）"
// @Param        sort query string false "ソート項目" Enums(name, formation_date, created_at) default(created_at)
// @Param        order query string false "ソート順" Enums(asc, desc) default(desc)
// @Param        page query int false "ページ番号" default(1)
// @Param        limit query int false "1ページあたりの件数" default(20)
// @Success      200 {object} group.GroupSearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /groups [get]
func (h *GroupHandler) ListGroup(c *gin.Context) {
	var query group.ListGroupQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	result, err := h.usecase.ListGroup(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "グループ一覧の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateGroup はグループを更新する
// @Summary      グループ更新
// @Description  既存のグループを更新する
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id path string true "グループID"
// @Param        group body UpdateGroupRequest true "グループ更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := group.UpdateGroupCommand{
		ID:            id,
		Name:          req.Name,
		FormationDate: req.FormationDate,
		DisbandDate:   req.DisbandDate,
	}

	err := h.usecase.UpdateGroup(middleware.AuditContextFor(c), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "グループ",
			Message:  "グループの更新に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "グループが更新されました"})
}

// DeleteGroup はグループを削除する
// @Summary      グループ削除
// @Description  既存のグループを削除する
// @Tags         groups
// @Param        id path string true "グループID"
// @Success      204
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	cmd := group.DeleteGroupCommand{ID: id}

	err := h.usecase.DeleteGroup(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "グループ",
			Message:  "グループの削除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
