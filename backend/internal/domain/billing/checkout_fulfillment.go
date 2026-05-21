package billing

import (
	"time"

	"github.com/kuro48/idol-api/internal/domain/plan"
)

// CheckoutFulfillment は Stripe Checkout 完了後のプロビジョニング状態を保持する。
type CheckoutFulfillment struct {
	sessionID  string
	customerID string
	email      string
	name       string
	planType   plan.Type
	apiKeyID   string
	notifiedAt *time.Time
	createdAt  time.Time
}

// NewCheckoutFulfillment は新しい fulfillment を作成する。
func NewCheckoutFulfillment(sessionID, customerID, email, name string, planType plan.Type, apiKeyID string) *CheckoutFulfillment {
	return &CheckoutFulfillment{
		sessionID:  sessionID,
		customerID: customerID,
		email:      email,
		name:       name,
		planType:   planType,
		apiKeyID:   apiKeyID,
		createdAt:  time.Now(),
	}
}

// ReconstructCheckoutFulfillment は永続化データから再構築する。
func ReconstructCheckoutFulfillment(
	sessionID, customerID, email, name string,
	planType plan.Type,
	apiKeyID string,
	notifiedAt *time.Time,
	createdAt time.Time,
) *CheckoutFulfillment {
	return &CheckoutFulfillment{
		sessionID:  sessionID,
		customerID: customerID,
		email:      email,
		name:       name,
		planType:   planType,
		apiKeyID:   apiKeyID,
		notifiedAt: notifiedAt,
		createdAt:  createdAt,
	}
}

func (f *CheckoutFulfillment) SessionID() string      { return f.sessionID }
func (f *CheckoutFulfillment) CustomerID() string     { return f.customerID }
func (f *CheckoutFulfillment) Email() string          { return f.email }
func (f *CheckoutFulfillment) Name() string           { return f.name }
func (f *CheckoutFulfillment) PlanType() plan.Type    { return f.planType }
func (f *CheckoutFulfillment) APIKeyID() string       { return f.apiKeyID }
func (f *CheckoutFulfillment) NotifiedAt() *time.Time { return f.notifiedAt }
func (f *CheckoutFulfillment) CreatedAt() time.Time   { return f.createdAt }

// Notified は API キー通知済みかを返す。
func (f *CheckoutFulfillment) Notified() bool {
	return f.notifiedAt != nil
}

// MarkNotified は通知済みにする。
func (f *CheckoutFulfillment) MarkNotified() {
	now := time.Now()
	f.notifiedAt = &now
}

// UpdatePlanType は現在の契約プランを同期する。
func (f *CheckoutFulfillment) UpdatePlanType(planType plan.Type) {
	f.planType = planType
}
