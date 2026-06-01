package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/venue"
	"github.com/kuro48/idol-api/internal/shared/audit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type VenueRepository struct {
	collection *mongo.Collection
}

func NewVenueRepository(db *mongo.Database) *VenueRepository {
	return &VenueRepository{collection: db.Collection("venues")}
}

type venueDocument struct {
	ID          bson.ObjectID    `bson:"_id,omitempty"`
	Name        string           `bson:"name"`
	NameEn      *string          `bson:"name_en,omitempty"`
	Prefecture  *string          `bson:"prefecture,omitempty"`
	City        *string          `bson:"city,omitempty"`
	Address     *string          `bson:"address,omitempty"`
	Capacity    *int             `bson:"capacity,omitempty"`
	OfficialURL *string          `bson:"official_url,omitempty"`
	Sources     []sourceDocument `bson:"sources,omitempty"`
	Version     int              `bson:"version"`
	CreatedAt   time.Time        `bson:"created_at"`
	UpdatedAt   time.Time        `bson:"updated_at"`
	CreatedBy   string           `bson:"created_by,omitempty"`
	UpdatedBy   string           `bson:"updated_by,omitempty"`
	Source      string           `bson:"source,omitempty"`
	IsDeleted   bool             `bson:"is_deleted,omitempty"`
	DeletedAt   *time.Time       `bson:"deleted_at,omitempty"`
	DeletedBy   string           `bson:"deleted_by,omitempty"`
}

func (r *VenueRepository) Save(ctx context.Context, v *venue.Venue) error {
	doc := toVenueDocument(v)
	doc.ID = bson.NewObjectID()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.CreatedBy = audit.ActorFrom(ctx)
	doc.UpdatedBy = audit.ActorFrom(ctx)
	doc.Source = audit.SourceFrom(ctx)

	id, err := venue.NewVenueID(doc.ID.Hex())
	if err != nil {
		return fmt.Errorf("ID生成エラー: %w", err)
	}
	v.SetID(id)

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("会場の保存エラー: %w", err)
	}
	return nil
}

func (r *VenueRepository) FindByID(ctx context.Context, id venue.VenueID) (*venue.Venue, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc venueDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "is_deleted": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("会場が見つかりません")
		}
		return nil, fmt.Errorf("会場取得エラー: %w", err)
	}
	return fromVenueDocument(&doc)
}

func (r *VenueRepository) Search(ctx context.Context, criteria venue.SearchCriteria) ([]*venue.Venue, error) {
	filter := buildVenueFilter(criteria)

	sortOrder := 1
	if criteria.Order == "desc" {
		sortOrder = -1
	}
	sortField := criteria.Sort
	if sortField == "" {
		sortField = "created_at"
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(int64(criteria.Offset)).
		SetLimit(int64(criteria.Limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("会場検索エラー: %w", err)
	}
	defer cursor.Close(ctx)
	return scanVenueCursor(ctx, cursor)
}

func (r *VenueRepository) Count(ctx context.Context, criteria venue.SearchCriteria) (int64, error) {
	return r.collection.CountDocuments(ctx, buildVenueFilter(criteria))
}

func (r *VenueRepository) Update(ctx context.Context, v *venue.Venue) error {
	objectID, err := bson.ObjectIDFromHex(v.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	doc := toVenueDocument(v)
	doc.UpdatedAt = time.Now()
	doc.UpdatedBy = audit.ActorFrom(ctx)

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, doc)
	if err != nil {
		return fmt.Errorf("会場の更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("会場が見つかりません")
	}
	return nil
}

func (r *VenueRepository) Delete(ctx context.Context, id venue.VenueID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID, "is_deleted": bson.M{"$ne": true}},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": audit.ActorFrom(ctx),
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("会場の削除エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("会場が見つかりません")
	}
	return nil
}

func (r *VenueRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "prefecture", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}
	return nil
}

func toVenueDocument(v *venue.Venue) *venueDocument {
	var objectID bson.ObjectID
	if v.ID().Value() != "" {
		if oid, err := bson.ObjectIDFromHex(v.ID().Value()); err == nil {
			objectID = oid
		}
	}
	return &venueDocument{
		ID:          objectID,
		Name:        v.Name(),
		NameEn:      v.NameEn(),
		Prefecture:  v.Prefecture(),
		City:        v.City(),
		Address:     v.Address(),
		Capacity:    v.Capacity(),
		OfficialURL: v.OfficialURL(),
		Sources:     toSourceDocuments(v.Sources()),
		CreatedAt:   v.CreatedAt(),
		UpdatedAt:   v.UpdatedAt(),
	}
}

func fromVenueDocument(doc *venueDocument) (*venue.Venue, error) {
	id, err := venue.NewVenueID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	return venue.Reconstruct(
		id,
		doc.Name,
		doc.NameEn,
		doc.Prefecture,
		doc.City,
		doc.Address,
		doc.Capacity,
		doc.OfficialURL,
		fromSourceDocuments(doc.Sources),
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

func buildVenueFilter(criteria venue.SearchCriteria) bson.M {
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}
	if criteria.Name != nil {
		filter["name"] = bson.M{"$regex": safePartialMatchRegex(*criteria.Name), "$options": "i"}
	}
	if criteria.Prefecture != nil {
		filter["prefecture"] = *criteria.Prefecture
	}
	return filter
}

func scanVenueCursor(ctx context.Context, cursor *mongo.Cursor) ([]*venue.Venue, error) {
	var result []*venue.Venue
	for cursor.Next(ctx) {
		var doc venueDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("デコードエラー: %w", err)
		}
		v, err := fromVenueDocument(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		result = append(result, v)
	}
	if cursor.Err() != nil {
		return nil, fmt.Errorf("カーソルエラー: %w", cursor.Err())
	}
	if result == nil {
		return []*venue.Venue{}, nil
	}
	return result, nil
}
