package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type IdolRepository struct {
	collection *mongo.Collection
}

func NewIdolRepository(db *mongo.Database) *IdolRepository {
	return &IdolRepository{
		collection: db.Collection("idols"),
	}
}

func (r *IdolRepository) Create(idol *models.Idol) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx,idol)
	if err != nil {
		return fmt.Errorf("アイドル作成エラー: %w", err)
	}

	idol.ID = result.InsertedID.(bson.ObjectID)

	return nil
}

func (r *IdolRepository) FindAll() ([]models.Idol, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("アイドル一覧取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var idols []models.Idol
	if err := cursor.All(ctx, &idols); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	return idols, nil
}

func (r *IdolRepository) FindByID(id string) (*models.Idol, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectID, err := bson.ObjectIDFromHex(id)
    if err != nil {
        return nil, fmt.Errorf("無効なID形式: %w", err)
    }

    var idol models.Idol
    err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&idol)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("アイドルが見つかりません")
        }
        return nil, fmt.Errorf("アイドル取得エラー: %w", err)
    }

    return &idol, nil
}

func (r *IdolRepository) Update(id string, update *models.UpdateIdolRequest) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectID, err := bson.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("無効なID形式: %w", err)
    }

    // 空でないフィールドのみ更新
    updateDoc := bson.M{}
    if update.Name != "" {
        updateDoc["name"] = update.Name
    }
    if update.Group != "" {
        updateDoc["group"] = update.Group
    }
    if update.Birthdate != "" {
        updateDoc["birthdate"] = update.Birthdate
    }
    if update.Nationality != "" {
        updateDoc["nationality"] = update.Nationality
    }
    if update.ImageURL != "" {
        updateDoc["image_url"] = update.ImageURL
    }

    result, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": objectID},
        bson.M{"$set": updateDoc},
    )
    if err != nil {
        return fmt.Errorf("アイドル更新エラー: %w", err)
    }

    if result.MatchedCount == 0 {
        return fmt.Errorf("アイドルが見つかりません")
    }

    return nil
}

func (r *IdolRepository) Delete(id string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectID, err := bson.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("無効なID形式: %w", err)
    }

    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return fmt.Errorf("アイドル削除エラー: %w", err)
    }

    if result.DeletedCount == 0 {
        return fmt.Errorf("アイドルが見つかりません")
    }

    return nil
}