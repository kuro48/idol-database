package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainbilling "github.com/kuro48/idol-api/internal/domain/billing"
	"github.com/kuro48/idol-api/internal/domain/plan"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BillingFulfillmentRepository は MongoDB を使った Checkout fulfillment のリポジトリ実装。
type BillingFulfillmentRepository struct {
	collection *mongo.Collection
}

// NewBillingFulfillmentRepository はリポジトリを作成する。
func NewBillingFulfillmentRepository(db *mongo.Database) *BillingFulfillmentRepository {
	return &BillingFulfillmentRepository{
		collection: db.Collection("billing_fulfillments"),
	}
}

type billingFulfillmentDocument struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	SessionID  string        `bson:"session_id"`
	CustomerID string        `bson:"customer_id"`
	Email      string        `bson:"email"`
	Name       string        `bson:"name"`
	PlanType   string        `bson:"plan_type"`
	APIKeyID   string        `bson:"api_key_id"`
	NotifiedAt *time.Time    `bson:"notified_at,omitempty"`
	CreatedAt  time.Time     `bson:"created_at"`
}

// EnsureIndexes はコレクションインデックスを作成する。
func (r *BillingFulfillmentRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetName("idx_billing_session_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_billing_email_created_at"),
		},
		{
			Keys:    bson.D{{Key: "customer_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_billing_customer_created_at"),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Save は fulfillment を保存する。
func (r *BillingFulfillmentRepository) Save(ctx context.Context, fulfillment *domainbilling.CheckoutFulfillment) error {
	_, err := r.collection.InsertOne(ctx, toBillingFulfillmentDocument(fulfillment))
	if err != nil {
		return fmt.Errorf("billing fulfillment の保存に失敗しました: %w", err)
	}
	return nil
}

// FindBySessionID は session ID から fulfillment を取得する。
func (r *BillingFulfillmentRepository) FindBySessionID(ctx context.Context, sessionID string) (*domainbilling.CheckoutFulfillment, error) {
	var doc billingFulfillmentDocument
	err := r.collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("billing fulfillment の取得に失敗しました: %w", err)
	}
	return toBillingFulfillmentDomain(&doc)
}

// FindLatestByEmail はメールアドレスに紐づく最新 fulfillment を取得する。
func (r *BillingFulfillmentRepository) FindLatestByEmail(ctx context.Context, email string) (*domainbilling.CheckoutFulfillment, error) {
	var doc billingFulfillmentDocument
	err := r.collection.FindOne(
		ctx,
		bson.M{"email": email},
		options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("billing fulfillment の取得に失敗しました: %w", err)
	}
	return toBillingFulfillmentDomain(&doc)
}

// FindLatestByCustomerID は customer ID に紐づく最新 fulfillment を取得する。
func (r *BillingFulfillmentRepository) FindLatestByCustomerID(ctx context.Context, customerID string) (*domainbilling.CheckoutFulfillment, error) {
	var doc billingFulfillmentDocument
	err := r.collection.FindOne(
		ctx,
		bson.M{"customer_id": customerID},
		options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("billing fulfillment の取得に失敗しました: %w", err)
	}
	return toBillingFulfillmentDomain(&doc)
}

// Update は fulfillment を更新する。
func (r *BillingFulfillmentRepository) Update(ctx context.Context, fulfillment *domainbilling.CheckoutFulfillment) error {
	update := bson.M{
		"$set": bson.M{
			"customer_id": fulfillment.CustomerID(),
			"email":       fulfillment.Email(),
			"name":        fulfillment.Name(),
			"plan_type":   string(fulfillment.PlanType()),
			"api_key_id":  fulfillment.APIKeyID(),
			"notified_at": fulfillment.NotifiedAt(),
		},
	}
	if _, err := r.collection.UpdateOne(ctx, bson.M{"session_id": fulfillment.SessionID()}, update); err != nil {
		return fmt.Errorf("billing fulfillment の更新に失敗しました: %w", err)
	}
	return nil
}

func toBillingFulfillmentDocument(fulfillment *domainbilling.CheckoutFulfillment) billingFulfillmentDocument {
	return billingFulfillmentDocument{
		SessionID:  fulfillment.SessionID(),
		CustomerID: fulfillment.CustomerID(),
		Email:      fulfillment.Email(),
		Name:       fulfillment.Name(),
		PlanType:   string(fulfillment.PlanType()),
		APIKeyID:   fulfillment.APIKeyID(),
		NotifiedAt: fulfillment.NotifiedAt(),
		CreatedAt:  fulfillment.CreatedAt(),
	}
}

func toBillingFulfillmentDomain(doc *billingFulfillmentDocument) (*domainbilling.CheckoutFulfillment, error) {
	planType := plan.Type(doc.PlanType)
	if !plan.IsValid(planType) {
		return nil, fmt.Errorf("無効なプラン種別です: %s", doc.PlanType)
	}
	return domainbilling.ReconstructCheckoutFulfillment(
		doc.SessionID,
		doc.CustomerID,
		doc.Email,
		doc.Name,
		planType,
		doc.APIKeyID,
		doc.NotifiedAt,
		doc.CreatedAt,
	), nil
}
