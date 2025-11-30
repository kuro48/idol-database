package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/idol"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	ID          bson.ObjectID        `bson:"_id,omitempty"`
	Name        string               `bson:"name"`
	Birthdate   time.Time            `bson:"birthdate"`
	AgencyID    *string              `bson:"agency_id,omitempty"`
	SocialLinks *socialLinksDocument `bson:"social_links,omitempty"`
	CreatedAt   time.Time            `bson:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at"`
}

// socialLinksDocument はSNS/外部リンクのドキュメント構造
type socialLinksDocument struct {
	Twitter   *string `bson:"twitter,omitempty"`
	Instagram *string `bson:"instagram,omitempty"`
	TikTok    *string `bson:"tiktok,omitempty"`
	YouTube   *string `bson:"youtube,omitempty"`
	Facebook  *string `bson:"facebook,omitempty"`
	Official  *string `bson:"official,omitempty"`
	FanClub   *string `bson:"fan_club,omitempty"`
}

// toDocument はドメインモデルをMongoDBドキュメントに変換する
func toIdolDocument(i *idol.Idol) *idolDocument {
	// IDの文字列をObjectIDに変換
	objectID, _ := bson.ObjectIDFromHex(i.ID().Value())

	var socialLinksDoc *socialLinksDocument
	if i.SocialLinks() != nil {
		socialLinksDoc = toSocialLinksDocument(i.SocialLinks())
	}

	return &idolDocument{
		ID:          objectID,
		Name:        i.Name().Value(),
		Birthdate:   i.Birthdate().Value(),
		AgencyID:    i.AgencyID(),
		SocialLinks: socialLinksDoc,
		CreatedAt:   i.CreatedAt(),
		UpdatedAt:   i.UpdatedAt(),
	}
}

