package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	appAPIKey "github.com/kuro48/idol-api/internal/application/apikey"
	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	domainbilling "github.com/kuro48/idol-api/internal/domain/billing"
	"github.com/kuro48/idol-api/internal/domain/plan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStripeClient struct {
	checkoutInput *CreateCheckoutSessionInput
	portalInput   *CreatePortalSessionInput
	webhookEvent  *WebhookEvent
	checkoutOut   *CheckoutSession
	portalOut     *PortalSession
	verifyErr     error
}

func (f *fakeStripeClient) CreateCheckoutSession(_ context.Context, input CreateCheckoutSessionInput) (*CheckoutSession, error) {
	f.checkoutInput = &input
	if f.checkoutOut != nil {
		return f.checkoutOut, nil
	}
	return &CheckoutSession{ID: "cs_test_123", URL: "https://checkout.stripe.test/session"}, nil
}

func (f *fakeStripeClient) VerifyWebhookEvent(_ []byte, _ string) (*WebhookEvent, error) {
	if f.verifyErr != nil {
		return nil, f.verifyErr
	}
	return f.webhookEvent, nil
}

func (f *fakeStripeClient) CreatePortalSession(_ context.Context, input CreatePortalSessionInput) (*PortalSession, error) {
	f.portalInput = &input
	if f.portalOut != nil {
		return f.portalOut, nil
	}
	return &PortalSession{URL: "https://billing.stripe.test/session"}, nil
}

type fakeFulfillmentRepo struct {
	bySession  map[string]*domainbilling.CheckoutFulfillment
	latest     map[string]*domainbilling.CheckoutFulfillment
	byCustomer map[string]*domainbilling.CheckoutFulfillment
}

func newFakeFulfillmentRepo() *fakeFulfillmentRepo {
	return &fakeFulfillmentRepo{
		bySession:  make(map[string]*domainbilling.CheckoutFulfillment),
		latest:     make(map[string]*domainbilling.CheckoutFulfillment),
		byCustomer: make(map[string]*domainbilling.CheckoutFulfillment),
	}
}

func (f *fakeFulfillmentRepo) Save(_ context.Context, fulfillment *domainbilling.CheckoutFulfillment) error {
	f.bySession[fulfillment.SessionID()] = fulfillment
	f.latest[fulfillment.Email()] = fulfillment
	f.byCustomer[fulfillment.CustomerID()] = fulfillment
	return nil
}

func (f *fakeFulfillmentRepo) FindBySessionID(_ context.Context, sessionID string) (*domainbilling.CheckoutFulfillment, error) {
	return f.bySession[sessionID], nil
}

func (f *fakeFulfillmentRepo) FindLatestByEmail(_ context.Context, email string) (*domainbilling.CheckoutFulfillment, error) {
	return f.latest[email], nil
}

func (f *fakeFulfillmentRepo) FindLatestByCustomerID(_ context.Context, customerID string) (*domainbilling.CheckoutFulfillment, error) {
	return f.byCustomer[customerID], nil
}

func (f *fakeFulfillmentRepo) Update(_ context.Context, fulfillment *domainbilling.CheckoutFulfillment) error {
	f.bySession[fulfillment.SessionID()] = fulfillment
	f.latest[fulfillment.Email()] = fulfillment
	f.byCustomer[fulfillment.CustomerID()] = fulfillment
	return nil
}

type fakeAPIKeyIssuer struct {
	calls      int
	rawKey     string
	key        *domainapikey.APIKey
	updatedKey *domainapikey.APIKey
}

