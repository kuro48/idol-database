package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/membership"
	"github.com/kuro48/idol-api/internal/shared/audit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MembershipRepository struct {
	collection *mongo.Collection
}

func NewMembershipRepository(db *mongo.Database) *MembershipRepository {
	return &MembershipRepository{
		collection: db.Collection("memberships"),
	}
}

type membershipDocument struct {
	ID        bson.ObjectID    `bson:"_id,omitempty"`
	IdolID    string           `bson:"idol_id"`
	GroupID   string           `bson:"group_id"`
	Role      string           `bson:"role"`
	JoinedAt  *time.Time       `bson:"joined_at,omitempty"`
	LeftAt    *time.Time       `bson:"left_at,omitempty"`
	Sources   []sourceDocument `bson:"sources,omitempty"`
	Version   int              `bson:"version"`
	CreatedAt time.Time        `bson:"created_at"`
	UpdatedAt time.Time        `bson:"updated_at"`
	CreatedBy string           `bson:"created_by,omitempty"`
	UpdatedBy string           `bson:"updated_by,omitempty"`
	Source    string           `bson:"source,omitempty"`
	IsDeleted bool             `bson:"is_deleted,omitempty"`
	DeletedAt *time.Time       `bson:"deleted_at,omitempty"`
	DeletedBy string           `bson:"deleted_by,omitempty"`
}

func (r *MembershipRepository) Save(ctx context.Context, m *membership.Membership) error {
	doc := toMembershipDocument(m)
	doc.ID = bson.NewObjectID()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.CreatedBy = audit.ActorFrom(ctx)
	doc.UpdatedBy = audit.ActorFrom(ctx)
	doc.Source = audit.SourceFrom(ctx)

	id, err := membership.NewMembershipID(doc.ID.Hex())
	if err != nil {
		return fmt.Errorf("ID生成エラー: %w", err)
	}
	m.SetID(id)

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("メンバーシップの保存エラー: %w", err)
	}
	return nil
}

func (r *MembershipRepository) FindByID(ctx context.Context, id membership.MembershipID) (*membership.Membership, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc membershipDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "is_deleted": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("メンバーシップが見つかりません")
		}
		return nil, fmt.Errorf("メンバーシップ取得エラー: %w", err)
	}
	return fromMembershipDocument(&doc)
}

func (r *MembershipRepository) FindByIdolID(ctx context.Context, idolID string) ([]*membership.Membership, error) {
	cursor, err := r.collection.Find(ctx,
		bson.M{"idol_id": idolID, "is_deleted": bson.M{"$ne": true}},
		options.Find().SetSort(bson.D{{Key: "joined_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("アイドルのメンバーシップ取得エラー: %w", err)
	}
	defer cursor.Close(ctx)
	return scanMembershipCursor(ctx, cursor)
}

func (r *MembershipRepository) FindByGroupID(ctx context.Context, groupID string) ([]*membership.Membership, error) {
	cursor, err := r.collection.Find(ctx,
		bson.M{"group_id": groupID, "is_deleted": bson.M{"$ne": true}},
		options.Find().SetSort(bson.D{{Key: "joined_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("グループのメンバーシップ取得エラー: %w", err)
	}
	defer cursor.Close(ctx)
	return scanMembershipCursor(ctx, cursor)
}

func (r *MembershipRepository) Search(ctx context.Context, criteria membership.SearchCriteria) ([]*membership.Membership, error) {
	filter := buildMembershipFilter(criteria)

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
		return nil, fmt.Errorf("メンバーシップ検索エラー: %w", err)
	}
	defer cursor.Close(ctx)
	return scanMembershipCursor(ctx, cursor)
}

func (r *MembershipRepository) Count(ctx context.Context, criteria membership.SearchCriteria) (int64, error) {
	return r.collection.CountDocuments(ctx, buildMembershipFilter(criteria))
}

func (r *MembershipRepository) Update(ctx context.Context, m *membership.Membership) error {
	objectID, err := bson.ObjectIDFromHex(m.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	doc := toMembershipDocument(m)
	doc.UpdatedAt = time.Now()
	doc.UpdatedBy = audit.ActorFrom(ctx)

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, doc)
	if err != nil {
		return fmt.Errorf("メンバーシップの更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("メンバーシップが見つかりません")
	}
	return nil
}

func (r *MembershipRepository) Delete(ctx context.Context, id membership.MembershipID) error {
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
		return fmt.Errorf("メンバーシップの削除エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("メンバーシップが見つかりません")
	}
	return nil
}

func (r *MembershipRepository) Restore(ctx context.Context, id membership.MembershipID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID, "is_deleted": true},
		bson.M{
			"$set": bson.M{
				"is_deleted": false,
				"updated_at": now,
				"updated_by": audit.ActorFrom(ctx),
			},
			"$unset": bson.M{"deleted_at": "", "deleted_by": ""},
		},
	)
	if err != nil {
		return fmt.Errorf("メンバーシップ復元エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("削除済みメンバーシップが見つかりません")
	}
	return nil
}

func (r *MembershipRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "idol_id", Value: 1}}},
		{Keys: bson.D{{Key: "group_id", Value: 1}}},
		{Keys: bson.D{{Key: "joined_at", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "idol_id", Value: 1}, {Key: "group_id", Value: 1}}},
		{Keys: bson.D{{Key: "idol_id", Value: 1}, {Key: "joined_at", Value: -1}}},
		{Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "joined_at", Value: -1}}},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}
	return nil
}

func toMembershipDocument(m *membership.Membership) *membershipDocument {
	var objectID bson.ObjectID
	if m.ID().Value() != "" {
		if oid, err := bson.ObjectIDFromHex(m.ID().Value()); err == nil {
			objectID = oid
		}
	}
	return &membershipDocument{
		ID:        objectID,
		IdolID:    m.IdolID(),
		GroupID:   m.GroupID(),
		Role:      m.Role().String(),
		JoinedAt:  m.JoinedAt(),
		LeftAt:    m.LeftAt(),
		Sources:   toSourceDocuments(m.Sources()),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}
}

func fromMembershipDocument(doc *membershipDocument) (*membership.Membership, error) {
	id, err := membership.NewMembershipID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	role, err := membership.NewRole(doc.Role)
	if err != nil {
		role = membership.RoleMember
	}

	sources := fromSourceDocuments(doc.Sources)

	return membership.Reconstruct(
		id,
		doc.IdolID,
		doc.GroupID,
		role,
		doc.JoinedAt,
		doc.LeftAt,
		sources,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

func buildMembershipFilter(criteria membership.SearchCriteria) bson.M {
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}

	if criteria.IdolID != nil {
		filter["idol_id"] = *criteria.IdolID
	}
	if criteria.GroupID != nil {
		filter["group_id"] = *criteria.GroupID
	}
	if criteria.IsActive != nil {
		if *criteria.IsActive {
			filter["left_at"] = bson.M{"$exists": false}
		} else {
			filter["left_at"] = bson.M{"$exists": true}
		}
	}
	if criteria.Role != nil {
		filter["role"] = criteria.Role.String()
	}

	return filter
}

func scanMembershipCursor(ctx context.Context, cursor *mongo.Cursor) ([]*membership.Membership, error) {
	var result []*membership.Membership
	for cursor.Next(ctx) {
		var doc membershipDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("デコードエラー: %w", err)
		}
		m, err := fromMembershipDocument(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		result = append(result, m)
	}
	if cursor.Err() != nil {
		return nil, fmt.Errorf("カーソルエラー: %w", cursor.Err())
	}
	if result == nil {
		return []*membership.Membership{}, nil
	}
	return result, nil
}
