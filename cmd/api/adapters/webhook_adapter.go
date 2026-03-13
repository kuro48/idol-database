package adapters

import (
	"context"

	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/domain/webhook"
)

// WebhookAppAdapter は appWebhook.ApplicationService を handlers.webhookService インターフェースに適合させる
type WebhookAppAdapter struct {
	svc *appWebhook.ApplicationService
}

// NewWebhookAppAdapter は WebhookAppAdapter を作成する
func NewWebhookAppAdapter(svc *appWebhook.ApplicationService) *WebhookAppAdapter {
	return &WebhookAppAdapter{svc: svc}
}

func (a *WebhookAppAdapter) CreateSubscription(ctx context.Context, url string, events []webhook.EventType, createdBy string) (*webhook.Subscription, error) {
	return a.svc.CreateSubscription(ctx, appWebhook.CreateSubscriptionInput{
		URL:       url,
		Events:    events,
		CreatedBy: createdBy,
	})
}

func (a *WebhookAppAdapter) ListSubscriptions(ctx context.Context) ([]*webhook.Subscription, error) {
	return a.svc.ListSubscriptions(ctx)
}

func (a *WebhookAppAdapter) DeleteSubscription(ctx context.Context, id string) error {
	return a.svc.DeleteSubscription(ctx, id)
}

func (a *WebhookAppAdapter) VerifyWebhookRequest(ctx context.Context, subscriptionID, signature string, payload []byte) error {
	return a.svc.VerifyWebhookRequest(ctx, subscriptionID, signature, payload)
}
