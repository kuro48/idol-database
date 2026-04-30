package billing

import "context"

// FulfillmentRepository は Checkout fulfillment の永続化契約。
type FulfillmentRepository interface {
	Save(ctx context.Context, fulfillment *CheckoutFulfillment) error
	FindBySessionID(ctx context.Context, sessionID string) (*CheckoutFulfillment, error)
	FindLatestByEmail(ctx context.Context, email string) (*CheckoutFulfillment, error)
	Update(ctx context.Context, fulfillment *CheckoutFulfillment) error
}
