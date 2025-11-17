package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/idol"
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
	Name        string `json:"name" binding:"required"`
	Birthdate   string `json:"birthdate"`
}

// UpdateIdolRequest はアイドル更新リクエスト
type UpdateIdolRequest struct {
	Name        *string `json:"name"`
	Birthdate   *string `json:"birthdate"`
}

// CreateIdol はアイドルを作成する
func (h *IdolHandler) CreateIdol(c *gin.Context) {
	var req CreateIdolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := idol.CreateIdolCommand{
		Name:      req.Name,
		Birthdate: &req.Birthdate,
	}

	dto, err := h.appService.CreateIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetIdol はアイドルを取得する
func (h *IdolHandler) GetIdol(c *gin.Context) {
	id := c.Param("id")

	query := idol.GetIdolQuery{ID: id}

	dto, err := h.appService.GetIdol(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListIdols はアイドル一覧を取得する
func (h *IdolHandler) ListIdols(c *gin.Context) {
	query := idol.ListIdolsQuery{}

	dtos, err := h.appService.ListIdols(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dtos)
}

// UpdateIdol はアイドルを更新する
func (h *IdolHandler) UpdateIdol(c *gin.Context) {
	id := c.Param("id")

	var req UpdateIdolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := idol.UpdateIdolCommand{
		ID:        id,
		Name:      req.Name,
		Birthdate: req.Birthdate,
	}

	err := h.appService.UpdateIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "アイドルが更新されました"})
}

// DeleteIdol はアイドルを削除する
func (h *IdolHandler) DeleteIdol(c *gin.Context) {
	id := c.Param("id")

	cmd := idol.DeleteIdolCommand{ID: id}

	err := h.appService.DeleteIdol(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "アイドルが削除されました"})
}
