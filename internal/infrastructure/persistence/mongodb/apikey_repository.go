package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	"github.com/kuro48/idol-api/internal/domain/plan"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// APIKeyRepository はMongoDBを使用したAPIキーリポジトリの実装
type APIKeyRepository struct {
	collection *mongo.Collection
}

// NewAPIKeyRepository はMongoDBのAPIキーリポジトリを作成する
func NewAPIKeyRepository(db *mongo.Database) *APIKeyRepository {
	return &APIKeyRepository{
		collection: db.Collection("api_keys"),
	}
}

// apikeyDocument はMongoDBに保存するドキュメント構造
type apikeyDocument struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Prefix    string        `bson:"prefix"`
	KeyHash   string        `bson:"key_hash"`
	MaskedKey string        `bson:"masked_key"`
	Email     string        `bson:"email"`
	Name      string        `bson:"name"`
	PlanType  string        `bson:"plan_type"`
	IsActive  bool          `bson:"is_active"`
	CreatedAt time.Time     `bson:"created_at"`
	OshiColor string        `bson:"oshi_color,omitempty"`
}

// EnsureIndexes はコレクションのインデックスを作成する
func (r *APIKeyRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "prefix", Value: 1}},
			Options: options.Index().SetName("idx_apikey_prefix"),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetName("idx_apikey_email"),
		},
		{
			Keys:    bson.D{{Key: "key_hash", Value: 1}},
			Options: options.Index().SetName("idx_apikey_hash").SetUnique(true),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Save は新しいAPIキーを保存する
func (r *APIKeyRepository) Save(ctx context.Context, key *domainapikey.APIKey) error {
	objectID, err := bson.ObjectIDFromHex(key.ID())
	if err != nil {
		return fmt.Errorf("無効なAPIキーID: %w", err)
	}

	doc := apikeyDocument{
		ID:        objectID,
		Prefix:    key.Prefix(),
		KeyHash:   key.KeyHash(),
		MaskedKey: key.MaskedKey(),
		Email:     key.Email(),
		Name:      key.Name(),
		PlanType:  string(key.PlanType()),
		IsActive:  key.IsActive(),
		CreatedAt: key.CreatedAt(),
		OshiColor: key.OshiColor(),
	}
	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("APIキーの保存に失敗しました: %w", err)
	}
	return nil
}

// FindByPrefix はプレフィックスでアクティブなAPIキーを取得する
func (r *APIKeyRepository) FindByPrefix(ctx context.Context, prefix string) ([]*domainapikey.APIKey, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"prefix": prefix, "is_active": true})
	if err != nil {
		return nil, fmt.Errorf("APIキーの検索に失敗しました: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []*domainapikey.APIKey
	for cursor.Next(ctx) {
		var doc apikeyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("APIキーのデコードに失敗しました: %w", err)
		}
		key, err := toAPIKeyDomain(&doc)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, cursor.Err()
}

// FindByID はIDでAPIキーを取得する
func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*domainapikey.APIKey, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("無効なAPIキーID: %w", err)
	}

	var doc apikeyDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("APIキーの取得に失敗しました: %w", err)
	}
	return toAPIKeyDomain(&doc)
}

// FindByEmail はメールアドレスで全APIキーを取得する
func (r *APIKeyRepository) FindByEmail(ctx context.Context, email string) ([]*domainapikey.APIKey, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"email": email})
	if err != nil {
		return nil, fmt.Errorf("APIキーの検索に失敗しました: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []*domainapikey.APIKey
	for cursor.Next(ctx) {
		var doc apikeyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("APIキーのデコードに失敗しました: %w", err)
		}
		key, err := toAPIKeyDomain(&doc)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, cursor.Err()
}

// Update はAPIキーを更新する
func (r *APIKeyRepository) Update(ctx context.Context, key *domainapikey.APIKey) error {
	objectID, err := bson.ObjectIDFromHex(key.ID())
	if err != nil {
		return fmt.Errorf("無効なAPIキーID: %w", err)
	}

	update := bson.M{"$set": bson.M{
		"is_active":  key.IsActive(),
		"plan_type":  string(key.PlanType()),
		"oshi_color": key.OshiColor(),
	}}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("APIキーの更新に失敗しました: %w", err)
	}
	return nil
}

func toAPIKeyDomain(doc *apikeyDocument) (*domainapikey.APIKey, error) {
	key, err := domainapikey.Reconstruct(
		doc.ID.Hex(),
		doc.Prefix,
		doc.KeyHash,
		doc.MaskedKey,
		doc.Email,
		doc.Name,
		plan.Type(doc.PlanType),
		doc.IsActive,
		doc.CreatedAt,
		doc.OshiColor,
	)
	if err != nil {
		return nil, fmt.Errorf("APIキードメインモデルの再構築に失敗しました: %w", err)
	}
	return key, nil
}
