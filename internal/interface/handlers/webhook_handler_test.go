package handlers_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/kuro48/idol-api/internal/interface/handlers"
)

// stubSubscriptionRepo はテスト用スタブリポジトリ
type stubSubscriptionRepo struct {
	subs map[string]*webhook.Subscription
}

func newStubSubscriptionRepo() *stubSubscriptionRepo {
	return &stubSubscriptionRepo{subs: make(map[string]*webhook.Subscription)}
}

func (r *stubSubscriptionRepo) Save(_ context.Context, sub *webhook.Subscription) error {
	r.subs[sub.ID()] = sub
	return nil
}

func (r *stubSubscriptionRepo) FindByID(_ context.Context, id string) (*webhook.Subscription, error) {
	sub, ok := r.subs[id]
	if !ok {
		return nil, fmt.Errorf("サブスクリプションが見つかりません: %s", id)
	}
	return sub, nil
}

func (r *stubSubscriptionRepo) FindAll(_ context.Context) ([]*webhook.Subscription, error) {
	result := make([]*webhook.Subscription, 0, len(r.subs))
	for _, s := range r.subs {
		result = append(result, s)
	}
	return result, nil
}

func (r *stubSubscriptionRepo) FindActiveByEvent(_ context.Context, event webhook.EventType) ([]*webhook.Subscription, error) {
	var result []*webhook.Subscription
	for _, s := range r.subs {
		if s.Active() && s.MatchesEvent(event) {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *stubSubscriptionRepo) Delete(_ context.Context, id string) error {
	delete(r.subs, id)
	return nil
}

// stubDeliveryRepo はテスト用スタブ配信リポジトリ
type stubDeliveryRepo struct{}

func (r *stubDeliveryRepo) Save(_ context.Context, _ *webhook.Delivery) error { return nil }
func (r *stubDeliveryRepo) Update(_ context.Context, _ *webhook.Delivery) error { return nil }
func (r *stubDeliveryRepo) FindByID(_ context.Context, _ string) (*webhook.Delivery, error) {
	return nil, fmt.Errorf("not found")
}
func (r *stubDeliveryRepo) FindPendingRetries(_ context.Context) ([]*webhook.Delivery, error) {
	return nil, nil
}

// computeTestSignature はテスト用にHMAC-SHA256署名を計算する
func computeTestSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// webhookAppAdapter は appWebhook.ApplicationService をテスト用に handlers.webhookService へ適合させる
type webhookAppAdapter struct {
	svc *appWebhook.ApplicationService
}

func (a *webhookAppAdapter) CreateSubscription(ctx context.Context, url string, events []webhook.EventType, createdBy string) (*webhook.Subscription, error) {
	return a.svc.CreateSubscription(ctx, appWebhook.CreateSubscriptionInput{URL: url, Events: events, CreatedBy: createdBy})
}

func (a *webhookAppAdapter) ListSubscriptions(ctx context.Context) ([]*webhook.Subscription, error) {
	return a.svc.ListSubscriptions(ctx)
}

func (a *webhookAppAdapter) DeleteSubscription(ctx context.Context, id string) error {
	return a.svc.DeleteSubscription(ctx, id)
}

func (a *webhookAppAdapter) VerifyWebhookRequest(ctx context.Context, subscriptionID, signature string, payload []byte) error {
	return a.svc.VerifyWebhookRequest(ctx, subscriptionID, signature, payload)
}

func setupTestRouter() (*gin.Engine, *stubSubscriptionRepo) {
	gin.SetMode(gin.TestMode)

	subRepo := newStubSubscriptionRepo()
	deliveryRepo := &stubDeliveryRepo{}
	appService := appWebhook.NewApplicationService(subRepo, deliveryRepo)
	h := handlers.NewWebhookHandler(&webhookAppAdapter{svc: appService})

	router := gin.New()
	router.POST("/api/v1/webhooks/receive/:subscription_id", h.ReceiveWebhook)

	return router, subRepo
}

func TestReceiveWebhook_NoSignatureHeader(t *testing.T) {
	router, subRepo := setupTestRouter()

	// テスト用サブスクリプション登録
	sub := webhook.NewSubscription("test-sub-id", "https://example.com", "test-secret", []webhook.EventType{webhook.EventIdolCreated}, "test")
	_ = subRepo.Save(context.Background(), sub)

	body := []byte(`{"event":"idol.created"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/receive/test-sub-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// X-Webhook-Signature ヘッダーなし

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReceiveWebhook_InvalidSignature(t *testing.T) {
	router, subRepo := setupTestRouter()

	// テスト用サブスクリプション登録
	sub := webhook.NewSubscription("test-sub-id-2", "https://example.com", "correct-secret", []webhook.EventType{webhook.EventIdolCreated}, "test")
	err := subRepo.Save(context.Background(), sub)
	require.NoError(t, err)

	body := []byte(`{"event":"idol.created"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/receive/test-sub-id-2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "sha256=invalidsignature")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReceiveWebhook_ValidSignature(t *testing.T) {
	router, subRepo := setupTestRouter()

	secret := "valid-secret-for-test"
	// テスト用サブスクリプション登録
	sub := webhook.NewSubscription("test-sub-id-3", "https://example.com", secret, []webhook.EventType{webhook.EventIdolCreated}, "test")
	err := subRepo.Save(context.Background(), sub)
	require.NoError(t, err)

	body := []byte(`{"event":"idol.created"}`)
	signature := computeTestSignature(secret, body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/receive/test-sub-id-3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiveWebhook_SubscriptionNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	body := []byte(`{"event":"idol.created"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/receive/nonexistent-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "sha256=somesignature")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
