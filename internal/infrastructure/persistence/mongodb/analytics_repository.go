package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/analytics"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AnalyticsRepository はMongoDBを使用したAPI利用記録リポジトリの実装
type AnalyticsRepository struct {
	collection *mongo.Collection
}

// NewAnalyticsRepository はAnalyticsRepositoryを作成する
func NewAnalyticsRepository(db *mongo.Database) *AnalyticsRepository {
	return &AnalyticsRepository{
		collection: db.Collection("api_usage_logs"),
	}
}

// usageDocument はMongoDBに保存するドキュメント構造
type usageDocument struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	MaskedKey  string        `bson:"masked_key"`
	Endpoint   string        `bson:"endpoint"`
	Method     string        `bson:"method"`
	StatusCode int           `bson:"status_code"`
	LatencyMs  int64         `bson:"latency_ms"`
	RecordedAt time.Time     `bson:"recorded_at"`
}

// Save はAPI利用記録を保存する
func (r *AnalyticsRepository) Save(ctx context.Context, record *analytics.APIUsageRecord) error {
	doc := usageDocument{
		ID:         bson.NewObjectID(),
		MaskedKey:  record.MaskedKey,
		Endpoint:   record.Endpoint,
		Method:     record.Method,
		StatusCode: record.StatusCode,
		LatencyMs:  record.LatencyMs,
		RecordedAt: record.RecordedAt,
	}

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("API利用記録の保存エラー: %w", err)
	}

	return nil
}

// FindByMaskedKey はマスク済みキーで利用記録を取得する
func (r *AnalyticsRepository) FindByMaskedKey(ctx context.Context, maskedKey string, from, to time.Time) ([]*analytics.APIUsageRecord, error) {
	filter := bson.M{
		"masked_key": maskedKey,
		"recorded_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("API利用記録の取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []usageDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	records := make([]*analytics.APIUsageRecord, 0, len(docs))
	for _, doc := range docs {
		records = append(records, &analytics.APIUsageRecord{
			ID:         doc.ID.Hex(),
			MaskedKey:  doc.MaskedKey,
			Endpoint:   doc.Endpoint,
			Method:     doc.Method,
			StatusCode: doc.StatusCode,
			LatencyMs:  doc.LatencyMs,
			RecordedAt: doc.RecordedAt,
		})
	}

	return records, nil
}

// AggregateByKey はAPIキー単位で利用統計を集計する
func (r *AnalyticsRepository) AggregateByKey(ctx context.Context, from, to time.Time) ([]*analytics.KeyUsageSummary, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"recorded_at": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}},
		bson.M{"$group": bson.M{
			"_id":            "$masked_key",
			"total_requests": bson.M{"$sum": 1},
			"success_count": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$lt": bson.A{"$status_code", 400}},
					1,
					0,
				},
			}},
			"error_count": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$gte": bson.A{"$status_code", 400}},
					1,
					0,
				},
			}},
			"avg_latency_ms": bson.M{"$avg": "$latency_ms"},
			"last_used_at":   bson.M{"$max": "$recorded_at"},
		}},
		bson.M{"$sort": bson.M{"total_requests": -1}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("API利用集計エラー: %w", err)
	}
	defer cursor.Close(ctx)

	type aggregateResult struct {
		MaskedKey     string    `bson:"_id"`
		TotalRequests int64     `bson:"total_requests"`
		SuccessCount  int64     `bson:"success_count"`
		ErrorCount    int64     `bson:"error_count"`
		AvgLatencyMs  float64   `bson:"avg_latency_ms"`
		LastUsedAt    time.Time `bson:"last_used_at"`
	}

	var results []aggregateResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("集計データ変換エラー: %w", err)
	}

	summaries := make([]*analytics.KeyUsageSummary, 0, len(results))
	for _, r := range results {
		summaries = append(summaries, &analytics.KeyUsageSummary{
			MaskedKey:     r.MaskedKey,
			TotalRequests: r.TotalRequests,
			SuccessCount:  r.SuccessCount,
			ErrorCount:    r.ErrorCount,
			AvgLatencyMs:  r.AvgLatencyMs,
			LastUsedAt:    r.LastUsedAt,
		})
	}

	return summaries, nil
}

// EnsureIndexes はインデックスを作成する
func (r *AnalyticsRepository) EnsureIndexes(ctx context.Context) error {
	// TTLインデックス: 90日後に自動削除
	ttlDuration := int32(90 * 24 * 60 * 60) // 90日 in seconds
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "recorded_at", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(ttlDuration).
				SetName("ttl_recorded_at"),
		},
		{
			Keys: bson.D{
				{Key: "masked_key", Value: 1},
				{Key: "recorded_at", Value: -1},
			},
			Options: options.Index().SetName("idx_masked_key_recorded_at"),
		},
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("analyticsインデックス作成エラー: %w", err)
	}

	return nil
}
