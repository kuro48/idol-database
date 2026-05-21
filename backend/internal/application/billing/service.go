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
	WebhookEventTypeSubscriptionUpdated      = "customer.subscription.updated"
	WebhookEventTypeSubscriptionDeleted      = "customer.subscription.deleted"
	WebhookEventTypeInvoicePaymentFailed     = "invoice.payment_failed"
	WebhookEventTypeInvoicePaid              = "invoice.paid"
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
	Type            string
	CheckoutSession *CheckoutSessionCompleted
	Subscription    *SubscriptionUpdated
	Invoice         *InvoiceUpdated
}

// SubscriptionUpdated は subscription 更新時の最小情報。
type SubscriptionUpdated struct {
	CustomerID string
	PriceID    string
	Status     string
}

// InvoiceUpdated は invoice 更新時の最小情報。
type InvoiceUpdated struct {
	CustomerID string
	Paid       bool
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
	UpdateKeyPlanAndStatus(ctx context.Context, id string, planType string, active bool) (*domainapikey.APIKey, error)
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
	if event == nil {
		return nil
	}

	switch event.Type {
	case WebhookEventTypeCheckoutSessionCompleted:
		if event.CheckoutSession == nil {
			return nil
		}
		return s.handleCheckoutSessionCompleted(ctx, event.CheckoutSession)
	case WebhookEventTypeSubscriptionUpdated:
		if event.Subscription == nil {
			return nil
		}
		return s.syncSubscriptionState(ctx, event.Subscription)
	case WebhookEventTypeSubscriptionDeleted:
		if event.Subscription == nil {
			return nil
		}
		return s.syncSubscriptionState(ctx, &SubscriptionUpdated{
			CustomerID: event.Subscription.CustomerID,
			PriceID:    event.Subscription.PriceID,
			Status:     "canceled",
		})
	case WebhookEventTypeInvoicePaymentFailed:
		if event.Invoice == nil {
			return nil
		}
		return s.syncKeyStatusByCustomer(ctx, event.Invoice.CustomerID, false)
	case WebhookEventTypeInvoicePaid:
		if event.Invoice == nil {
			return nil
		}
		return s.syncKeyStatusByCustomer(ctx, event.Invoice.CustomerID, true)
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

func (s *Service) handleCheckoutSessionCompleted(ctx context.Context, completed *CheckoutSessionCompleted) error {
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

func (s *Service) syncSubscriptionState(ctx context.Context, subscription *SubscriptionUpdated) error {
	fulfillment, err := s.repo.FindLatestByCustomerID(ctx, subscription.CustomerID)
	if err != nil {
		return err
	}
	if fulfillment == nil {
		return nil
	}

	planType, err := s.planTypeFromPriceID(subscription.PriceID)
	if err != nil {
		return err
	}
	active := isActiveSubscriptionStatus(subscription.Status)
	if _, err := s.apiKeyIssuer.UpdateKeyPlanAndStatus(ctx, fulfillment.APIKeyID(), string(planType), active); err != nil {
		return err
	}
	if fulfillment.PlanType() != planType {
		fulfillment.UpdatePlanType(planType)
		if err := s.repo.Update(ctx, fulfillment); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) syncKeyStatusByCustomer(ctx context.Context, customerID string, active bool) error {
	fulfillment, err := s.repo.FindLatestByCustomerID(ctx, customerID)
	if err != nil {
		return err
	}
	if fulfillment == nil {
		return nil
	}
	_, err = s.apiKeyIssuer.UpdateKeyPlanAndStatus(ctx, fulfillment.APIKeyID(), string(fulfillment.PlanType()), active)
	return err
}

func (s *Service) planTypeFromPriceID(priceID string) (plan.Type, error) {
	for planType, configuredPriceID := range s.cfg.PriceIDs {
		if configuredPriceID == priceID {
			return planType, nil
		}
	}
	return "", fmt.Errorf("プランに対応する Stripe Price ID が見つかりません")
}

func isActiveSubscriptionStatus(status string) bool {
	switch status {
	case "active", "trialing":
		return true
	default:
		return false
	}
}
