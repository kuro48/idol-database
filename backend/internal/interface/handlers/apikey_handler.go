package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appAPIKey "github.com/kuro48/idol-api/internal/application/apikey"
	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// APIKeyHandler はAPIキー管理のHTTPハンドラー
type APIKeyHandler struct {
	service *appAPIKey.ApplicationService
}

// NewAPIKeyHandler はAPIKeyHandlerを作成する
func NewAPIKeyHandler(service *appAPIKey.ApplicationService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

type createAPIKeyRequest struct {
	Email    string `json:"email"     binding:"required,email"`
	Name     string `json:"name"      binding:"required,max=100"`
	PlanType string `json:"plan_type" binding:"required,oneof=free developer business"`
}

type apiKeyResponse struct {
	ID        string `json:"id"`
	MaskedKey string `json:"masked_key"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	PlanType  string `json:"plan_type"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// createAPIKeyResponse はAPIキー作成レスポンス（生キーを一度だけ含む）
type createAPIKeyResponse struct {
	apiKeyResponse
	RawKey string `json:"raw_key"`
}

// CreateAPIKey は新しいAPIキーを作成する
// @Summary     APIキーの作成
// @Tags        admin
// @Accept      json
// @Produce     json
// @Param       request body createAPIKeyRequest true "APIキー作成リクエスト"
// @Success     201 {object} createAPIKeyResponse
// @Failure     400 {object} middleware.ErrorResponse
// @Router      /admin/apikeys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "リクエストが不正です"})
		return
	}

	output, err := h.service.CreateKey(c.Request.Context(), appAPIKey.CreateKeyInput{
		Email:    req.Email,
		Name:     req.Name,
		PlanType: req.PlanType,
	})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "APIキーの作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, createAPIKeyResponse{
		apiKeyResponse: toAPIKeyResponse(output.Key),
		RawKey:         output.RawKey,
	})
}

// ListAPIKeys はメールアドレスに紐づくAPIキー一覧を返す
// @Summary     APIキー一覧取得
// @Tags        admin
// @Produce     json
// @Param       email query string true "メールアドレス"
// @Success     200 {array} apiKeyResponse
// @Failure     400 {object} middleware.ErrorResponse
// @Router      /admin/apikeys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("email クエリパラメータは必須です"))
		return
	}

	keys, err := h.service.ListKeysByEmail(c.Request.Context(), email)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "APIキーの取得に失敗しました"})
		return
	}

	resp := make([]apiKeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, toAPIKeyResponse(k))
	}
	c.JSON(http.StatusOK, resp)
}

// RevokeAPIKey はAPIキーを無効化する
// @Summary     APIキーの無効化
// @Tags        admin
// @Produce     json
// @Param       id path string true "APIキーID"
// @Success     204
// @Failure     500 {object} middleware.ErrorResponse
// @Router      /admin/apikeys/{id} [delete]
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.RevokeKey(c.Request.Context(), appAPIKey.RevokeKeyInput{ID: id}); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "APIキーの無効化に失敗しました"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toAPIKeyResponse(k *domainapikey.APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:        k.ID(),
		MaskedKey: k.MaskedKey(),
		Email:     k.Email(),
		Name:      k.Name(),
		PlanType:  string(k.PlanType()),
		IsActive:  k.IsActive(),
		CreatedAt: k.CreatedAt().UTC().Format("2006-01-02T15:04:05Z"),
	}
}
