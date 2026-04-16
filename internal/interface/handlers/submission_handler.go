package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/submission"
)

// SubmissionHandler は投稿審査のHTTPハンドラー
type SubmissionHandler struct {
	submissionUsecase submission.SubmissionUseCase
}

// NewSubmissionHandler は投稿審査ハンドラーを作成する
func NewSubmissionHandler(submissionUsecase submission.SubmissionUseCase) *SubmissionHandler {
	return &SubmissionHandler{
		submissionUsecase: submissionUsecase,
	}
}

// CreateSubmissionRequest は投稿審査作成リクエスト
type CreateSubmissionRequest struct {
	TargetType       string                 `json:"target_type" binding:"required,oneof=idol group agency event"`
	Payload          map[string]interface{} `json:"payload" binding:"required"`
	SourceURLs       []string               `json:"source_urls" binding:"required,min=1"`
	ContributorEmail string                 `json:"contributor_email" binding:"required,email"`
}

// UpdateStatusRequest はステータス更新リクエスト（管理者用）
type UpdateStatusRequest struct {
	Status       string `json:"status" binding:"required,oneof=approved rejected needs_revision"`
	ReviewedBy   string `json:"reviewed_by" binding:"required"`
	RevisionNote string `json:"revision_note"`
}

// ReviseSubmissionRequest は差し戻し後の再投稿リクエスト
type ReviseSubmissionRequest struct {
	Payload    map[string]interface{} `json:"payload" binding:"required"`
	SourceURLs []string               `json:"source_urls" binding:"required,min=1"`
}

// SubmissionListResponse は投稿審査一覧レスポンス（管理者用）
type SubmissionListResponse struct {
	Submissions []*submission.SubmissionDTO `json:"submissions"`
	Count       int                         `json:"count"`
}

// CreateSubmission は投稿審査を作成する
// @Summary      投稿審査作成
// @Description  新しい投稿審査を作成する（公開エンドポイント）
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Param        submission body CreateSubmissionRequest true "投稿審査作成リクエスト"
// @Success      201 {object} submission.PublicSubmissionDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /submissions [post]
func (h *SubmissionHandler) CreateSubmission(c *gin.Context) {
	var req CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := submission.CreateSubmissionCommand{
		TargetType:       req.TargetType,
		Payload:          req.Payload,
		SourceURLs:       req.SourceURLs,
		ContributorEmail: req.ContributorEmail,
	}

	dto, err := h.submissionUsecase.CreateSubmission(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "投稿審査",
			Message:  "投稿審査の作成に失敗しました",
		})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetSubmission は投稿審査を取得する（投稿者向け公開情報のみ）
// @Summary      投稿審査詳細取得
// @Description  IDを指定して投稿審査を取得する（公開情報のみ・メールアドレス等は除外）
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Param        id path string true "投稿審査ID"
// @Success      200 {object} submission.PublicSubmissionDTO
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /submissions/{id} [get]
func (h *SubmissionHandler) GetSubmission(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dto, err := h.submissionUsecase.GetSubmissionPublic(c.Request.Context(), id)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "投稿審査"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListAllSubmissions は全ての投稿審査を取得する（管理者用）
// @Summary      投稿審査一覧取得（管理者）
// @Description  全ての投稿審査を取得する（X-API-Key: admin スコープ必須）
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Param        X-API-Key header string true "Admin API Key"
// @Success      200 {object} SubmissionListResponse
// @Failure      401 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /submissions [get]
func (h *SubmissionHandler) ListAllSubmissions(c *gin.Context) {
	dtos, err := h.submissionUsecase.ListAllSubmissions(c.Request.Context())
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "投稿審査一覧の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": dtos,
		"count":       len(dtos),
	})
}

// ListPendingSubmissions は審査待ちの投稿審査を取得する（管理者用）
// @Summary      審査待ち投稿審査一覧取得（管理者）
// @Description  ステータスが pending の投稿審査のみを取得する（X-API-Key: admin スコープ必須）
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Param        X-API-Key header string true "Admin API Key"
// @Success      200 {object} SubmissionListResponse
// @Failure      401 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /submissions/pending [get]
func (h *SubmissionHandler) ListPendingSubmissions(c *gin.Context) {
	dtos, err := h.submissionUsecase.ListPendingSubmissions(c.Request.Context())
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Message: "審査待ち投稿審査の取得に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": dtos,
		"count":       len(dtos),
	})
}

// UpdateStatus は投稿審査のステータスを更新する（管理者用）
// @Summary      投稿審査ステータス更新（管理者）
// @Description  投稿審査のステータスを approved / rejected / needs_revision に更新する（X-API-Key: admin スコープ必須）
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Param        X-API-Key header string true "Admin API Key"
// @Param        id path string true "投稿審査ID"
// @Param        status body UpdateStatusRequest true "ステータス更新リクエスト"
// @Success      200 {object} submission.SubmissionDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      401 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /submissions/{id}/status [put]
func (h *SubmissionHandler) UpdateStatus(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	if req.Status == "needs_revision" && req.RevisionNote == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("差し戻し時は revision_note が必須です"))
		return
	}

	cmd := submission.UpdateStatusCommand{
		ID:           id,
		Status:       req.Status,
		ReviewedBy:   req.ReviewedBy,
		RevisionNote: req.RevisionNote,
	}

	dto, err := h.submissionUsecase.UpdateStatus(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "投稿審査",
			Message:  "ステータスの更新に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ReviseSubmission は差し戻し後の再投稿を行う（投稿者用）
// @Summary      差し戻し後の再投稿
// @Description  needs_revision 状態の投稿審査を修正して再投稿する（公開エンドポイント）
// @Tags         submissions
// @Accept       json
// @Produce      json
// @Param        id path string true "投稿審査ID"
// @Param        submission body ReviseSubmissionRequest true "再投稿リクエスト"
// @Success      200 {object} submission.PublicSubmissionDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /submissions/{id}/revise [put]
func (h *SubmissionHandler) ReviseSubmission(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	var req ReviseSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := submission.ReviseSubmissionCommand{
		ID:         id,
		Payload:    req.Payload,
		SourceURLs: req.SourceURLs,
	}

	dto, err := h.submissionUsecase.ReviseSubmission(c.Request.Context(), cmd)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{
			Resource: "投稿審査",
			Message:  "再投稿に失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, dto)
}
