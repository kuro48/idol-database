package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// JobRepository はMongoDBを使用した非同期ジョブリポジトリの実装
type JobRepository struct {
	collection *mongo.Collection
}

// NewJobRepository はJobRepositoryを作成する
func NewJobRepository(db *mongo.Database) *JobRepository {
	return &JobRepository{
		collection: db.Collection("async_jobs"),
	}
}

// jobDocument はMongoDBに保存するドキュメント構造
type jobDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	JobType     string        `bson:"job_type"`
	Status      string        `bson:"status"`
	Payload     []byte        `bson:"payload,omitempty"`
	Result      []byte        `bson:"result,omitempty"`
	ErrorMsg    string        `bson:"error_msg,omitempty"`
	CreatedBy   string        `bson:"created_by,omitempty"`
	CreatedAt   time.Time     `bson:"created_at"`
	StartedAt   *time.Time    `bson:"started_at,omitempty"`
	CompletedAt *time.Time    `bson:"completed_at,omitempty"`
}

// Save は新しいジョブを保存する
func (r *JobRepository) Save(ctx context.Context, job *domainJob.Job) error {
	doc := jobDocument{
		ID:          bson.NewObjectID(),
		JobType:     string(job.JobType()),
		Status:      string(job.Status()),
		Payload:     job.Payload(),
		CreatedBy:   job.CreatedBy(),
		CreatedAt:   job.CreatedAt(),
	}

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("ジョブの保存エラー: %w", err)
	}

	job.SetID(doc.ID.Hex())
	return nil
}

// FindByID はIDでジョブを検索する
func (r *JobRepository) FindByID(ctx context.Context, id string) (*domainJob.Job, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("無効なジョブID形式: %w", err)
	}

	var doc jobDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("ジョブが見つかりません")
		}
		return nil, fmt.Errorf("ジョブ取得エラー: %w", err)
	}

	return toJobDomain(&doc), nil
}

// Update はジョブを更新する
func (r *JobRepository) Update(ctx context.Context, job *domainJob.Job) error {
	objectID, err := bson.ObjectIDFromHex(job.ID())
	if err != nil {
		return fmt.Errorf("無効なジョブID形式: %w", err)
	}

	setFields := bson.M{
		"status":       string(job.Status()),
		"result":       job.Result(),
		"error_msg":    job.ErrorMsg(),
		"started_at":   job.StartedAt(),
		"completed_at": job.CompletedAt(),
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": setFields})
	if err != nil {
		return fmt.Errorf("ジョブ更新エラー: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("ジョブが見つかりません")
	}

	return nil
}

// FindByStatus はステータスでジョブを検索する
func (r *JobRepository) FindByStatus(ctx context.Context, status domainJob.JobStatus, limit int) ([]*domainJob.Job, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.M{"created_at": 1})

	cursor, err := r.collection.Find(ctx, bson.M{"status": string(status)}, opts)
	if err != nil {
		return nil, fmt.Errorf("ジョブ一覧取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []jobDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	jobs := make([]*domainJob.Job, 0, len(docs))
	for _, doc := range docs {
		jobs = append(jobs, toJobDomain(&doc))
	}

	return jobs, nil
}

// toJobDomain はDocumentをドメインモデルに変換する
func toJobDomain(doc *jobDocument) *domainJob.Job {
	return domainJob.ReconstructJob(
		doc.ID.Hex(),
		domainJob.JobType(doc.JobType),
		domainJob.JobStatus(doc.Status),
		doc.Payload,
		doc.Result,
		doc.ErrorMsg,
		doc.CreatedBy,
		doc.CreatedAt,
		doc.StartedAt,
		doc.CompletedAt,
	)
}

// EnsureIndexes はインデックスを作成する
func (r *JobRepository) EnsureIndexes(ctx context.Context) error {
	// 完了/失敗後7日でTTL削除
	sevenDaysSeconds := int32(7 * 24 * 60 * 60)

	indexes := []mongo.IndexModel{
		{
			// ステータスインデックス（ポーリング用）
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			// 作成日時インデックス（FIFO処理用）
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
		{
			// TTLインデックス: completed_at が設定された後7日で削除
			// completed_at が null のドキュメント（pending/running）は削除されない
			Keys: bson.D{{Key: "completed_at", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(sevenDaysSeconds).
				SetSparse(true).
				SetName("ttl_completed_at"),
		},
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("jobインデックス作成エラー: %w", err)
	}

	return nil
}
