package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/song"
	"github.com/kuro48/idol-api/internal/shared/audit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SongRepository struct {
	collection *mongo.Collection
}

func NewSongRepository(db *mongo.Database) *SongRepository {
	return &SongRepository{collection: db.Collection("songs")}
}

type songDocument struct {
	ID            bson.ObjectID    `bson:"_id,omitempty"`
	Title         string           `bson:"title"`
	TitleKana     *string          `bson:"title_kana,omitempty"`
	DurationSec   *int             `bson:"duration_sec,omitempty"`
	ISRC          *string          `bson:"isrc,omitempty"`
	CoverImageURL *string          `bson:"cover_image_url,omitempty"`
	Composers     []string         `bson:"composers,omitempty"`
	Lyricists     []string         `bson:"lyricists,omitempty"`
	Arrangers     []string         `bson:"arrangers,omitempty"`
	Sources       []sourceDocument `bson:"sources,omitempty"`
	Version       int              `bson:"version"`
	CreatedAt     time.Time        `bson:"created_at"`
	UpdatedAt     time.Time        `bson:"updated_at"`
	CreatedBy     string           `bson:"created_by,omitempty"`
	UpdatedBy     string           `bson:"updated_by,omitempty"`
	Source        string           `bson:"source,omitempty"`
	IsDeleted     bool             `bson:"is_deleted,omitempty"`
	DeletedAt     *time.Time       `bson:"deleted_at,omitempty"`
	DeletedBy     string           `bson:"deleted_by,omitempty"`
}

func (r *SongRepository) Save(ctx context.Context, s *song.Song) error {
	doc := toSongDocument(s)
	doc.ID = bson.NewObjectID()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.CreatedBy = audit.ActorFrom(ctx)
	doc.UpdatedBy = audit.ActorFrom(ctx)
	doc.Source = audit.SourceFrom(ctx)

	id, err := song.NewSongID(doc.ID.Hex())
	if err != nil {
		return fmt.Errorf("ID生成エラー: %w", err)
	}
	s.SetID(id)

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("楽曲の保存エラー: %w", err)
	}
	return nil
}

func (r *SongRepository) FindByID(ctx context.Context, id song.SongID) (*song.Song, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc songDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "is_deleted": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("楽曲が見つかりません")
		}
		return nil, fmt.Errorf("楽曲取得エラー: %w", err)
	}
	return fromSongDocument(&doc)
}

func (r *SongRepository) Search(ctx context.Context, criteria song.SearchCriteria) ([]*song.Song, error) {
	filter := buildSongFilter(criteria)

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
		return nil, fmt.Errorf("楽曲検索エラー: %w", err)
	}
	defer cursor.Close(ctx)
	return scanSongCursor(ctx, cursor)
}

func (r *SongRepository) Count(ctx context.Context, criteria song.SearchCriteria) (int64, error) {
	return r.collection.CountDocuments(ctx, buildSongFilter(criteria))
}

func (r *SongRepository) Update(ctx context.Context, s *song.Song) error {
	objectID, err := bson.ObjectIDFromHex(s.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	doc := toSongDocument(s)
	doc.UpdatedAt = time.Now()
	doc.UpdatedBy = audit.ActorFrom(ctx)

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, doc)
	if err != nil {
		return fmt.Errorf("楽曲の更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("楽曲が見つかりません")
	}
	return nil
}

func (r *SongRepository) Delete(ctx context.Context, id song.SongID) error {
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
		return fmt.Errorf("楽曲の削除エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("楽曲が見つかりません")
	}
	return nil
}

func (r *SongRepository) Restore(ctx context.Context, id song.SongID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID, "is_deleted": true},
		bson.M{
			"$set":   bson.M{"is_deleted": false, "updated_at": now, "updated_by": audit.ActorFrom(ctx)},
			"$unset": bson.M{"deleted_at": "", "deleted_by": ""},
		},
	)
	if err != nil {
		return fmt.Errorf("楽曲復元エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("削除済み楽曲が見つかりません")
	}
	return nil
}

func (r *SongRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "title", Value: "text"}}},
		{Keys: bson.D{{Key: "isrc", Value: 1}, {Key: "is_deleted", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}
	return nil
}

func toSongDocument(s *song.Song) *songDocument {
	var objectID bson.ObjectID
	if s.ID().Value() != "" {
		if oid, err := bson.ObjectIDFromHex(s.ID().Value()); err == nil {
			objectID = oid
		}
	}
	return &songDocument{
		ID:            objectID,
		Title:         s.Title(),
		TitleKana:     s.TitleKana(),
		DurationSec:   s.DurationSec(),
		ISRC:          s.ISRC(),
		CoverImageURL: s.CoverImageURL(),
		Composers:     s.Composers(),
		Lyricists:     s.Lyricists(),
		Arrangers:     s.Arrangers(),
		Sources:       toSourceDocuments(s.Sources()),
		CreatedAt:     s.CreatedAt(),
		UpdatedAt:     s.UpdatedAt(),
	}
}

func fromSongDocument(doc *songDocument) (*song.Song, error) {
	id, err := song.NewSongID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	composers := doc.Composers
	if composers == nil {
		composers = []string{}
	}
	lyricists := doc.Lyricists
	if lyricists == nil {
		lyricists = []string{}
	}
	arrangers := doc.Arrangers
	if arrangers == nil {
		arrangers = []string{}
	}

	return song.Reconstruct(
		id,
		doc.Title,
		doc.TitleKana,
		doc.DurationSec,
		doc.ISRC,
		doc.CoverImageURL,
		composers,
		lyricists,
		arrangers,
		fromSourceDocuments(doc.Sources),
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

func buildSongFilter(criteria song.SearchCriteria) bson.M {
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}
	if criteria.Title != nil {
		filter["title"] = bson.M{"$regex": safePartialMatchRegex(*criteria.Title), "$options": "i"}
	}
	if criteria.ISRC != nil {
		filter["isrc"] = *criteria.ISRC
	}
	return filter
}

func scanSongCursor(ctx context.Context, cursor *mongo.Cursor) ([]*song.Song, error) {
	var result []*song.Song
	for cursor.Next(ctx) {
		var doc songDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("デコードエラー: %w", err)
		}
		s, err := fromSongDocument(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		result = append(result, s)
	}
	if cursor.Err() != nil {
		return nil, fmt.Errorf("カーソルエラー: %w", cursor.Err())
	}
	if result == nil {
		return []*song.Song{}, nil
	}
	return result, nil
}
