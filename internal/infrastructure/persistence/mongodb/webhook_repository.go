package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/webhook"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ---- Subscription Repository ----

// WebhookSubscriptionRepository はMongoDBを使用したWebhook購読リポジトリ
type WebhookSubscriptionRepository struct {
	collection *mongo.Collection
}

// NewWebhookSubscriptionRepository はリポジトリを作成する
func NewWebhookSubscriptionRepository(db *mongo.Database) *WebhookSubscriptionRepository {
	return &WebhookSubscriptionRepository{
		collection: db.Collection("webhook_subscriptions"),
	}
}

type subscriptionDocument struct {
	ID        string    `bson:"_id"`
	URL       string    `bson:"url"`
	Secret    string    `bson:"secret"`
	Events    []string  `bson:"events"`
	Active    bool      `bson:"active"`
	CreatedAt time.Time `bson:"created_at"`
	CreatedBy string    `bson:"created_by,omitempty"`
}

func (r *WebhookSubscriptionRepository) Save(ctx context.Context, sub *webhook.Subscription) error {
	doc := &subscriptionDocument{
		ID:        sub.ID(),
		URL:       sub.URL(),
		Secret:    sub.Secret(),
		Events:    eventsToStrings(sub.Events()),
		Active:    sub.Active(),
		CreatedAt: sub.CreatedAt(),
		CreatedBy: sub.CreatedBy(),
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *WebhookSubscriptionRepository) FindByID(ctx context.Context, id string) (*webhook.Subscription, error) {
	var doc subscriptionDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Webhook購読が見つかりません")
		}
		return nil, err
	}
	return docToSubscription(&doc), nil
}

func (r *WebhookSubscriptionRepository) FindAll(ctx context.Context) ([]*webhook.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []subscriptionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	subs := make([]*webhook.Subscription, 0, len(docs))
	for _, doc := range docs {
		subs = append(subs, docToSubscription(&doc))
	}
	return subs, nil
}

func (r *WebhookSubscriptionRepository) FindActiveByEvent(ctx context.Context, event webhook.EventType) ([]*webhook.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"active": true,
		"events": string(event),
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []subscriptionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	subs := make([]*webhook.Subscription, 0, len(docs))
	for _, doc := range docs {
		subs = append(subs, docToSubscription(&doc))
	}
	return subs, nil
}

func (r *WebhookSubscriptionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("Webhook購読が見つかりません")
	}
	return nil
}

func docToSubscription(doc *subscriptionDocument) *webhook.Subscription {
	events := stringsToEvents(doc.Events)
	return webhook.NewSubscription(doc.ID, doc.URL, doc.Secret, events, doc.CreatedBy)
}

func eventsToStrings(events []webhook.EventType) []string {
	ss := make([]string, len(events))
	for i, e := range events {
		ss[i] = string(e)
	}
	return ss
}

func stringsToEvents(ss []string) []webhook.EventType {
	events := make([]webhook.EventType, len(ss))
	for i, s := range ss {
		events[i] = webhook.EventType(s)
	}
	return events
}

// ---- Delivery Repository ----

// WebhookDeliveryRepository はMongoDBを使用したWebhook配信記録リポジトリ
type WebhookDeliveryRepository struct {
	collection *mongo.Collection
}

// NewWebhookDeliveryRepository はリポジトリを作成する
func NewWebhookDeliveryRepository(db *mongo.Database) *WebhookDeliveryRepository {
	return &WebhookDeliveryRepository{
		collection: db.Collection("webhook_deliveries"),
	}
}

type deliveryDocument struct {
	ID             string     `bson:"_id"`
	SubscriptionID string     `bson:"subscription_id"`
	Event          string     `bson:"event"`
	Payload        []byte     `bson:"payload"`
	Status         string     `bson:"status"`
	Attempts       int        `bson:"attempts"`
	MaxAttempts    int        `bson:"max_attempts"`
	LastAttemptAt  *time.Time `bson:"last_attempt_at,omitempty"`
	NextRetryAt    *time.Time `bson:"next_retry_at,omitempty"`
	ResponseCode   *int       `bson:"response_code,omitempty"`
	ErrorMessage   string     `bson:"error_message,omitempty"`
	CreatedAt      time.Time  `bson:"created_at"`
}

func (r *WebhookDeliveryRepository) Save(ctx context.Context, delivery *webhook.Delivery) error {
	doc := toDeliveryDocument(delivery)
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *WebhookDeliveryRepository) Update(ctx context.Context, delivery *webhook.Delivery) error {
	doc := toDeliveryDocument(delivery)
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": delivery.ID()}, doc)
	return err
}

func (r *WebhookDeliveryRepository) FindByID(ctx context.Context, id string) (*webhook.Delivery, error) {
	var doc deliveryDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Webhook配信記録が見つかりません")
		}
		return nil, err
	}
	return docToDelivery(&doc), nil
}

func (r *WebhookDeliveryRepository) FindPendingRetries(ctx context.Context) ([]*webhook.Delivery, error) {
	now := time.Now()
	cursor, err := r.collection.Find(ctx, bson.M{
		"status":        "failed",
		"next_retry_at": bson.M{"$lte": now},
	}, options.Find().SetSort(bson.D{{Key: "next_retry_at", Value: 1}}).SetLimit(100))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []deliveryDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	deliveries := make([]*webhook.Delivery, 0, len(docs))
	for _, doc := range docs {
		deliveries = append(deliveries, docToDelivery(&doc))
	}
	return deliveries, nil
}

func toDeliveryDocument(d *webhook.Delivery) *deliveryDocument {
	return &deliveryDocument{
		ID:             d.ID(),
		SubscriptionID: d.SubscriptionID(),
		Event:          string(d.Event()),
		Payload:        d.Payload(),
		Status:         string(d.Status()),
		Attempts:       d.Attempts(),
		MaxAttempts:    d.MaxAttempts(),
		LastAttemptAt:  d.LastAttemptAt(),
		NextRetryAt:    d.NextRetryAt(),
		ResponseCode:   d.ResponseCode(),
		ErrorMessage:   d.ErrorMessage(),
		CreatedAt:      d.CreatedAt(),
	}
}

func docToDelivery(doc *deliveryDocument) *webhook.Delivery {
	d := webhook.NewDelivery(doc.ID, doc.SubscriptionID, webhook.EventType(doc.Event), doc.Payload)
	_ = fmt.Sprintf("restored delivery %s", d.ID()) // suppress unused warning
	return d
}