func (f *fakeAPIKeyIssuer) CreateOrGetKeyWithRawKey(_ context.Context, input appAPIKey.CreateKeyInput, rawKey string) (*appAPIKey.CreateKeyOutput, error) {
	f.calls++
	f.rawKey = rawKey
	if f.key == nil {
		key, err := domainapikey.Reconstruct(
			"507f1f77bcf86cd799439013",
			domainapikey.PrefixOf(rawKey),
			domainapikey.HashKey(rawKey),
			domainapikey.MaskKey(rawKey),
			input.Email,
			input.Name,
			plan.Type(input.PlanType),
			true,
			mustTime(),
		)
		if err != nil {
			return nil, err
		}
		f.key = key
	}
	return &appAPIKey.CreateKeyOutput{RawKey: rawKey, Key: f.key}, nil
}

func (f *fakeAPIKeyIssuer) UpdateKeyPlanAndStatus(_ context.Context, id string, planType string, active bool) (*domainapikey.APIKey, error) {
	key, err := domainapikey.Reconstruct(
		id,
		f.key.Prefix(),
		f.key.KeyHash(),
		f.key.MaskedKey(),
		f.key.Email(),
		f.key.Name(),
		plan.Type(planType),
		active,
		f.key.CreatedAt(),
	)
	if err != nil {
		return nil, err
	}
	f.key = key
	f.updatedKey = key
	return key, nil
}

type fakeNotifier struct {
	calls        int
	notification APIKeyIssuedNotification
	err          error
}

func (f *fakeNotifier) NotifyAPIKeyIssued(_ context.Context, notification APIKeyIssuedNotification) error {
	f.calls++
	f.notification = notification
	return f.err
}

func TestCreateCheckoutSession_UsesStripePriceID(t *testing.T) {
	t.Parallel()

	stripeClient := &fakeStripeClient{}
	service := NewService(
		stripeClient,
		newFakeFulfillmentRepo(),
		&fakeAPIKeyIssuer{},
		&fakeNotifier{},
		Config{
			StripeSigningSecret: "whsec_test",
			KeySeedSecret:       "seed",
			PriceIDs: map[plan.Type]string{
				plan.TypeDeveloper: "price_dev_123",
				plan.TypeBusiness:  "price_biz_123",
			},
		},
	)

	result, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		Email:      "user@example.com",
		Name:       "Example App",
		PlanType:   string(plan.TypeDeveloper),
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	})

	require.NoError(t, err)
	require.NotNil(t, stripeClient.checkoutInput)
	assert.Equal(t, "price_dev_123", stripeClient.checkoutInput.PriceID)
	assert.Equal(t, string(plan.TypeDeveloper), stripeClient.checkoutInput.PlanType)
	assert.Equal(t, "user@example.com", stripeClient.checkoutInput.Email)
	assert.Equal(t, "https://checkout.stripe.test/session", result.URL)
}

func TestHandleStripeWebhook_FulfillsCheckoutSessionIdempotently(t *testing.T) {
	t.Parallel()

	stripeClient := &fakeStripeClient{
		webhookEvent: &WebhookEvent{
			Type: WebhookEventTypeCheckoutSessionCompleted,
			CheckoutSession: &CheckoutSessionCompleted{
				SessionID:  "cs_test_123",
				CustomerID: "cus_123",
				Email:      "user@example.com",
				Name:       "Example App",
				PlanType:   plan.TypeBusiness,
			},
		},
	}
	repo := newFakeFulfillmentRepo()
	issuer := &fakeAPIKeyIssuer{}
	notifier := &fakeNotifier{}
	service := NewService(
		stripeClient,
		repo,
		issuer,
		notifier,
		Config{
			StripeSigningSecret: "whsec_test",
			KeySeedSecret:       "seed",
			PriceIDs: map[plan.Type]string{
				plan.TypeBusiness: "price_biz_123",
			},
		},
	)

	err := service.HandleStripeWebhook(context.Background(), []byte("{}"), "sig")
	require.NoError(t, err)

	fulfillment, err := repo.FindBySessionID(context.Background(), "cs_test_123")
	require.NoError(t, err)
	require.NotNil(t, fulfillment)
	assert.Equal(t, "cus_123", fulfillment.CustomerID())
	assert.True(t, fulfillment.Notified())
	assert.Equal(t, 1, issuer.calls)
	assert.Equal(t, 1, notifier.calls)

	err = service.HandleStripeWebhook(context.Background(), []byte("{}"), "sig")
	require.NoError(t, err)
	assert.Equal(t, 1, issuer.calls)
	assert.Equal(t, 1, notifier.calls)
}

