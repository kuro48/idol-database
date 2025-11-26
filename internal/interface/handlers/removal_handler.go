package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/removal"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// RemovalHandler は削除申請のHTTPハンドラー
type RemovalHandler struct {
	removalService *removal.ApplicationService
}

// NewRemovalHandler は削除申請ハンドラーを作成する
func NewRemovalHandler(removalService *removal.ApplicationService) *RemovalHandler {
	return &RemovalHandler{
		removalService: removalService,
	}
}

// CreateRemovalRequestDTO はバリデーション付きリクエスト
type CreateRemovalRequestDTO struct {
	TargetType      string `json:"target_type" binding:"required,oneof=idol group"`
	TargetID        string `json:"target_id" binding:"required"`
	Reason          string `json:"reason" binding:"required,min=10,max=1000"`
	RequesterEmail  string `json:"requester_email" binding:"required,email"`
}

// UpdateStatusDTO はステータス更新リクエスト
type UpdateStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=pending approved rejected"`
}

// CreateRemovalRequest は削除申請を作成する
// POST /api/v1/removal-requests
func (h *RemovalHandler) CreateRemovalRequest(c *gin.Context) {
	var req CreateRemovalRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := removal.CreateRemovalRequestCommand{
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		Requester:   req.RequesterEmail,
		Reason:      req.Reason,
		ContactInfo: req.RequesterEmail,
		Description: req.Reason,
	}

	dto, err := h.removalService.CreateRemovalRequest(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("削除申請の作成に失敗しました"))
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetRemovalRequest は削除申請を取得する
// GET /api/v1/removal-requests/:id
func (h *RemovalHandler) GetRemovalRequest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	dto, err := h.removalService.GetRemovalRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, middleware.NewNotFoundError("削除申請"))
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListAllRemovalRequests は全ての削除申請を取得する（管理者用）
// GET /api/v1/removal-requests
func (h *RemovalHandler) ListAllRemovalRequests(c *gin.Context) {
	dtos, err := h.removalService.ListAllRemovalRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("削除申請一覧の取得に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"removal_requests": dtos,
		"count":            len(dtos),
	})
}

// ListPendingRemovalRequests は保留中の削除申請を取得する（管理者用）
// GET /api/v1/removal-requests/pending
func (h *RemovalHandler) ListPendingRemovalRequests(c *gin.Context) {
	dtos, err := h.removalService.ListPendingRemovalRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("保留中削除申請の取得に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"removal_requests": dtos,
		"count":            len(dtos),
	})
}

// UpdateStatus はステータスを更新する（管理者用）
// PUT /api/v1/removal-requests/:id
func (h *RemovalHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	var req UpdateStatusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := removal.UpdateStatusCommand{
		ID:     id,
		Status: req.Status,
	}

	dto, err := h.removalService.UpdateStatus(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("ステータスの更新に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, dto)
}
