package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/group"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type GroupRepository struct {
	collection *mongo.Collection
}

// FindAll implements group.Repository.
func (r *GroupRepository) FindAll(ctx context.Context) ([]*group.Group, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("グループ一覧取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []*group.Group
	for cursor.Next(ctx) {
		var doc groupDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("グループデコードエラー: %w", err)
		}

		g, err := toGroupDomain(&doc)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("カーソルエラー: %w", err)
	}

	return groups, nil
}

// Update implements group.Repository.
func (r *GroupRepository) Update(ctx context.Context, g *group.Group) error {
	doc := toGroupDocument(g)
	doc.UpdatedAt = time.Now()

	objectID, err := bson.ObjectIDFromHex(g.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": objectID},
		doc,
	)
	if err != nil {
		return fmt.Errorf("グループ更新エラー: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("グループが見つかりません")
	}

	return nil
}

func NewGroupRepository(db *mongo.Database) *GroupRepository {
	return &GroupRepository{
		collection: db.Collection("groups"),
	}
}

type groupDocument struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	Name          string        `bson:"name"`
	FormationDate *time.Time    `bson:"formation_date,omitempty"`
	DisbandDate   *time.Time    `bson:"disband_date,omitempty"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
}

func toGroupDocument(g *group.Group) *groupDocument {
	objectID, _ := bson.ObjectIDFromHex(g.ID().Value())

	var formationDate *time.Time
	if g.FormationDate() != nil {
		t := g.FormationDate().Value()
		formationDate = &t
	}

	var disbandDate *time.Time
	if g.DisbandDate() != nil {
		t := g.DisbandDate().Value()
		disbandDate = &t
	}

	return &groupDocument{
		ID:            objectID,
		Name:          g.Name().Value(),
		FormationDate: formationDate,
		DisbandDate:   disbandDate,
		CreatedAt:     g.CreatedAt(),
		UpdatedAt:     g.UpdatedAt(),
	}
}

func toGroupDomain(doc *groupDocument) (*group.Group, error) {
	id, err := group.NewGroupID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	name, err := group.NewGroupName(doc.Name)
	if err != nil {
		return nil, err
	}

	// FormationDateの変換（nilの場合はnilのまま）
	var formationDate *group.FormationDate
	if doc.FormationDate != nil {
		fdYear, fdMonth, fdDay := doc.FormationDate.Date()
		fd, err := group.NewFormationDate(fdYear, int(fdMonth), fdDay)
		if err != nil {
			return nil, err
		}
		formationDate = &fd
	}

	// DisbandDateの変換（nilの場合はnilのまま）
	var disbandDate *group.DisbandDate
	if doc.DisbandDate != nil {
		ddYear, ddMonth, ddDay := doc.DisbandDate.Date()
		dd, err := group.NewDisbandDate(ddYear, int(ddMonth), ddDay)
		if err != nil {
			return nil, err
		}
		disbandDate = &dd
	}

	return group.Reconstruct(id, name, formationDate, disbandDate, doc.CreatedAt, doc.UpdatedAt), nil
}

func (r *GroupRepository) Save(ctx context.Context, g *group.Group) error {
	doc := toGroupDocument(g)

	// 新規作成の場合はIDを生成
	if doc.ID.IsZero() {
		doc.ID = bson.NewObjectID()
		doc.CreatedAt = time.Now()
		doc.UpdatedAt = time.Now()

		// エンティティにIDを設定
		id, err := group.NewGroupID(doc.ID.Hex())
		if err != nil {
			return fmt.Errorf("ID生成エラー: %w", err)
		}
		g.SetID(id)
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("グループの保存エラー: %w", err)
	}

	return nil
}

// FindByID はIDでグループを検索する
func (r *GroupRepository) FindByID(ctx context.Context, id group.GroupID) (*group.Group, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc groupDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("グループが見つかりません")
		}
		return nil, fmt.Errorf("グループ取得エラー: %w", err)
	}

	return toGroupDomain(&doc)
}

// Delete はグループを削除する
func (r *GroupRepository) Delete(ctx context.Context, id group.GroupID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("グループ削除エラー: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("グループが見つかりません")
	}

	return nil
}

// ExistsByName は同じ名前のグループが存在するかチェック
func (r *GroupRepository) ExistsByName(ctx context.Context, name group.GroupName) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"name": name.Value()})
	if err != nil {
		return false, fmt.Errorf("名前チェックエラー: %w", err)
	}

	return count > 0, nil
}