func TestCreatePortalSession_UsesLatestFulfillmentCustomer(t *testing.T) {
	t.Parallel()

	stripeClient := &fakeStripeClient{}
	repo := newFakeFulfillmentRepo()
	fulfillment := domainbilling.NewCheckoutFulfillment(
		"cs_test_123",
		"cus_123",
		"user@example.com",
		"Example App",
		plan.TypeDeveloper,
		"507f1f77bcf86cd799439013",
	)
	require.NoError(t, repo.Save(context.Background(), fulfillment))

	service := NewService(
		stripeClient,
		repo,
		&fakeAPIKeyIssuer{},
		&fakeNotifier{},
		Config{
			StripeSigningSecret: "whsec_test",
			KeySeedSecret:       "seed",
		},
	)

	result, err := service.CreatePortalSession(context.Background(), CreatePortalSessionRequest{
		Email:     "user@example.com",
		ReturnURL: "https://example.com/account",
	})

	require.NoError(t, err)
	require.NotNil(t, stripeClient.portalInput)
	assert.Equal(t, "cus_123", stripeClient.portalInput.CustomerID)
	assert.Equal(t, "https://billing.stripe.test/session", result.URL)
}

func TestHandleStripeWebhook_ReturnsErrorWhenSignatureInvalid(t *testing.T) {
	t.Parallel()

	service := NewService(
		&fakeStripeClient{verifyErr: errors.New("invalid signature")},
		newFakeFulfillmentRepo(),
		&fakeAPIKeyIssuer{},
		&fakeNotifier{},
		Config{
			StripeSigningSecret: "whsec_test",
			KeySeedSecret:       "seed",
		},
	)

	err := service.HandleStripeWebhook(context.Background(), []byte("{}"), "bad")
	require.Error(t, err)
}

func TestHandleStripeWebhook_DeactivatesKeyOnSubscriptionDeleted(t *testing.T) {
	t.Parallel()

	stripeClient := &fakeStripeClient{
		webhookEvent: &WebhookEvent{
			Type: WebhookEventTypeSubscriptionDeleted,
			Subscription: &SubscriptionUpdated{
				CustomerID: "cus_123",
				PriceID:    "price_biz_123",
				Status:     "canceled",
			},
		},
	}
	repo := newFakeFulfillmentRepo()
	fulfillment := domainbilling.NewCheckoutFulfillment("cs_test_123", "cus_123", "user@example.com", "Example App", plan.TypeBusiness, "507f1f77bcf86cd799439013")
	require.NoError(t, repo.Save(context.Background(), fulfillment))
	issuer := &fakeAPIKeyIssuer{}
	issuer.key, _ = domainapikey.Reconstruct("507f1f77bcf86cd799439013", "ik_live_12345678", "hash", "ik_live_1234****5678", "user@example.com", "Example App", plan.TypeBusiness, true, mustTime())

	service := NewService(
		stripeClient,
		repo,
		issuer,
		&fakeNotifier{},
		Config{KeySeedSecret: "seed", PriceIDs: map[plan.Type]string{plan.TypeBusiness: "price_biz_123"}},
	)

	err := service.HandleStripeWebhook(context.Background(), []byte("{}"), "sig")
	require.NoError(t, err)
	require.NotNil(t, issuer.updatedKey)
	assert.False(t, issuer.updatedKey.IsActive())
	assert.Equal(t, plan.TypeBusiness, issuer.updatedKey.PlanType())
}

