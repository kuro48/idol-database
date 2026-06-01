package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/membership"
)

type MembershipHandler struct {
	usecase membership.MembershipUseCase
}

func NewMembershipHandler(uc membership.MembershipUseCase) *MembershipHandler {
	return &MembershipHandler{usecase: uc}
}

type CreateMembershipRequest struct {
	IdolID   string  `json:"idol_id" binding:"required"`
	GroupID  string  `json:"group_id" binding:"required"`
	Role     string  `json:"role" binding:"required"`
	JoinedAt *string `json:"joined_at" binding:"omitempty,datetime=2006-01-02"`
}

type UpdateMembershipRequest struct {
	Role     *string `json:"role"`
	JoinedAt *string `json:"joined_at" binding:"omitempty,datetime=2006-01-02"`
	LeftAt   *string `json:"left_at" binding:"omitempty,datetime=2006-01-02"`
}

// CreateMembership はメンバーシップを作成する
// @Summary      メンバーシップ作成
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Param        membership body CreateMembershipRequest true "メンバーシップ作成リクエスト"
// @Success      201 {object} membership.MembershipDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /memberships [post]
func (h *MembershipHandler) CreateMembership(c *gin.Context) {
	var req CreateMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := membership.CreateMembershipCommand{
		IdolID:   req.IdolID,
		GroupID:  req.GroupID,
		Role:     req.Role,
		JoinedAt: req.JoinedAt,
	}

	dto, err := h.usecase.CreateMembership(middleware.AuditContextFor(c), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "メンバーシップ",
			Message:  "メンバーシップの作成に失敗しました",
		})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetMembership はメンバーシップを取得する
// @Summary      メンバーシップ詳細取得
// @Tags         memberships
// @Produce      json
// @Param        id path string true "メンバーシップID"
// @Success      200 {object} membership.MembershipDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /memberships/{id} [get]
func (h *MembershipHandler) GetMembership(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dto, err := h.usecase.GetMembership(c.Request.Context(), membership.GetMembershipQuery{ID: id})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "メンバーシップ"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListMemberships はメンバーシップ一覧を取得する
// @Summary      メンバーシップ一覧取得
// @Tags         memberships
// @Produce      json
// @Param        idol_id   query string false "アイドルID"
// @Param        group_id  query string false "グループID"
// @Param        is_active query bool   false "アクティブのみ"
// @Param        role      query string false "ロール"
// @Param        sort      query string false "ソート項目" Enums(joined_at, left_at, created_at) default(created_at)
// @Param        order     query string false "ソート順" Enums(asc, desc) default(desc)
// @Param        page      query int    false "ページ番号" default(1)
// @Param        limit     query int    false "件数" default(20)
// @Success      200 {object} membership.MembershipSearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Router       /memberships [get]
func (h *MembershipHandler) ListMemberships(c *gin.Context) {
	var query membership.ListMembershipQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです"))
		return
	}

	result, err := h.usecase.ListMemberships(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "メンバーシップ一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListIdolMemberships はアイドルのメンバーシップ一覧を取得する
// @Summary      アイドルのメンバーシップ一覧
// @Tags         idols
// @Produce      json
// @Param        id path string true "アイドルID"
// @Success      200 {array} membership.MembershipDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /idols/{id}/memberships [get]
func (h *MembershipHandler) ListIdolMemberships(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dtos, err := h.usecase.ListByIdolID(c.Request.Context(), id)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "メンバーシップ"})
		return
	}

	c.JSON(http.StatusOK, dtos)
}

// ListGroupMemberships はグループのメンバーシップ一覧を取得する
// @Summary      グループのメンバーシップ一覧
// @Tags         groups
// @Produce      json
// @Param        id path string true "グループID"
// @Success      200 {array} membership.MembershipDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /groups/{id}/memberships [get]
func (h *MembershipHandler) ListGroupMemberships(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dtos, err := h.usecase.ListByGroupID(c.Request.Context(), id)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "メンバーシップ"})
		return
	}

	c.JSON(http.StatusOK, dtos)
}

// UpdateMembership はメンバーシップを更新する
// @Summary      メンバーシップ更新
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Param        id         path string true "メンバーシップID"
// @Param        membership body UpdateMembershipRequest true "メンバーシップ更新リクエスト"
// @Success      200 {object} map[string]string
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /memberships/{id} [put]
func (h *MembershipHandler) UpdateMembership(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req UpdateMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := membership.UpdateMembershipCommand{
		ID:       id,
		Role:     req.Role,
		JoinedAt: req.JoinedAt,
		LeftAt:   req.LeftAt,
	}

	if err := h.usecase.UpdateMembership(middleware.AuditContextFor(c), cmd); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "メンバーシップ",
			Message:  "メンバーシップの更新に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "メンバーシップが更新されました"})
}

// DeleteMembership はメンバーシップを削除する
// @Summary      メンバーシップ削除
// @Tags         memberships
// @Param        id path string true "メンバーシップID"
// @Success      204
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /memberships/{id} [delete]
func (h *MembershipHandler) DeleteMembership(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	cmd := membership.DeleteMembershipCommand{ID: id}
	if err := h.usecase.DeleteMembership(c.Request.Context(), cmd); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "メンバーシップ",
			Message:  "メンバーシップの削除に失敗しました",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
