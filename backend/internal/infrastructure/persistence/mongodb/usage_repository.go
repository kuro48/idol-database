package mongodb

import (
	"context"
	"fmt"
	"time"

	domainusage "github.com/kuro48/idol-api/internal/domain/usage"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UsageRepository はMongoDBを使用した月次使用量リポジトリの実装
type UsageRepository struct {
	collection *mongo.Collection
}

// NewUsageRepository はMongoDBの月次使用量リポジトリを作成する
func NewUsageRepository(db *mongo.Database) *UsageRepository {
	return &UsageRepository{
		collection: db.Collection("api_key_usage"),
	}
}

// apiKeyUsageDocument はMongoDBに保存するドキュメント構造
type apiKeyUsageDocument struct {
	ID        string    `bson:"_id"`        // "{key_prefix}_{year_month}" の複合キー
	KeyPrefix string    `bson:"key_prefix"`
	YearMonth string    `bson:"year_month"`
	Count     int       `bson:"count"`
	Limit     int       `bson:"limit"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// EnsureIndexes はコレクションのインデックスを作成する
func (r *UsageRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "key_prefix", Value: 1}, {Key: "year_month", Value: 1}},
			Options: options.Index().SetName("idx_usage_prefix_month"),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// IncrementAndGet は使用量を1増やし、最新の MonthlyUsage を返す
// ドキュメントが存在しない場合は upsert で新規作成する
func (r *UsageRepository) IncrementAndGet(ctx context.Context, keyPrefix, yearMonth string, limit int) (*domainusage.MonthlyUsage, error) {
	docID := keyPrefix + "_" + yearMonth

	filter := bson.M{"_id": docID}
	update := bson.M{
		"$inc": bson.M{"count": 1},
		"$set": bson.M{
			"key_prefix": keyPrefix,
			"year_month": yearMonth,
			"limit":      limit,
			"updated_at": time.Now().UTC(),
		},
		"$setOnInsert": bson.M{"_id": docID},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var doc apiKeyUsageDocument
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		return nil, fmt.Errorf("使用量のインクリメントに失敗しました: %w", err)
	}

	return domainusage.Reconstruct(doc.KeyPrefix, doc.YearMonth, doc.Count, doc.Limit, doc.UpdatedAt), nil
}

// Get は使用量を取得する（インクリメントなし）
// ドキュメントが存在しない場合は count=0 の MonthlyUsage を返す
func (r *UsageRepository) Get(ctx context.Context, keyPrefix, yearMonth string, limit int) (*domainusage.MonthlyUsage, error) {
	docID := keyPrefix + "_" + yearMonth

	var doc apiKeyUsageDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domainusage.New(keyPrefix, yearMonth, limit), nil
		}
		return nil, fmt.Errorf("使用量の取得に失敗しました: %w", err)
	}

	return domainusage.Reconstruct(doc.KeyPrefix, doc.YearMonth, doc.Count, doc.Limit, doc.UpdatedAt), nil
}