func TestHandleStripeWebhook_SyncsPlanOnSubscriptionUpdated(t *testing.T) {
	t.Parallel()

	stripeClient := &fakeStripeClient{
		webhookEvent: &WebhookEvent{
			Type: WebhookEventTypeSubscriptionUpdated,
			Subscription: &SubscriptionUpdated{
				CustomerID: "cus_123",
				PriceID:    "price_biz_123",
				Status:     "active",
			},
		},
	}
	repo := newFakeFulfillmentRepo()
	fulfillment := domainbilling.NewCheckoutFulfillment("cs_test_123", "cus_123", "user@example.com", "Example App", plan.TypeDeveloper, "507f1f77bcf86cd799439013")
	require.NoError(t, repo.Save(context.Background(), fulfillment))
	issuer := &fakeAPIKeyIssuer{}
	issuer.key, _ = domainapikey.Reconstruct("507f1f77bcf86cd799439013", "ik_live_12345678", "hash", "ik_live_1234****5678", "user@example.com", "Example App", plan.TypeDeveloper, true, mustTime())

	service := NewService(
		stripeClient,
		repo,
		issuer,
		&fakeNotifier{},
		Config{KeySeedSecret: "seed", PriceIDs: map[plan.Type]string{plan.TypeBusiness: "price_biz_123"}},
	)

	err := service.HandleStripeWebhook(context.Background(), []byte("{}"), "sig")
	require.NoError(t, err)
	require.NotNil(t, issuer.updatedKey)
	assert.True(t, issuer.updatedKey.IsActive())
	assert.Equal(t, plan.TypeBusiness, issuer.updatedKey.PlanType())

	updatedFulfillment, err := repo.FindLatestByCustomerID(context.Background(), "cus_123")
	require.NoError(t, err)
	require.NotNil(t, updatedFulfillment)
	assert.Equal(t, plan.TypeBusiness, updatedFulfillment.PlanType())
}

func TestHandleStripeWebhook_DeactivatesAndReactivatesKeyOnInvoiceEvents(t *testing.T) {
	t.Parallel()

	repo := newFakeFulfillmentRepo()
	fulfillment := domainbilling.NewCheckoutFulfillment("cs_test_123", "cus_123", "user@example.com", "Example App", plan.TypeDeveloper, "507f1f77bcf86cd799439013")
	require.NoError(t, repo.Save(context.Background(), fulfillment))
	issuer := &fakeAPIKeyIssuer{}
	issuer.key, _ = domainapikey.Reconstruct("507f1f77bcf86cd799439013", "ik_live_12345678", "hash", "ik_live_1234****5678", "user@example.com", "Example App", plan.TypeDeveloper, true, mustTime())

	service := NewService(
		&fakeStripeClient{
			webhookEvent: &WebhookEvent{
				Type: WebhookEventTypeInvoicePaymentFailed,
				Invoice: &InvoiceUpdated{
					CustomerID: "cus_123",
					Paid:       false,
				},
			},
		},
		repo,
		issuer,
		&fakeNotifier{},
		Config{KeySeedSecret: "seed"},
	)

	err := service.HandleStripeWebhook(context.Background(), []byte("{}"), "sig")
	require.NoError(t, err)
	require.NotNil(t, issuer.updatedKey)
	assert.False(t, issuer.updatedKey.IsActive())

	service.stripeClient = &fakeStripeClient{
		webhookEvent: &WebhookEvent{
			Type: WebhookEventTypeInvoicePaid,
			Invoice: &InvoiceUpdated{
				CustomerID: "cus_123",
				Paid:       true,
			},
		},
	}
	err = service.HandleStripeWebhook(context.Background(), []byte("{}"), "sig")
	require.NoError(t, err)
	assert.True(t, issuer.updatedKey.IsActive())
	assert.Equal(t, plan.TypeDeveloper, issuer.updatedKey.PlanType())
}

func mustTime() (tm time.Time) {
	return
}
