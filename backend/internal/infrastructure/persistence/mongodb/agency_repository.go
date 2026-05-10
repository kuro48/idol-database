package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/agency"
	"github.com/kuro48/idol-api/internal/shared/audit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AgencyRepository はMongoDBを使用した事務所リポジトリの実装
type AgencyRepository struct {
	collection *mongo.Collection
}

// NewAgencyRepository はMongoDBAgencyRepositoryを作成する
func NewAgencyRepository(db *mongo.Database) *AgencyRepository {
	return &AgencyRepository{
		collection: db.Collection("agencies"),
	}
}

// agencyDocument はMongoDBに保存する事務所ドキュメント
type agencyDocument struct {
	ID              string     `bson:"_id"`
	Name            string     `bson:"name"`
	NameEn          *string    `bson:"name_en,omitempty"`
	FoundedDate     *time.Time `bson:"founded_date,omitempty"`
	Country         string     `bson:"country"`
	OfficialWebsite *string    `bson:"official_website,omitempty"`
	Description     *string    `bson:"description,omitempty"`
	LogoURL         *string    `bson:"logo_url,omitempty"`
	CreatedAt       time.Time  `bson:"created_at"`
	UpdatedAt       time.Time  `bson:"updated_at"`
	CreatedBy       string     `bson:"created_by,omitempty"`
	UpdatedBy       string     `bson:"updated_by,omitempty"`
	Source          string     `bson:"source,omitempty"`
	IsDeleted       bool       `bson:"is_deleted,omitempty"`
	DeletedAt       *time.Time `bson:"deleted_at,omitempty"`
	DeletedBy       string     `bson:"deleted_by,omitempty"`
}

// Save は事務所を保存する
func (r *AgencyRepository) Save(ctx context.Context, a *agency.Agency) error {
	doc := toAgencyDocument(a)
	doc.CreatedBy = audit.ActorFrom(ctx)
	doc.UpdatedBy = audit.ActorFrom(ctx)
	doc.Source = audit.SourceFrom(ctx)
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("事務所の保存エラー: %w", err)
	}
	return nil
}

// FindByID はIDで事務所を検索する
func (r *AgencyRepository) FindByID(ctx context.Context, id agency.AgencyID) (*agency.Agency, error) {
	var doc agencyDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.Value(), "is_deleted": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("事務所が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("事務所の検索エラー: %w", err)
	}
	return fromAgencyDocument(&doc)
}

// FindAll は全ての事務所を取得する
func (r *AgencyRepository) FindAll(ctx context.Context) ([]*agency.Agency, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_deleted": bson.M{"$ne": true}})
	if err != nil {
		return nil, fmt.Errorf("事務所一覧の取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var agencies []*agency.Agency
	for cursor.Next(ctx) {
		var doc agencyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("ドキュメントのデコードエラー: %w", err)
		}

		a, err := fromAgencyDocument(&doc)
		if err != nil {
			return nil, err
		}
		agencies = append(agencies, a)
	}

	return agencies, nil
}

// FindWithPagination はページネーション付きで事務所を検索する
func (r *AgencyRepository) FindWithPagination(ctx context.Context, opts agency.SearchOptions) (*agency.SearchResult, error) {
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}

	if opts.Name != nil {
		filter["name"] = bson.M{"$regex": *opts.Name, "$options": "i"}
	}
	if opts.Country != nil {
		filter["country"] = *opts.Country
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", err)
	}

	sortOrder := 1
	if opts.Order == "desc" {
		sortOrder = -1
	}
	sortField := opts.Sort
	if sortField == "" {
		sortField = "created_at"
	}

	skip := int64((opts.Page - 1) * opts.Limit)
	limit := int64(opts.Limit)

	findOptions := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("事務所一覧の取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var agencies []*agency.Agency
	for cursor.Next(ctx) {
		var doc agencyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("ドキュメントのデコードエラー: %w", err)
		}
		a, err := fromAgencyDocument(&doc)
		if err != nil {
			return nil, err
		}
		agencies = append(agencies, a)
	}

	if agencies == nil {
		agencies = []*agency.Agency{}
	}

	return &agency.SearchResult{Agencies: agencies, Total: total}, nil
}

