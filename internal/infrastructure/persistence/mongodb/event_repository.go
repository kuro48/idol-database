package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/event"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EventRepository はMongoDBを使用したイベントリポジトリの実装
type EventRepository struct {
	collection *mongo.Collection
}

// NewEventRepository はMongoDBEventRepositoryを作成する
func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		collection: db.Collection("events"),
	}
}

// eventDocument はMongoDBに保存するイベントドキュメント
type eventDocument struct {
	ID            string    `bson:"_id,omitempty"`
	Title         string    `bson:"title"`
	EventType     string    `bson:"event_type"`
	StartDateTime time.Time `bson:"start_date_time"`
	EndDateTime   *time.Time `bson:"end_date_time,omitempty"`
	VenueID       *string   `bson:"venue_id,omitempty"`
	PerformerIDs  []string  `bson:"performer_ids"`
	TicketURL     *string   `bson:"ticket_url,omitempty"`
	OfficialURL   *string   `bson:"official_url,omitempty"`
	Description   *string   `bson:"description,omitempty"`
	Tags          []string  `bson:"tags"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
}

// Save は新しいイベントを保存する
func (r *EventRepository) Save(ctx context.Context, e *event.Event) error {
	doc := toEventDocument(e)
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("イベントの保存エラー: %w", err)
	}
	return nil
}

// FindByID はIDでイベントを検索する
func (r *EventRepository) FindByID(ctx context.Context, id event.EventID) (*event.Event, error) {
	var doc eventDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.Value()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("イベントが見つかりません: %w", err)
		}
		return nil, fmt.Errorf("イベントの検索エラー: %w", err)
	}
	return fromEventDocument(&doc)
}

// Search は条件を指定してイベントを検索する
func (r *EventRepository) Search(ctx context.Context, criteria event.SearchCriteria) ([]*event.Event, error) {
	filter := buildEventFilter(criteria)

	opts := options.Find()

	// ソート設定
	sortOrder := 1
	if criteria.Order == "desc" {
		sortOrder = -1
	}
	opts.SetSort(bson.D{{Key: criteria.Sort, Value: sortOrder}})

	// ページネーション
	opts.SetSkip(int64(criteria.Offset))
	opts.SetLimit(int64(criteria.Limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("イベント検索エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []eventDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	events := make([]*event.Event, 0, len(docs))
	for _, doc := range docs {
		e, err := fromEventDocument(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

// Count は検索条件に一致するイベント数を返す
func (r *EventRepository) Count(ctx context.Context, criteria event.SearchCriteria) (int64, error) {
	filter := buildEventFilter(criteria)
	return r.collection.CountDocuments(ctx, filter)
}

// Update は既存のイベントを更新する
func (r *EventRepository) Update(ctx context.Context, e *event.Event) error {
	doc := toEventDocument(e)
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": e.ID().Value()}, doc)
	if err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("イベントが見つかりません")
	}
	return nil
}

// Delete はイベントを削除する
func (r *EventRepository) Delete(ctx context.Context, id event.EventID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.Value()})
	if err != nil {
		return fmt.Errorf("イベントの削除エラー: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("イベントが見つかりません")
	}
	return nil
}

// FindUpcoming は今後開催されるイベントを取得する
func (r *EventRepository) FindUpcoming(ctx context.Context, limit int) ([]*event.Event, error) {
	filter := bson.M{
		"start_date_time": bson.M{"$gte": time.Now()},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "start_date_time", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("今後のイベント取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []eventDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	events := make([]*event.Event, 0, len(docs))
	for _, doc := range docs {
		e, err := fromEventDocument(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

// FindByPerformer はパフォーマーIDでイベントを検索する
func (r *EventRepository) FindByPerformer(ctx context.Context, performerID string, limit int) ([]*event.Event, error) {
	filter := bson.M{
		"performer_ids": bson.M{"$in": []string{performerID}},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "start_date_time", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("パフォーマーのイベント取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []eventDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	events := make([]*event.Event, 0, len(docs))
	for _, doc := range docs {
		e, err := fromEventDocument(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

// toEventDocument はドメインモデルをMongoDBドキュメントに変換する
func toEventDocument(e *event.Event) *eventDocument {
	return &eventDocument{
		ID:            e.ID().Value(),
		Title:         e.Title().Value(),
		EventType:     e.EventType().Value(),
		StartDateTime: e.StartDateTime(),
		EndDateTime:   e.EndDateTime(),
		VenueID:       e.VenueID(),
		PerformerIDs:  e.PerformerIDs(),
		TicketURL:     e.TicketURL(),
		OfficialURL:   e.OfficialURL(),
		Description:   e.Description(),
		Tags:          e.Tags(),
		CreatedAt:     e.CreatedAt(),
		UpdatedAt:     e.UpdatedAt(),
	}
}

// fromEventDocument はMongoDBドキュメントをドメインモデルに変換する
func fromEventDocument(doc *eventDocument) (*event.Event, error) {
	id, err := event.NewEventID(doc.ID)
	if err != nil {
		return nil, err
	}

	title, err := event.NewEventTitle(doc.Title)
	if err != nil {
		return nil, err
	}

	eventType, err := event.NewEventType(doc.EventType)
	if err != nil {
		return nil, err
	}

	return event.Reconstruct(
		id,
		title,
		eventType,
		doc.StartDateTime,
		doc.EndDateTime,
		doc.VenueID,
		doc.PerformerIDs,
		doc.TicketURL,
		doc.OfficialURL,
		doc.Description,
		doc.Tags,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

// buildEventFilter は検索条件からMongoDBフィルタを構築する
func buildEventFilter(criteria event.SearchCriteria) bson.M {
	filter := bson.M{}

	// イベントタイプ
	if criteria.EventType != nil {
		filter["event_type"] = criteria.EventType.Value()
	}

	// 開始日時範囲
	if criteria.StartDateFrom != nil || criteria.StartDateTo != nil {
		dateFilter := bson.M{}
		if criteria.StartDateFrom != nil {
			dateFilter["$gte"] = *criteria.StartDateFrom
		}
		if criteria.StartDateTo != nil {
			dateFilter["$lte"] = *criteria.StartDateTo
		}
		if len(dateFilter) > 0 {
			filter["start_date_time"] = dateFilter
		}
	}

	// 会場ID
	if criteria.VenueID != nil {
		filter["venue_id"] = *criteria.VenueID
	}

	// パフォーマーID
	if criteria.PerformerID != nil {
		filter["performer_ids"] = bson.M{"$in": []string{*criteria.PerformerID}}
	}

	// タグ
	if len(criteria.Tags) > 0 {
		filter["tags"] = bson.M{"$all": criteria.Tags}
	}

	return filter
}

// EnsureIndexes は検索パフォーマンス向上のためのインデックスを作成
func (r *EventRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		// イベントタイプインデックス
		{
			Keys: bson.D{
				{Key: "event_type", Value: 1},
			},
		},
		// 開始日時インデックス（ソートとフィルタ用）
		{
			Keys: bson.D{
				{Key: "start_date_time", Value: 1},
			},
		},
		// 会場IDインデックス
		{
			Keys: bson.D{
				{Key: "venue_id", Value: 1},
			},
		},
		// パフォーマーIDインデックス
		{
			Keys: bson.D{
				{Key: "performer_ids", Value: 1},
			},
		},
		// タグインデックス
		{
			Keys: bson.D{
				{Key: "tags", Value: 1},
			},
		},
		// 作成日時インデックス（デフォルトソート用）
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
			},
		},
		// 複合インデックス1: イベントタイプ + 開始日時（タイプ別時系列検索の最適化）
		{
			Keys: bson.D{
				{Key: "event_type", Value: 1},
				{Key: "start_date_time", Value: 1},
			},
		},
		// 複合インデックス2: 会場ID + 開始日時（会場別イベント一覧の最適化）
		{
			Keys: bson.D{
				{Key: "venue_id", Value: 1},
				{Key: "start_date_time", Value: 1},
			},
		},
		// 複合インデックス3: パフォーマーID + 開始日時（アイドル別イベント一覧の最適化）
		{
			Keys: bson.D{
				{Key: "performer_ids", Value: 1},
				{Key: "start_date_time", Value: 1},
			},
		},
		// 複合インデックス4: 開始日時 + 作成日時（時系列検索 + ソート最適化）
		{
			Keys: bson.D{
				{Key: "start_date_time", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		// 複合インデックス5: イベントタイプ + 開始日時 + 作成日時（複雑な検索の最適化）
		{
			Keys: bson.D{
				{Key: "event_type", Value: 1},
				{Key: "start_date_time", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}

	return nil
}
