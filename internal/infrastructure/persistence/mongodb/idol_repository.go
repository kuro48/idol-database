package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/idol"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// IdolRepository はMongoDBを使用したアイドルリポジトリの実装
type IdolRepository struct {
	collection *mongo.Collection
}

// NewIdolRepository はMongoDBアイドルリポジトリを作成する
func NewIdolRepository(db *mongo.Database) *IdolRepository {
	return &IdolRepository{
		collection: db.Collection("idols"),
	}
}

// idolDocument はMongoDBに保存するドキュメント構造
type idolDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Birthdate   time.Time     `bson:"birthdate"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}

// toDocument はドメインモデルをMongoDBドキュメントに変換する
func toIdolDocument(i *idol.Idol) *idolDocument {
	// IDの文字列をObjectIDに変換
	objectID, _ := bson.ObjectIDFromHex(i.ID().Value())

	return &idolDocument{
		ID:          objectID,
		Name:        i.Name().Value(),
		Birthdate:   i.Birthdate().Value(),
		CreatedAt:   i.CreatedAt(),
		UpdatedAt:   i.UpdatedAt(),
	}
}

// toDomain はMongoDBドキュメントをドメインモデルに変換する
func toDomain(doc *idolDocument) (*idol.Idol, error) {
	id, err := idol.NewIdolID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	name, err := idol.NewIdolName(doc.Name)
	if err != nil {
		return nil, err
	}

	// time.Timeから年月日を抽出してBirthdateを作成
	year, month, day := doc.Birthdate.Date()
	birthdate, err := idol.NewBirthdate(year, int(month), day)
	if err != nil {
		return nil, err
	}


	return idol.Reconstruct(id, name, &birthdate, doc.CreatedAt, doc.UpdatedAt), nil
}

// Save は新しいアイドルを保存する
func (r *IdolRepository) Save(ctx context.Context, i *idol.Idol) error {
	doc := toIdolDocument(i)

	// 新規作成の場合はIDを生成
	if doc.ID.IsZero() {
		doc.ID = bson.NewObjectID()
		doc.CreatedAt = time.Now()
		doc.UpdatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("アイドルの保存エラー: %w", err)
	}

	return nil
}

// FindByID はIDでアイドルを検索する
func (r *IdolRepository) FindByID(ctx context.Context, id idol.IdolID) (*idol.Idol, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc idolDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("アイドルが見つかりません")
		}
		return nil, fmt.Errorf("アイドル取得エラー: %w", err)
	}

	return toDomain(&doc)
}

// FindAll は全てのアイドルを取得する
func (r *IdolRepository) FindAll(ctx context.Context) ([]*idol.Idol, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("アイドル一覧取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []idolDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	idols := make([]*idol.Idol, 0, len(docs))
	for _, doc := range docs {
		i, err := toDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		idols = append(idols, i)
	}

	return idols, nil
}

// Update は既存のアイドルを更新する
func (r *IdolRepository) Update(ctx context.Context, i *idol.Idol) error {
	objectID, err := bson.ObjectIDFromHex(i.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	doc := toIdolDocument(i)
	doc.UpdatedAt = time.Now()

	updateDoc := bson.M{
		"$set": bson.M{
			"name":        doc.Name,
			"birthdate":   doc.Birthdate,
			"updated_at":  doc.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		return fmt.Errorf("アイドル更新エラー: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("アイドルが見つかりません")
	}

	return nil
}

// Delete はアイドルを削除する
func (r *IdolRepository) Delete(ctx context.Context, id idol.IdolID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("アイドル削除エラー: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("アイドルが見つかりません")
	}

	return nil
}

// ExistsByName は同じ名前のアイドルが存在するかチェック
func (r *IdolRepository) ExistsByName(ctx context.Context, name idol.IdolName) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"name": name.Value()})
	if err != nil {
		return false, fmt.Errorf("名前チェックエラー: %w", err)
	}

	return count > 0, nil
}
