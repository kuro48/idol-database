package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appBilling "github.com/kuro48/idol-api/internal/application/billing"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubBillingService struct {
	checkoutInput  *appBilling.CreateCheckoutSessionInput
	portalInput    *appBilling.CreatePortalSessionRequest
	webhookPayload []byte
	webhookSig     string
}

func (s *stubBillingService) CreateCheckoutSession(_ context.Context, input appBilling.CreateCheckoutSessionInput) (*appBilling.CreateCheckoutSessionResult, error) {
	s.checkoutInput = &input
	return &appBilling.CreateCheckoutSessionResult{
		ID:  "cs_test_123",
		URL: "https://checkout.stripe.test/session",
	}, nil
}

func (s *stubBillingService) CreatePortalSession(_ context.Context, input appBilling.CreatePortalSessionRequest) (*appBilling.CreatePortalSessionResult, error) {
	s.portalInput = &input
	return &appBilling.CreatePortalSessionResult{
		URL: "https://billing.stripe.test/session",
	}, nil
}

func (s *stubBillingService) HandleStripeWebhook(_ context.Context, payload []byte, signature string) error {
	s.webhookPayload = append([]byte(nil), payload...)
	s.webhookSig = signature
	return nil
}

func setupBillingRouter(service handlers.BillingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := handlers.NewBillingHandler(service)
	router.POST("/billing/checkout-sessions", h.CreateCheckoutSession)
	router.POST("/billing/portal-sessions", func(c *gin.Context) {
		c.Set(middleware.CtxPlanEmail, "paid@example.com")
		h.CreatePortalSession(c)
	})
	router.POST("/billing/webhooks/stripe", h.HandleStripeWebhook)
	return router
}

func setupBillingRouterWithRedirectOrigins(service handlers.BillingService, origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := handlers.NewBillingHandlerWithAllowedRedirectOrigins(service, origins)
	router.POST("/billing/checkout-sessions", h.CreateCheckoutSession)
	router.POST("/billing/portal-sessions", func(c *gin.Context) {
		c.Set(middleware.CtxPlanEmail, "paid@example.com")
		h.CreatePortalSession(c)
	})
	return router
}

func TestCreateCheckoutSession(t *testing.T) {
	t.Parallel()

	service := &stubBillingService{}
	router := setupBillingRouter(service)

	body := `{
		"email":"paid@example.com",
		"name":"Paid App",
		"plan_type":"developer",
		"success_url":"https://example.com/success",
		"cancel_url":"https://example.com/cancel"
	}`
	req := httptest.NewRequest(http.MethodPost, "/billing/checkout-sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, service.checkoutInput)
	assert.Equal(t, "paid@example.com", service.checkoutInput.Email)
	assert.Equal(t, "developer", service.checkoutInput.PlanType)
}

func TestCreatePortalSession_UsesAuthenticatedEmail(t *testing.T) {
	t.Parallel()

	service := &stubBillingService{}
	router := setupBillingRouter(service)

	body := `{"return_url":"https://example.com/account"}`
	req := httptest.NewRequest(http.MethodPost, "/billing/portal-sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, service.portalInput)
	assert.Equal(t, "paid@example.com", service.portalInput.Email)
	assert.Equal(t, "https://example.com/account", service.portalInput.ReturnURL)
}

func TestCreateCheckoutSession_RejectsUntrustedRedirectOrigin(t *testing.T) {
	t.Parallel()

	service := &stubBillingService{}
	router := setupBillingRouterWithRedirectOrigins(service, []string{"https://app.example.com"})

	body := `{
		"email":"paid@example.com",
		"name":"Paid App",
		"plan_type":"developer",
		"success_url":"https://evil.example.com/success",
		"cancel_url":"https://app.example.com/cancel"
	}`
	req := httptest.NewRequest(http.MethodPost, "/billing/checkout-sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Nil(t, service.checkoutInput)
}

func TestCreatePortalSession_RejectsUntrustedReturnOrigin(t *testing.T) {
	t.Parallel()

	service := &stubBillingService{}
	router := setupBillingRouterWithRedirectOrigins(service, []string{"https://app.example.com"})

	body := `{"return_url":"https://evil.example.com/account"}`
	req := httptest.NewRequest(http.MethodPost, "/billing/portal-sessions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Nil(t, service.portalInput)
}

func TestHandleStripeWebhook_UsesRawBodyAndSignature(t *testing.T) {
	t.Parallel()

	service := &stubBillingService{}
	router := setupBillingRouter(service)

	payload := []byte(`{"id":"evt_test","type":"checkout.session.completed"}`)
	req := httptest.NewRequest(http.MethodPost, "/billing/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", "t=123,v1=abc")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, payload, service.webhookPayload)
	assert.Equal(t, "t=123,v1=abc", service.webhookSig)
}
