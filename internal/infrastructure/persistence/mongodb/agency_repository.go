package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/agency"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
}

// Save は事務所を保存する
func (r *AgencyRepository) Save(ctx context.Context, a *agency.Agency) error {
	doc := toAgencyDocument(a)
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("事務所の保存エラー: %w", err)
	}
	return nil
}

// FindByID はIDで事務所を検索する
func (r *AgencyRepository) FindByID(ctx context.Context, id agency.AgencyID) (*agency.Agency, error) {
	var doc agencyDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.Value()}).Decode(&doc)
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
	cursor, err := r.collection.Find(ctx, bson.M{})
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

// Update は事務所を更新する
func (r *AgencyRepository) Update(ctx context.Context, a *agency.Agency) error {
	doc := toAgencyDocument(a)
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": a.ID().Value()}, doc)
	if err != nil {
		return fmt.Errorf("事務所の更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("事務所が見つかりません")
	}
	return nil
}

// Delete は事務所を削除する
func (r *AgencyRepository) Delete(ctx context.Context, id agency.AgencyID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.Value()})
	if err != nil {
		return fmt.Errorf("事務所の削除エラー: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("事務所が見つかりません")
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
