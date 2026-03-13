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
