package webhook

import "context"

// SubscriptionRepository はWebhook購読リポジトリインターフェース
type SubscriptionRepository interface {
	Save(ctx context.Context, sub *Subscription) error
	FindByID(ctx context.Context, id string) (*Subscription, error)
	FindAll(ctx context.Context) ([]*Subscription, error)
	FindActiveByEvent(ctx context.Context, event EventType) ([]*Subscription, error)
	Delete(ctx context.Context, id string) error
}

// DeliveryRepository はWebhook配信記録リポジトリインターフェース
type DeliveryRepository interface {
	Save(ctx context.Context, delivery *Delivery) error
	Update(ctx context.Context, delivery *Delivery) error
	FindByID(ctx context.Context, id string) (*Delivery, error)
	FindPendingRetries(ctx context.Context) ([]*Delivery, error)
}
