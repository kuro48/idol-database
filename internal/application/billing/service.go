package billing

import (
	"context"
	"fmt"

	appAPIKey "github.com/kuro48/idol-api/internal/application/apikey"
	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	domainbilling "github.com/kuro48/idol-api/internal/domain/billing"
	"github.com/kuro48/idol-api/internal/domain/plan"
)

const (
	// WebhookEventTypeCheckoutSessionCompleted は Checkout 完了イベント。
	WebhookEventTypeCheckoutSessionCompleted = "checkout.session.completed"
)

// Config は billing サービスの設定。
type Config struct {
	StripeSigningSecret string
	KeySeedSecret       string
	PriceIDs            map[plan.Type]string
}

// CheckoutSession は Stripe Checkout Session の最小表現。
type CheckoutSession struct {
	ID  string
	URL string
}

// CheckoutSessionCompleted は fulfillment に必要な Checkout 完了情報。
type CheckoutSessionCompleted struct {
	SessionID  string
	CustomerID string
	Email      string
	Name       string
	PlanType   plan.Type
}

// WebhookEvent は Stripe Webhook の最小表現。
type WebhookEvent struct {
	Type             string
	CheckoutSession  *CheckoutSessionCompleted
}

// PortalSession は Stripe Customer Portal Session の最小表現。
type PortalSession struct {
	URL string
}

// CreateCheckoutSessionInput は Stripe Checkout Session 作成入力。
type CreateCheckoutSessionInput struct {
	Email      string
	Name       string
	PlanType   string
	SuccessURL string
	CancelURL  string
	PriceID    string
}

// CreatePortalSessionInput は Stripe Portal Session 作成入力。
type CreatePortalSessionInput struct {
	CustomerID string
	ReturnURL  string
}

// CreatePortalSessionRequest は API から受ける portal session 作成入力。
type CreatePortalSessionRequest struct {
	Email     string
	ReturnURL string
}

// CreateCheckoutSessionResult は Checkout Session 作成結果。
type CreateCheckoutSessionResult struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// CreatePortalSessionResult は Portal Session 作成結果。
type CreatePortalSessionResult struct {
	URL string `json:"url"`
}

// APIKeyIssuedNotification は課金完了後の API キー通知。
type APIKeyIssuedNotification struct {
	To       string
	Name     string
	PlanType string
	RawKey   string
}

// StripeClient は Stripe とのやり取りを抽象化する。
type StripeClient interface {
	CreateCheckoutSession(ctx context.Context, input CreateCheckoutSessionInput) (*CheckoutSession, error)
	VerifyWebhookEvent(payload []byte, signature string) (*WebhookEvent, error)
	CreatePortalSession(ctx context.Context, input CreatePortalSessionInput) (*PortalSession, error)
}

// APIKeyIssuer は決済後に API キーを発行する契約。
type APIKeyIssuer interface {
	CreateOrGetKeyWithRawKey(ctx context.Context, input appAPIKey.CreateKeyInput, rawKey string) (*appAPIKey.CreateKeyOutput, error)
}

// Notifier は API キー発行通知を送る契約。
type Notifier interface {
	NotifyAPIKeyIssued(ctx context.Context, notification APIKeyIssuedNotification) error
}

// Service は Stripe 課金導線を提供する。
type Service struct {
	stripeClient StripeClient
	repo         domainbilling.FulfillmentRepository
	apiKeyIssuer APIKeyIssuer
	notifier     Notifier
	cfg          Config
}

// NewService は billing サービスを作成する。
func NewService(
	stripeClient StripeClient,
	repo domainbilling.FulfillmentRepository,
	apiKeyIssuer APIKeyIssuer,
	notifier Notifier,
	cfg Config,
) *Service {
	return &Service{
		stripeClient: stripeClient,
		repo:         repo,
		apiKeyIssuer: apiKeyIssuer,
		notifier:     notifier,
		cfg:          cfg,
	}
}

// CreateCheckoutSession は Stripe Checkout Session を作成する。
func (s *Service) CreateCheckoutSession(ctx context.Context, input CreateCheckoutSessionInput) (*CreateCheckoutSessionResult, error) {
	planType := plan.Type(input.PlanType)
	if planType != plan.TypeDeveloper && planType != plan.TypeBusiness {
		return nil, fmt.Errorf("有料プランのみ購入できます")
	}

	priceID := s.cfg.PriceIDs[planType]
	if priceID == "" {
		return nil, fmt.Errorf("プランに対応する Stripe Price ID が未設定です")
	}
	input.PriceID = priceID
	input.PlanType = string(planType)

	session, err := s.stripeClient.CreateCheckoutSession(ctx, input)
	if err != nil {
		return nil, err
	}

	return &CreateCheckoutSessionResult{ID: session.ID, URL: session.URL}, nil
}

// HandleStripeWebhook は Stripe Webhook を処理する。
func (s *Service) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := s.stripeClient.VerifyWebhookEvent(payload, signature)
	if err != nil {
		return err
	}
	if event == nil || event.Type != WebhookEventTypeCheckoutSessionCompleted || event.CheckoutSession == nil {
		return nil
	}

	completed := event.CheckoutSession
	fulfillment, err := s.repo.FindBySessionID(ctx, completed.SessionID)
	if err != nil {
		return err
	}

	var rawKey string
	if fulfillment == nil {
		rawKey, err = domainapikey.GenerateRawKeyFromSeed(s.cfg.KeySeedSecret, completed.SessionID)
		if err != nil {
			return err
		}

		output, err := s.apiKeyIssuer.CreateOrGetKeyWithRawKey(ctx, appAPIKey.CreateKeyInput{
			Email:    completed.Email,
			Name:     completed.Name,
			PlanType: string(completed.PlanType),
		}, rawKey)
		if err != nil {
			return err
		}

		fulfillment = domainbilling.NewCheckoutFulfillment(
			completed.SessionID,
			completed.CustomerID,
			completed.Email,
			completed.Name,
			completed.PlanType,
			output.Key.ID(),
		)
		if err := s.repo.Save(ctx, fulfillment); err != nil {
			return err
		}
	} else if !fulfillment.Notified() {
		rawKey, err = domainapikey.GenerateRawKeyFromSeed(s.cfg.KeySeedSecret, completed.SessionID)
		if err != nil {
			return err
		}
	}

	if !fulfillment.Notified() {
		if s.notifier == nil {
			return fmt.Errorf("APIキー通知手段が未設定です")
		}
		if err := s.notifier.NotifyAPIKeyIssued(ctx, APIKeyIssuedNotification{
			To:       fulfillment.Email(),
			Name:     fulfillment.Name(),
			PlanType: string(fulfillment.PlanType()),
			RawKey:   rawKey,
		}); err != nil {
			return err
		}
		fulfillment.MarkNotified()
		if err := s.repo.Update(ctx, fulfillment); err != nil {
			return err
		}
	}

	return nil
}

// CreatePortalSession は Billing Portal Session を作成する。
func (s *Service) CreatePortalSession(ctx context.Context, input CreatePortalSessionRequest) (*CreatePortalSessionResult, error) {
	fulfillment, err := s.repo.FindLatestByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if fulfillment == nil {
		return nil, fmt.Errorf("課金済みの顧客情報が見つかりません")
	}

	session, err := s.stripeClient.CreatePortalSession(ctx, CreatePortalSessionInput{
		CustomerID: fulfillment.CustomerID(),
		ReturnURL:  input.ReturnURL,
	})
	if err != nil {
		return nil, err
	}

	return &CreatePortalSessionResult{URL: session.URL}, nil
}
