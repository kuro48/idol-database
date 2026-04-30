package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	appBilling "github.com/kuro48/idol-api/internal/application/billing"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

const stripeSignatureHeader = "Stripe-Signature"

// BillingService は課金導線ハンドラーが依存する契約。
type BillingService interface {
	CreateCheckoutSession(ctx context.Context, input appBilling.CreateCheckoutSessionInput) (*appBilling.CreateCheckoutSessionResult, error)
	CreatePortalSession(ctx context.Context, input appBilling.CreatePortalSessionRequest) (*appBilling.CreatePortalSessionResult, error)
	HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error
}

// BillingHandler は Stripe 課金導線の HTTP ハンドラー。
type BillingHandler struct {
	service BillingService
}

// NewBillingHandler は BillingHandler を作成する。
func NewBillingHandler(service BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

type createCheckoutSessionRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Name       string `json:"name" binding:"required,max=100"`
	PlanType   string `json:"plan_type" binding:"required,oneof=developer business"`
	SuccessURL string `json:"success_url" binding:"required,url"`
	CancelURL  string `json:"cancel_url" binding:"required,url"`
}

type createPortalSessionRequest struct {
	ReturnURL string `json:"return_url" binding:"required,url"`
}

// CreateCheckoutSession は Stripe Checkout Session を作成する。
// @Summary     Checkout Session 作成
// @Description Developer または Business プラン購入用の Stripe Checkout Session を作成する
// @Tags        billing
// @Accept      json
// @Produce     json
// @Param       request body createCheckoutSessionRequest true "Checkout Session 作成リクエスト"
// @Success     201 {object} appBilling.CreateCheckoutSessionResult
// @Failure     400 {object} middleware.ErrorResponse
// @Failure     500 {object} middleware.ErrorResponse
// @Router      /billing/checkout-sessions [post]
func (h *BillingHandler) CreateCheckoutSession(c *gin.Context) {
	var req createCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	result, err := h.service.CreateCheckoutSession(c.Request.Context(), appBilling.CreateCheckoutSessionInput{
		Email:      req.Email,
		Name:       req.Name,
		PlanType:   req.PlanType,
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
	})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "Checkout Session の作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// CreatePortalSession は認証済み API キー所有者の Billing Portal Session を作成する。
// @Summary     Billing Portal Session 作成
// @Description 認証済み API キーに紐づく顧客の Stripe Billing Portal Session を作成する
// @Tags        billing
// @Accept      json
// @Produce     json
// @Param       Authorization header string true "Bearer API Key"
// @Param       request body createPortalSessionRequest true "Portal Session 作成リクエスト"
// @Success     201 {object} appBilling.CreatePortalSessionResult
// @Failure     400 {object} middleware.ErrorResponse
// @Failure     401 {object} middleware.ErrorResponse
// @Failure     500 {object} middleware.ErrorResponse
// @Router      /billing/portal-sessions [post]
func (h *BillingHandler) CreatePortalSession(c *gin.Context) {
	email := c.GetString(middleware.CtxKeyAPIKeyEmail)
	if email == "" {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return
	}

	var req createPortalSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	result, err := h.service.CreatePortalSession(c.Request.Context(), appBilling.CreatePortalSessionRequest{
		Email:     email,
		ReturnURL: req.ReturnURL,
	})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "Portal Session の作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// HandleStripeWebhook は Stripe Webhook を受信して処理する。
// @Summary     Stripe Webhook 受信
// @Description Stripe の checkout.session.completed Webhook を受信し API キー発行を行う
// @Tags        billing
// @Accept      json
// @Produce     json
// @Param       Stripe-Signature header string true "Stripe webhook signature"
// @Success     200 {object} map[string]string
// @Failure     400 {object} middleware.ErrorResponse
// @Failure     500 {object} middleware.ErrorResponse
// @Router      /billing/webhooks/stripe [post]
func (h *BillingHandler) HandleStripeWebhook(c *gin.Context) {
	signature := c.GetHeader(stripeSignatureHeader)
	if signature == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("Stripe-Signature ヘッダーは必須です"))
		return
	}

	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("Webhook ボディの読み取りに失敗しました"))
		return
	}

	if err := h.service.HandleStripeWebhook(c.Request.Context(), payload, signature); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "Stripe Webhook の処理に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
