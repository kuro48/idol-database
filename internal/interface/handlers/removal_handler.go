package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/removal"
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

// CreateRemovalRequest は削除申請を作成する
// POST /api/v1/removal-requests
func (h *RemovalHandler) CreateRemovalRequest(c *gin.Context) {
	var cmd removal.CreateRemovalRequestCommand

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "リクエストが不正です",
			"details": err.Error(),
		})
		return
	}

	dto, err := h.removalService.CreateRemovalRequest(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "削除申請の作成に失敗しました",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetRemovalRequest は削除申請を取得する
// GET /api/v1/removal-requests/:id
func (h *RemovalHandler) GetRemovalRequest(c *gin.Context) {
	id := c.Param("id")

	dto, err := h.removalService.GetRemovalRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "削除申請が見つかりません",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListAllRemovalRequests は全ての削除申請を取得する（管理者用）
// GET /api/v1/removal-requests
func (h *RemovalHandler) ListAllRemovalRequests(c *gin.Context) {
	dtos, err := h.removalService.ListAllRemovalRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "削除申請一覧の取得に失敗しました",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "保留中削除申請の取得に失敗しました",
			"details": err.Error(),
		})
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

	var cmd removal.UpdateStatusCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "リクエストが不正です",
			"details": err.Error(),
		})
		return
	}

	// URLのIDをコマンドに設定
	cmd.ID = id

	dto, err := h.removalService.UpdateStatus(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ステータスの更新に失敗しました",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto)
}
