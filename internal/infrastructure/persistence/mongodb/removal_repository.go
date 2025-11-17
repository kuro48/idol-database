package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/idol"
	"github.com/kuro48/idol-api/internal/domain/removal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RemovalRepository はMongoDBを使用した削除申請リポジトリの実装
type RemovalRepository struct {
	collection *mongo.Collection
}

// NewRemovalRepository はMongoDB削除申請リポジトリを作成する
func NewRemovalRepository(db *mongo.Database) *RemovalRepository {
	return &RemovalRepository{
		collection: db.Collection("removal_requests"),
	}
}

// removalDocument はMongoDBに保存するドキュメント構造
type removalDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	IdolID      string        `bson:"idol_id"`
	Requester   string        `bson:"requester"`
	Reason      string        `bson:"reason"`
	ContactInfo string        `bson:"contact_info"`
	Evidence    string        `bson:"evidence,omitempty"`
	Description string        `bson:"description"`
	Status      string        `bson:"status"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}

// toRemovalDocument はドメインモデルをMongoDBドキュメントに変換する
func toRemovalDocument(r *removal.RemovalRequest) *removalDocument {
	// IDの文字列をObjectIDに変換（空の場合はゼロ値）
	var objectID bson.ObjectID
	if r.ID().Value() != "" {
		objectID, _ = bson.ObjectIDFromHex(r.ID().Value())
	}

	return &removalDocument{
		ID:          objectID,
		IdolID:      r.IdolID().Value(),
		Requester:   string(r.Requester().Type()),
		Reason:      r.Reason().Value(),
		ContactInfo: r.ContactInfo().Value(),
		Evidence:    r.Evidence().Value(),
		Description: r.Description().Value(),
		Status:      string(r.Status()),
		CreatedAt:   r.CreatedAt(),
		UpdatedAt:   r.UpdatedAt(),
	}
}

// toRemovalDomain はMongoDBドキュメントをドメインモデルに変換する
func toRemovalDomain(doc *removalDocument) (*removal.RemovalRequest, error) {
	id, err := removal.NewRemovalID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	idolID, err := idol.NewIdolID(doc.IdolID)
	if err != nil {
		return nil, err
	}

	requester, err := removal.NewRequester(doc.Requester)
	if err != nil {
		return nil, err
	}

	reason, err := removal.NewRemovalReason(doc.Reason)
	if err != nil {
		return nil, err
	}

	contactInfo, err := removal.NewContactInfo(doc.ContactInfo)
	if err != nil {
		return nil, err
	}

	evidence, err := removal.NewEvidenceURL(doc.Evidence)
	if err != nil {
		return nil, err
	}

	description, err := removal.NewRemovalReason(doc.Description)
	if err != nil {
		return nil, err
	}

	status, err := removal.NewRemovalStatus(doc.Status)
	if err != nil {
		return nil, err
	}

	return removal.Reconstruct(
		id,
		idolID,
		requester,
		reason,
		contactInfo,
		evidence,
		description,
		status,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

// Save は新しい削除申請を保存する
func (r *RemovalRepository) Save(ctx context.Context, request *removal.RemovalRequest) error {
	doc := toRemovalDocument(request)

	// 新規作成の場合はIDを生成
	if doc.ID.IsZero() {
		doc.ID = bson.NewObjectID()
		doc.CreatedAt = time.Now()
		doc.UpdatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("削除申請の保存エラー: %w", err)
	}

	return nil
}

// FindByID はIDで削除申請を検索する
func (r *RemovalRepository) FindByID(ctx context.Context, id removal.RemovalID) (*removal.RemovalRequest, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc removalDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("削除申請が見つかりません")
		}
		return nil, fmt.Errorf("削除申請取得エラー: %w", err)
	}

	return toRemovalDomain(&doc)
}

// FindAll は全ての削除申請を取得する
func (r *RemovalRepository) FindAll(ctx context.Context) ([]*removal.RemovalRequest, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("削除申請一覧取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []removalDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	requests := make([]*removal.RemovalRequest, 0, len(docs))
	for _, doc := range docs {
		request, err := toRemovalDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// FindPending は保留中の削除申請を取得する
func (r *RemovalRepository) FindPending(ctx context.Context) ([]*removal.RemovalRequest, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": "pending"})
	if err != nil {
		return nil, fmt.Errorf("保留中削除申請取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []removalDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	requests := make([]*removal.RemovalRequest, 0, len(docs))
	for _, doc := range docs {
		request, err := toRemovalDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// Update は削除申請を更新する
func (r *RemovalRepository) Update(ctx context.Context, request *removal.RemovalRequest) error {
	objectID, err := bson.ObjectIDFromHex(request.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	doc := toRemovalDocument(request)
	doc.UpdatedAt = time.Now()

	updateDoc := bson.M{
		"$set": bson.M{
			"idol_id":      doc.IdolID,
			"requester":    doc.Requester,
			"reason":       doc.Reason,
			"contact_info": doc.ContactInfo,
			"evidence":     doc.Evidence,
			"description":  doc.Description,
			"status":       doc.Status,
			"updated_at":   doc.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		return fmt.Errorf("削除申請更新エラー: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("削除申請が見つかりません")
	}

	return nil
}

// Delete は削除申請を削除する
func (r *RemovalRepository) Delete(ctx context.Context, id removal.RemovalID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("削除申請削除エラー: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("削除申請が見つかりません")
	}

	return nil
}
