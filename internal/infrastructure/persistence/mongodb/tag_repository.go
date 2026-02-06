package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/tag"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TagRepository はMongoDBを使用したタグリポジトリの実装
type TagRepository struct {
	collection *mongo.Collection
}

// NewTagRepository はMongoDBタグリポジトリを作成する
func NewTagRepository(db *mongo.Database) *TagRepository {
	return &TagRepository{
		collection: db.Collection("tags"),
	}
}

// tagDocument はMongoDBに保存するドキュメント構造
type tagDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Category    string        `bson:"category"`
	Description string        `bson:"description"`
	CreatedAt   time.Time     `bson:"created_at"`
}

// toTagDocument はドメインモデルをMongoDBドキュメントに変換する
func toTagDocument(t *tag.Tag) (*tagDocument, error) {
	objectID, err := bson.ObjectIDFromHex(t.ID().String())
	if err != nil {
		return nil, fmt.Errorf("無効なタグID: %w", err)
	}

	return &tagDocument{
		ID:          objectID,
		Name:        t.Name().String(),
		Category:    t.Category().String(),
		Description: t.Description(),
		CreatedAt:   t.CreatedAt(),
	}, nil
}

// toTagDomain はMongoDBドキュメントをドメインモデルに変換する
func toTagDomain(doc *tagDocument) (*tag.Tag, error) {
	return tag.Reconstruct(
		doc.ID.Hex(),
		doc.Name,
		doc.Category,
		doc.Description,
		doc.CreatedAt,
	)
}

// Save はタグを保存する
func (r *TagRepository) Save(ctx context.Context, t *tag.Tag) error {
	doc, err := toTagDocument(t)
	if err != nil {
		return err
	}

	// ドメイン層で生成されたIDが必須
	if doc.ID.IsZero() {
		return errors.New("タグIDが設定されていません")
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("同じ名前のタグが既に存在します")
		}
		return fmt.Errorf("タグの保存エラー: %w", err)
	}

	return nil
}

// Update はタグを更新する
func (r *TagRepository) Update(ctx context.Context, t *tag.Tag) error {
	doc, err := toTagDocument(t)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": doc.ID}
	update := bson.M{
		"$set": bson.M{
			"name":        doc.Name,
			"category":    doc.Category,
			"description": doc.Description,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("タグの更新エラー: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("タグが見つかりません")
	}

	return nil
}

// Delete はタグを削除する
func (r *TagRepository) Delete(ctx context.Context, id tag.TagID) error {
	objectID, err := bson.ObjectIDFromHex(id.String())
	if err != nil {
		return fmt.Errorf("無効なタグID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("タグの削除エラー: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("タグが見つかりません")
	}

	return nil
}

// FindByID はIDでタグを検索する
func (r *TagRepository) FindByID(ctx context.Context, id tag.TagID) (*tag.Tag, error) {
	objectID, err := bson.ObjectIDFromHex(id.String())
	if err != nil {
		return nil, fmt.Errorf("無効なタグID: %w", err)
	}

	var doc tagDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("タグが見つかりません")
		}
		return nil, fmt.Errorf("タグの取得エラー: %w", err)
	}

	return toTagDomain(&doc)
}

// FindByName は名前でタグを検索する（完全一致）
func (r *TagRepository) FindByName(ctx context.Context, name string) (*tag.Tag, error) {
	var doc tagDocument
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("タグが見つかりません")
		}
		return nil, fmt.Errorf("タグの取得エラー: %w", err)
	}

	return toTagDomain(&doc)
}

// FindByCategory はカテゴリでタグを検索する
func (r *TagRepository) FindByCategory(ctx context.Context, category tag.TagCategory) ([]*tag.Tag, error) {
	filter := bson.M{"category": category.String()}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("タグの検索エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var tags []*tag.Tag
	for cursor.Next(ctx) {
		var doc tagDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("タグのデコードエラー: %w", err)
		}

		t, err := toTagDomain(&doc)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("カーソルエラー: %w", err)
	}

	return tags, nil
}

// Search は検索条件に基づいてタグを検索する
func (r *TagRepository) Search(ctx context.Context, criteria tag.SearchCriteria) ([]*tag.Tag, int64, error) {
	filter := bson.M{}

	// 名前による部分一致検索
	if criteria.Name != nil && *criteria.Name != "" {
		filter["name"] = bson.M{"$regex": *criteria.Name, "$options": "i"}
	}

	// カテゴリフィルタ
	if criteria.Category != nil {
		filter["category"] = criteria.Category.String()
	}

	// 総件数を取得
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("件数取得エラー: %w", err)
	}

	// ページネーション設定
	page := criteria.Page
	if page < 1 {
		page = 1
	}
	limit := criteria.Limit
	if limit < 1 {
		limit = 20
	}

	skip := int64((page - 1) * limit)

	// 検索実行
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("タグの検索エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var tags []*tag.Tag
	for cursor.Next(ctx) {
		var doc tagDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, fmt.Errorf("タグのデコードエラー: %w", err)
		}

		t, err := toTagDomain(&doc)
		if err != nil {
			return nil, 0, err
		}
		tags = append(tags, t)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, fmt.Errorf("カーソルエラー: %w", err)
	}

	return tags, total, nil
}

// Exists はタグが存在するか確認する
func (r *TagRepository) Exists(ctx context.Context, id tag.TagID) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(id.String())
	if err != nil {
		return false, fmt.Errorf("無効なタグID: %w", err)
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": objectID})
	if err != nil {
		return false, fmt.Errorf("存在確認エラー: %w", err)
	}

	return count > 0, nil
}

// EnsureIndexes はMongoDBのインデックスを作成する
func (r *TagRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}

	return nil
}