// Update は事務所を更新する
func (r *AgencyRepository) Update(ctx context.Context, a *agency.Agency) error {
	doc := toAgencyDocument(a)
	doc.UpdatedBy = audit.ActorFrom(ctx)
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": a.ID().Value()}, doc)
	if err != nil {
		return fmt.Errorf("事務所の更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("事務所が見つかりません")
	}
	return nil
}

// Delete は事務所をソフトデリートする
func (r *AgencyRepository) Delete(ctx context.Context, id agency.AgencyID) error {
	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id.Value(), "is_deleted": bson.M{"$ne": true}},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": audit.ActorFrom(ctx),
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("事務所の削除エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("事務所が見つかりません")
	}
	return nil
}

// Restore はソフトデリートされた事務所を復元する
func (r *AgencyRepository) Restore(ctx context.Context, id agency.AgencyID) error {
	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id.Value(), "is_deleted": true},
		bson.M{
			"$set": bson.M{
				"is_deleted": false,
				"updated_at": now,
				"updated_by": audit.ActorFrom(ctx),
			},
			"$unset": bson.M{
				"deleted_at": "",
				"deleted_by": "",
			},
		},
	)
	if err != nil {
		return fmt.Errorf("事務所復元エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("削除済み事務所が見つかりません")
	}
	return nil
}

// ExistsByID はIDで事務所の存在をチェックする
func (r *AgencyRepository) ExistsByID(ctx context.Context, id agency.AgencyID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.Value()})
	if err != nil {
		return false, fmt.Errorf("事務所の存在チェックエラー: %w", err)
	}
	return count > 0, nil
}

// ExistsByName は名前で事務所の存在をチェックする
func (r *AgencyRepository) ExistsByName(ctx context.Context, name agency.AgencyName) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"name": name.Value()})
	if err != nil {
		return false, fmt.Errorf("名前チェックエラー: %w", err)
	}
	return count > 0, nil
}

// toAgencyDocument はドメインモデルをMongoDBドキュメントに変換する
func toAgencyDocument(a *agency.Agency) *agencyDocument {
	return &agencyDocument{
		ID:              a.ID().Value(),
		Name:            a.Name().Value(),
		NameEn:          a.NameEn(),
		FoundedDate:     a.FoundedDate(),
		Country:         a.Country().Value(),
		OfficialWebsite: a.OfficialWebsite(),
		Description:     a.Description(),
		LogoURL:         a.LogoURL(),
		CreatedAt:       a.CreatedAt(),
		UpdatedAt:       a.UpdatedAt(),
	}
}

// fromAgencyDocument はMongoDBドキュメントをドメインモデルに変換する
func fromAgencyDocument(doc *agencyDocument) (*agency.Agency, error) {
	id, err := agency.NewAgencyID(doc.ID)
	if err != nil {
		return nil, err
	}

	name, err := agency.NewAgencyName(doc.Name)
	if err != nil {
		return nil, err
	}

	country, err := agency.NewCountry(doc.Country)
	if err != nil {
		return nil, err
	}

	a := agency.NewAgency(id, name, country)
	a.UpdateDetails(nil, doc.NameEn, doc.FoundedDate, doc.OfficialWebsite, doc.Description, doc.LogoURL)

	return a, nil
}

// EnsureIndexes はMongoDBのインデックスを作成する
func (r *AgencyRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		// 事務所名インデックス（検索用）
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
		},
		// 国インデックス（フィルタリング用）
		{
			Keys: bson.D{
				{Key: "country", Value: 1},
			},
		},
		// 設立日インデックス（時系列検索用）
		{
			Keys: bson.D{
				{Key: "founded_date", Value: 1},
			},
		},
		// 作成日時インデックス（デフォルトソート用）
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
			},
		},
		// 複合インデックス: 国 + 作成日時（国別一覧の最適化）
		{
			Keys: bson.D{
				{Key: "country", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		// 複合インデックス: 設立日 + 作成日時（時系列検索 + ソート最適化）
		{
			Keys: bson.D{
				{Key: "founded_date", Value: 1},
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
