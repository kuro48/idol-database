package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/removal"
)

// RemovalHandler は削除申請のHTTPハンドラー
type RemovalHandler struct {
	removalService removal.RemovalUseCase
}

// NewRemovalHandler は削除申請ハンドラーを作成する
func NewRemovalHandler(removalService removal.RemovalUseCase) *RemovalHandler {
	return &RemovalHandler{
		removalService: removalService,
	}
}

// CreateRemovalRequestDTO はバリデーション付きリクエスト
type CreateRemovalRequestDTO struct {
	TargetType    string `json:"target_type" binding:"required,oneof=idol group"`
	TargetID      string `json:"target_id" binding:"required"`
	RequesterType string `json:"requester_type" binding:"required,oneof=idol_themself agency third_party"`
	Reason        string `json:"reason" binding:"required,min=10,max=1000"`
	ContactInfo   string `json:"contact_info" binding:"required,email"`
	Evidence      string `json:"evidence" binding:"omitempty,url"`
	Description   string `json:"description" binding:"required,min=10,max=1000"`
}

// UpdateStatusDTO はステータス更新リクエスト
type UpdateStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=pending approved rejected"`
}

// RemovalRequestListResponse は削除申請一覧レスポンス
type RemovalRequestListResponse struct {
	RemovalRequests []*removal.RemovalRequestDTO `json:"removal_requests"`
	Count           int                          `json:"count"`
}

// CreateRemovalRequest は削除申請を作成する
// @Summary      削除申請作成
// @Description  新しい削除申請を作成する
// @Tags         removal-requests
// @Accept       json
// @Produce      json
// @Param        request body CreateRemovalRequestDTO true "削除申請作成リクエスト"
// @Success      201 {object} removal.CreateRemovalRequestResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /removal-requests [post]
func (h *RemovalHandler) CreateRemovalRequest(c *gin.Context) {
	var req CreateRemovalRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := removal.CreateRemovalRequestCommand{
		TargetType:    req.TargetType,
		TargetID:      req.TargetID,
		RequesterType: req.RequesterType,
		Reason:        req.Reason,
		ContactInfo:   req.ContactInfo,
		Evidence:      req.Evidence,
		Description:   req.Description,
	}

	result, err := h.removalService.CreateRemovalRequest(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "削除申請",
			Message:  "削除申請の作成に失敗しました",
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetRemovalRequest は削除申請を取得する
// @Summary      削除申請取得
// @Description  投稿者用アクセストークンで削除申請を取得する
// @Tags         removal-requests
// @Produce      json
// @Param        id path string true "削除申請ID"
// @Param        X-Access-Token header string false "投稿者用アクセストークン"
// @Param        access_token query string false "投稿者用アクセストークン"
// @Success      200 {object} removal.PublicRemovalRequestDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      401 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /removal-requests/{id} [get]
func (h *RemovalHandler) GetRemovalRequest(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}
	accessToken, ok := getAccessToken(c)
	if !ok {
		return
	}

	dto, err := h.removalService.GetRemovalRequestPublic(c.Request.Context(), id, accessToken)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "削除申請"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListAllRemovalRequests は全ての削除申請を取得する（管理者用）
// @Summary      削除申請一覧取得
// @Description  全ての削除申請を取得する（管理者用）
// @Tags         removal-requests
// @Produce      json
// @Success      200 {object} RemovalRequestListResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /removal-requests [get]
func (h *RemovalHandler) ListAllRemovalRequests(c *gin.Context) {
	dtos, err := h.removalService.ListAllRemovalRequests(c.Request.Context())
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "削除申請一覧の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"removal_requests": dtos,
		"count":            len(dtos),
	})
}

// ListPendingRemovalRequests は保留中の削除申請を取得する（管理者用）
// @Summary      保留中削除申請一覧取得
// @Description  保留中の削除申請を取得する（管理者用）
// @Tags         removal-requests
// @Produce      json
// @Success      200 {object} RemovalRequestListResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /removal-requests/pending [get]
func (h *RemovalHandler) ListPendingRemovalRequests(c *gin.Context) {
	dtos, err := h.removalService.ListPendingRemovalRequests(c.Request.Context())
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "保留中削除申請の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"removal_requests": dtos,
		"count":            len(dtos),
	})
}

// UpdateStatus はステータスを更新する（管理者用）
// @Summary      削除申請ステータス更新
// @Description  削除申請のステータスを更新する（管理者用）
// @Tags         removal-requests
// @Accept       json
// @Produce      json
// @Param        id path string true "削除申請ID"
// @Param        request body UpdateStatusDTO true "ステータス更新リクエスト"
// @Success      200 {object} removal.RemovalRequestDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /removal-requests/{id} [put]
func (h *RemovalHandler) UpdateStatus(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
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
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "削除申請",
			Message:  "ステータスの更新に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto)
}
