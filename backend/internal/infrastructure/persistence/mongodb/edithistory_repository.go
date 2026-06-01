package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/edithistory"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ensure options is used
var _ = options.Find

// EditHistoryRepository はMongoDBを使用した編集履歴リポジトリ
type EditHistoryRepository struct {
	collection *mongo.Collection
}

// NewEditHistoryRepository は編集履歴リポジトリを作成する
func NewEditHistoryRepository(db *mongo.Database) *EditHistoryRepository {
	return &EditHistoryRepository{
		collection: db.Collection("edit_history"),
	}
}

type fieldChangeDocument struct {
	Before interface{} `bson:"before"`
	After  interface{} `bson:"after"`
}

type editHistoryDocument struct {
	ID         bson.ObjectID                  `bson:"_id,omitempty"`
	EntityType string                         `bson:"entity_type"`
	EntityID   string                         `bson:"entity_id"`
	Action     string                         `bson:"action"`
	Changes    map[string]fieldChangeDocument `bson:"changes"`
	ChangedBy  string                         `bson:"changed_by"`
	CreatedAt  time.Time                      `bson:"created_at"`
}

func (r *EditHistoryRepository) Save(ctx context.Context, h *edithistory.EditHistory) error {
	doc := toEditHistoryDocument(h)

	oid, err := bson.ObjectIDFromHex(h.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}
	doc.ID = oid

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("編集履歴の保存に失敗: %w", err)
	}
	return nil
}

func (r *EditHistoryRepository) FindByID(ctx context.Context, id edithistory.EditHistoryID) (*edithistory.EditHistory, error) {
	oid, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc editHistoryDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("編集履歴が見つかりません: %s", id.Value())
		}
		return nil, fmt.Errorf("編集履歴の取得に失敗: %w", err)
	}
	return fromEditHistoryDocument(doc), nil
}

func (r *EditHistoryRepository) Search(ctx context.Context, criteria edithistory.SearchCriteria) ([]*edithistory.EditHistory, error) {
	filter := buildEditHistoryFilter(criteria)
	opts := options.Find().
		SetSkip(int64(criteria.Offset)).
		SetLimit(int64(criteria.Limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("編集履歴の検索に失敗: %w", err)
	}
	defer cursor.Close(ctx)

	return scanEditHistoryCursor(ctx, cursor)
}

func (r *EditHistoryRepository) Count(ctx context.Context, criteria edithistory.SearchCriteria) (int64, error) {
	filter := buildEditHistoryFilter(criteria)
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("件数取得に失敗: %w", err)
	}
	return count, nil
}

// EnsureIndexes はインデックスを作成する
func (r *EditHistoryRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "entity_type", Value: 1}, {Key: "entity_id", Value: 1}}},
		{Keys: bson.D{{Key: "changed_by", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}
	return nil
}

func buildEditHistoryFilter(criteria edithistory.SearchCriteria) bson.M {
	filter := bson.M{}
	if criteria.EntityType != nil {
		filter["entity_type"] = criteria.EntityType.Value()
	}
	if criteria.EntityID != nil {
		filter["entity_id"] = *criteria.EntityID
	}
	if criteria.Action != nil {
		filter["action"] = criteria.Action.Value()
	}
	if criteria.ChangedBy != nil {
		filter["changed_by"] = *criteria.ChangedBy
	}
	return filter
}

func toEditHistoryDocument(h *edithistory.EditHistory) editHistoryDocument {
	changes := make(map[string]fieldChangeDocument, len(h.Changes()))
	for field, fc := range h.Changes() {
		changes[field] = fieldChangeDocument{Before: fc.Before, After: fc.After}
	}
	return editHistoryDocument{
		EntityType: h.EntityType().Value(),
		EntityID:   h.EntityID(),
		Action:     h.Action().Value(),
		Changes:    changes,
		ChangedBy:  h.ChangedBy(),
		CreatedAt:  h.CreatedAt(),
	}
}

func fromEditHistoryDocument(doc editHistoryDocument) *edithistory.EditHistory {
	id, _ := edithistory.NewEditHistoryID(doc.ID.Hex())
	entityType, _ := edithistory.NewEntityType(doc.EntityType)
	action, _ := edithistory.NewAction(doc.Action)

	changes := make(map[string]edithistory.FieldChange, len(doc.Changes))
	for field, fc := range doc.Changes {
		changes[field] = edithistory.FieldChange{Before: fc.Before, After: fc.After}
	}

	return edithistory.Reconstruct(id, entityType, doc.EntityID, action, changes, doc.ChangedBy, doc.CreatedAt)
}

func scanEditHistoryCursor(ctx context.Context, cursor *mongo.Cursor) ([]*edithistory.EditHistory, error) {
	var results []*edithistory.EditHistory
	for cursor.Next(ctx) {
		var doc editHistoryDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("デコードエラー: %w", err)
		}
		results = append(results, fromEditHistoryDocument(doc))
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("カーソルエラー: %w", err)
	}
	return results, nil
}
