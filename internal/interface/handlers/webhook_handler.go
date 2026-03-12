package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// WebhookHandler はWebhook管理ハンドラー
type WebhookHandler struct {
	appService *appWebhook.ApplicationService
}

// NewWebhookHandler はWebhookハンドラーを作成する
func NewWebhookHandler(appService *appWebhook.ApplicationService) *WebhookHandler {
	return &WebhookHandler{appService: appService}
}

// CreateSubscriptionRequest はWebhook購読作成リクエスト
type CreateSubscriptionRequest struct {
	URL    string   `json:"url" binding:"required,url"`
	Events []string `json:"events" binding:"required,min=1"`
}

// SubscriptionResponse はWebhook購読レスポンス
type SubscriptionResponse struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret,omitempty"` // 作成時のみ返す
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
	CreatedBy string   `json:"created_by"`
}

// CreateSubscription はWebhook購読を作成する
// @Summary      Webhook購読作成
// @Description  新しいWebhook購読を作成する（管理者専用）
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        request body CreateSubscriptionRequest true "購読設定"
// @Success      201 {object} SubscriptionResponse
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/webhooks [post]
func (h *WebhookHandler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	events := make([]webhook.EventType, len(req.Events))
	for i, e := range req.Events {
		events[i] = webhook.EventType(e)
	}

	sub, err := h.appService.CreateSubscription(middleware.AuditContextFor(c), appWebhook.CreateSubscriptionInput{
		URL:       req.URL,
		Events:    events,
		CreatedBy: middleware.GetActor(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("Webhook購読の作成に失敗しました"))
		return
	}

	c.JSON(http.StatusCreated, toSubscriptionResponse(sub, true))
}

// ListSubscriptions はWebhook購読一覧を返す
// @Summary      Webhook購読一覧
// @Description  Webhook購読一覧を返す（管理者専用）
// @Tags         webhooks
// @Produce      json
// @Success      200 {array} SubscriptionResponse
// @Router       /admin/webhooks [get]
func (h *WebhookHandler) ListSubscriptions(c *gin.Context) {
	subs, err := h.appService.ListSubscriptions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("Webhook購読一覧の取得に失敗しました"))
		return
	}

	responses := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		responses = append(responses, toSubscriptionResponse(sub, false))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// DeleteSubscription はWebhook購読を削除する
// @Summary      Webhook購読削除
// @Description  Webhook購読を削除する（管理者専用）
// @Tags         webhooks
// @Param        id path string true "購読ID"
// @Success      204
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /admin/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := h.appService.DeleteSubscription(c.Request.Context(), id); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "Webhook購読"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ReceiveWebhook はWebhookを受信して署名検証を行う
// POST /webhooks/receive/:subscription_id
func (h *WebhookHandler) ReceiveWebhook(c *gin.Context) {
	subscriptionID := c.Param("subscription_id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("サブスクリプションIDは必須です"))
		return
	}

	// 署名ヘッダーの確認
	signature := c.GetHeader("X-Webhook-Signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return
	}

	// リクエストボディを読み取る
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストボディの読み取りエラー"))
		return
	}

	// 署名検証
	if err := h.appService.VerifyWebhookRequest(c.Request.Context(), subscriptionID, signature, body); err != nil {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhookを受信しました"})
}

func toSubscriptionResponse(sub *webhook.Subscription, includeSecret bool) SubscriptionResponse {
	events := make([]string, len(sub.Events()))
	for i, e := range sub.Events() {
		events[i] = string(e)
	}
	resp := SubscriptionResponse{
		ID:        sub.ID(),
		URL:       sub.URL(),
		Events:    events,
		Active:    sub.Active(),
		CreatedAt: sub.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy: sub.CreatedBy(),
	}
	if includeSecret {
		resp.Secret = sub.Secret()
	}
	return resp
}
