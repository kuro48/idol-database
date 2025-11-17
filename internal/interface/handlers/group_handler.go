package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/group"
)

type GroupHandler struct {
	appService *group.ApplicationService
}

func NewGroupHandler(appService *group.ApplicationService) *GroupHandler {
	return  &GroupHandler{
		appService: appService,
	}
}

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required"`
	FormationDate  *string `json:"formation_date"`
	DisbandDate    *string `json:"disband_date"`
}

type UpdateGroupRequest struct {
	Name           *string `json:"name"`
	FormationDate  *string `json:"formation_date"`
	DisbandDate    *string `json:"disband_date"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := group.CreateGroupCommand {
		Name:          req.Name,
		FormationDate: req.FormationDate,
		DisbandDate:   req.DisbandDate,
	}

	dto, err := h.appService.CreateGroup(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	id := c.Param("id")
	query := group.GetGroupQuery{ID: id}

	dto, err := h.appService.GetGroup(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto)
}

func (h *GroupHandler) ListGroup(c *gin.Context) {
	query := group.ListGroupQuery{}

	dtos, err := h.appService.ListGroup(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dtos)
}

func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := group.UpdateGroupCommand{
		ID:            id,
		Name:          req.Name,
		FormationDate: req.FormationDate,
		DisbandDate:   req.DisbandDate,
	}

	err := h.appService.UpdateGroup(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "グループが更新されました"})
}

// DeleteGroup はグループを削除する
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")

	cmd := group.DeleteGroupCommand{ID: id}

	err := h.appService.DeleteGroup(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "グループが削除されました"})
}