// toSocialLinksDocument はSocialLinksをドキュメントに変換する
func toSocialLinksDocument(links *idol.SocialLinks) *socialLinksDocument {
	return &socialLinksDocument{
		Twitter:   links.Twitter(),
		Instagram: links.Instagram(),
		TikTok:    links.TikTok(),
		YouTube:   links.YouTube(),
		Facebook:  links.Facebook(),
		Official:  links.Official(),
		FanClub:   links.FanClub(),
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

	var socialLinks *idol.SocialLinks
	if doc.SocialLinks != nil {
		socialLinks = toSocialLinksDomain(doc.SocialLinks)
	}

	return idol.Reconstruct(id, name, &birthdate, doc.AgencyID, socialLinks, doc.CreatedAt, doc.UpdatedAt), nil
}

// toSocialLinksDomain はドキュメントからSocialLinksドメインモデルを作成する
func toSocialLinksDomain(doc *socialLinksDocument) *idol.SocialLinks {
	links := idol.NewSocialLinks()

	if doc.Twitter != nil {
		_ = links.SetTwitter(*doc.Twitter)
	}
	if doc.Instagram != nil {
		_ = links.SetInstagram(*doc.Instagram)
	}
	if doc.TikTok != nil {
		_ = links.SetTikTok(*doc.TikTok)
	}
	if doc.YouTube != nil {
		_ = links.SetYouTube(*doc.YouTube)
	}
	if doc.Facebook != nil {
		_ = links.SetFacebook(*doc.Facebook)
	}
	if doc.Official != nil {
		_ = links.SetOfficial(*doc.Official)
	}
	if doc.FanClub != nil {
		_ = links.SetFanClub(*doc.FanClub)
	}

	return links
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
			"agency_id":   doc.AgencyID,
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

// FindByAgencyID は事務所IDでアイドルを検索する
func (r *IdolRepository) FindByAgencyID(ctx context.Context, agencyID string) ([]*idol.Idol, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"agency_id": agencyID})
	if err != nil {
		return nil, fmt.Errorf("事務所IDによるアイドル検索エラー: %w", err)
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

func (r *IdolRepository) Search(ctx context.Context, criteria idol.SearchCriteria) ([]*idol.Idol, error) {
    filter := buildMongoFilter(criteria)

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
        return nil, err
    }
    defer cursor.Close(ctx)

    var idols []*idol.Idol
    if err := cursor.All(ctx, &idols); err != nil {
        return nil, err
    }

    return idols, nil
}

func buildMongoFilter(criteria idol.SearchCriteria) bson.M {
    filter := bson.M{}

    // 名前検索（部分一致）
    if criteria.Name != nil {
        filter["name"] = bson.M{"$regex": *criteria.Name, "$options": "i"}
    }

    // 国籍（完全一致）
    if criteria.Nationality != nil {
        filter["nationality"] = *criteria.Nationality
    }

    // グループID
    if criteria.GroupID != nil {
        filter["group_id"] = *criteria.GroupID
    }

    // 事務所ID
    if criteria.AgencyID != nil {
        filter["agency_id"] = *criteria.AgencyID
    }

    // 年齢範囲（生年月日から逆算）
    if criteria.AgeMin != nil || criteria.AgeMax != nil {
        now := time.Now()
        birthdateFilter := bson.M{}

        if criteria.AgeMax != nil {
            // AgeMax歳より若い → 生年月日がこれより後
            minBirthdate := now.AddDate(-*criteria.AgeMax-1, 0, 0)
            birthdateFilter["$gte"] = minBirthdate
        }
        if criteria.AgeMin != nil {
            // AgeMin歳以上 → 生年月日がこれより前
            maxBirthdate := now.AddDate(-*criteria.AgeMin, 0, 0)
            birthdateFilter["$lte"] = maxBirthdate
        }

        if len(birthdateFilter) > 0 {
            filter["birthdate"] = birthdateFilter
        }
    }

    // 生年月日範囲
    if criteria.BirthdateFrom != nil || criteria.BirthdateTo != nil {
        birthdateFilter := bson.M{}
        if criteria.BirthdateFrom != nil {
            birthdateFilter["$gte"] = *criteria.BirthdateFrom
        }
        if criteria.BirthdateTo != nil {
            birthdateFilter["$lte"] = *criteria.BirthdateTo
        }
        filter["birthdate"] = birthdateFilter
    }

    return filter
}

func (r *IdolRepository) Count(ctx context.Context, criteria idol.SearchCriteria) (int64, error) {
    filter := buildMongoFilter(criteria)
    return r.collection.CountDocuments(ctx, filter)
}

// EnsureIndexes は検索パフォーマンス向上のためのインデックスを作成
func (r *IdolRepository) EnsureIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        // 名前インデックス（部分一致検索用）
        {
            Keys: bson.D{
                {Key: "name", Value: 1},
            },
        },
        // 国籍インデックス（フィルタリング用）
        {
            Keys: bson.D{
                {Key: "nationality", Value: 1},
            },
        },
        // グループIDインデックス（フィルタリング用）
        {
            Keys: bson.D{
                {Key: "group_id", Value: 1},
            },
        },
        // 事務所IDインデックス（フィルタリング用）
        {
            Keys: bson.D{
                {Key: "agency_id", Value: 1},
            },
        },
        // 生年月日インデックス（年齢範囲検索・ソート用）
        {
            Keys: bson.D{
                {Key: "birthdate", Value: 1},
            },
        },
        // 作成日時インデックス（デフォルトソート用）
        {
            Keys: bson.D{
                {Key: "created_at", Value: -1},
            },
        },
        // 複合インデックス（国籍＋生年月日での検索最適化）
        {
            Keys: bson.D{
                {Key: "nationality", Value: 1},
                {Key: "birthdate", Value: 1},
            },
        },
    }

    _, err := r.collection.Indexes().CreateMany(ctx, indexes)
    if err != nil {
        return fmt.Errorf("インデックス作成エラー: %w", err)
    }

    return nil
}
