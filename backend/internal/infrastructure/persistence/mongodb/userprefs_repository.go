package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/kuro48/idol-api/internal/domain/userprefs"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UserPrefsRepository は MongoDB を使ったユーザー設定リポジトリ
type UserPrefsRepository struct {
	collection *mongo.Collection
}

// NewUserPrefsRepository はリポジトリを作成する
func NewUserPrefsRepository(db *mongo.Database) *UserPrefsRepository {
	return &UserPrefsRepository{collection: db.Collection("user_preferences")}
}

type userPrefsDocument struct {
	Sub       string    `bson:"sub"`
	OshiColor string    `bson:"oshi_color,omitempty"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// EnsureIndexes はインデックスを作成する
func (r *UserPrefsRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sub", Value: 1}},
		Options: options.Index().SetName("idx_userprefs_sub").SetUnique(true),
	})
	return err
}

// FindBySub は sub でユーザー設定を取得する
func (r *UserPrefsRepository) FindBySub(ctx context.Context, sub string) (*domain.UserPrefs, error) {
	var doc userPrefsDocument
	err := r.collection.FindOne(ctx, bson.M{"sub": sub}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ユーザー設定の取得に失敗しました: %w", err)
	}
	return domain.Reconstruct(doc.Sub, doc.OshiColor, doc.UpdatedAt)
}

// Upsert はユーザー設定を作成または更新する
func (r *UserPrefsRepository) Upsert(ctx context.Context, prefs *domain.UserPrefs) error {
	filter := bson.M{"sub": prefs.Sub()}
	update := bson.M{"$set": bson.M{
		"sub":        prefs.Sub(),
		"oshi_color": prefs.OshiColor(),
		"updated_at": prefs.UpdatedAt(),
	}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("ユーザー設定の保存に失敗しました: %w", err)
	}
	return nil
}
