package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/submission"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SubmissionRepository はMongoDBを使用した投稿審査リポジトリの実装
type SubmissionRepository struct {
	collection *mongo.Collection
}

// NewSubmissionRepository はMongoDB投稿審査リポジトリを作成する
func NewSubmissionRepository(db *mongo.Database) *SubmissionRepository {
	return &SubmissionRepository{
		collection: db.Collection("submissions"),
	}
}

// submissionDocument はMongoDBに保存するドキュメント構造
type submissionDocument struct {
	ID               bson.ObjectID `bson:"_id,omitempty"`
	TargetType       string        `bson:"target_type"`
	Payload          string        `bson:"payload"`
	SourceURLs       []string      `bson:"source_urls"`
	ContributorEmail string        `bson:"contributor_email"`
	SnsUserID        string        `bson:"sns_user_id,omitempty"`
	Status           string        `bson:"status"`
	RevisionNote     string        `bson:"revision_note,omitempty"`
	ReviewedBy       string        `bson:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time    `bson:"reviewed_at,omitempty"`
	CreatedAt        time.Time     `bson:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at"`
}

// toSubmissionDocument はドメインモデルをMongoDBドキュメントに変換する
func toSubmissionDocument(s *submission.Submission) (*submissionDocument, error) {
	var objectID bson.ObjectID
	if s.ID().Value() != "" {
		var err error
		objectID, err = bson.ObjectIDFromHex(s.ID().Value())
		if err != nil {
			return nil, fmt.Errorf("無効な投稿審査ID %q: %w", s.ID().Value(), err)
		}
	}

	sourceURLs := make([]string, 0, len(s.SourceURLs()))
	for _, u := range s.SourceURLs() {
		sourceURLs = append(sourceURLs, u.Value())
	}

	return &submissionDocument{
		ID:               objectID,
		TargetType:       string(s.TargetType()),
		Payload:          s.Payload(),
		SourceURLs:       sourceURLs,
		ContributorEmail: s.ContributorEmail().Value(),
		SnsUserID:        s.SnsUserID(),
		Status:           string(s.Status()),
		RevisionNote:     s.RevisionNote(),
		ReviewedBy:       s.ReviewedBy(),
		ReviewedAt:       s.ReviewedAt(),
		CreatedAt:        s.CreatedAt(),
		UpdatedAt:        s.UpdatedAt(),
	}, nil
}

// toSubmissionDomain はMongoDBドキュメントをドメインモデルに変換する
func toSubmissionDomain(doc *submissionDocument) (*submission.Submission, error) {
	id, err := submission.NewSubmissionID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}

	targetType, err := submission.NewSubmissionType(doc.TargetType)
	if err != nil {
		return nil, err
	}

	sourceURLs := make([]submission.SourceURL, 0, len(doc.SourceURLs))
	for _, rawURL := range doc.SourceURLs {
		srcURL, err := submission.NewSourceURL(rawURL)
		if err != nil {
			return nil, fmt.Errorf("無効な参照元URL %q: %w", rawURL, err)
		}
		sourceURLs = append(sourceURLs, srcURL)
	}

	contributorEmail, err := submission.NewContributorEmail(doc.ContributorEmail)
	if err != nil {
		return nil, err
	}

	status, err := submission.NewSubmissionStatus(doc.Status)
	if err != nil {
		return nil, err
	}

	return submission.Reconstruct(
		id,
		targetType,
		doc.Payload,
		sourceURLs,
		contributorEmail,
		doc.SnsUserID,
		status,
		doc.RevisionNote,
		doc.ReviewedBy,
		doc.ReviewedAt,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

// Save は新しい投稿審査を保存する
func (r *SubmissionRepository) Save(ctx context.Context, s *submission.Submission) error {
	doc, err := toSubmissionDocument(s)
	if err != nil {
		return fmt.Errorf("ドキュメント変換エラー: %w", err)
	}

	if doc.ID.IsZero() {
		doc.ID = bson.NewObjectID()
		doc.CreatedAt = time.Now()
		doc.UpdatedAt = time.Now()

		newID, err := submission.NewSubmissionID(doc.ID.Hex())
		if err != nil {
			return fmt.Errorf("ID生成エラー: %w", err)
		}
		s.SetID(newID)
	}

	if _, insertErr := r.collection.InsertOne(ctx, doc); insertErr != nil {
		return fmt.Errorf("投稿審査の保存エラー: %w", insertErr)
	}

	return nil
}

// FindByID はIDで投稿審査を取得する
func (r *SubmissionRepository) FindByID(ctx context.Context, id submission.SubmissionID) (*submission.Submission, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}

	var doc submissionDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("投稿審査が見つかりません")
		}
		return nil, fmt.Errorf("投稿審査取得エラー: %w", err)
	}

	return toSubmissionDomain(&doc)
}

// FindAll は全ての投稿審査を取得する（新しい順）
func (r *SubmissionRepository) FindAll(ctx context.Context) ([]*submission.Submission, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("投稿審査一覧取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []submissionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	submissions := make([]*submission.Submission, 0, len(docs))
	for _, doc := range docs {
		s, err := toSubmissionDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		submissions = append(submissions, s)
	}

	return submissions, nil
}

// FindPending は審査待ちの投稿審査を取得する（古い順）
func (r *SubmissionRepository) FindPending(ctx context.Context) ([]*submission.Submission, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"status": "pending"}, opts)
	if err != nil {
		return nil, fmt.Errorf("審査待ち投稿審査取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []submissionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	submissions := make([]*submission.Submission, 0, len(docs))
	for _, doc := range docs {
		s, err := toSubmissionDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		submissions = append(submissions, s)
	}

	return submissions, nil
}

// FindByContributorEmail は投稿者メールアドレスで投稿審査を取得する
func (r *SubmissionRepository) FindByContributorEmail(ctx context.Context, email string) ([]*submission.Submission, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"contributor_email": email}, opts)
	if err != nil {
		return nil, fmt.Errorf("メールアドレスによる投稿審査取得エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []submissionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	submissions := make([]*submission.Submission, 0, len(docs))
	for _, doc := range docs {
		s, err := toSubmissionDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		submissions = append(submissions, s)
	}

	return submissions, nil
}

// Update は投稿審査を更新する
func (r *SubmissionRepository) Update(ctx context.Context, s *submission.Submission) error {
	objectID, err := bson.ObjectIDFromHex(s.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}

	sourceURLs := make([]string, 0, len(s.SourceURLs()))
	for _, u := range s.SourceURLs() {
		sourceURLs = append(sourceURLs, u.Value())
	}

	now := time.Now()
	updateDoc := bson.M{
		"$set": bson.M{
			"payload":       s.Payload(),
			"source_urls":   sourceURLs,
			"status":        string(s.Status()),
			"revision_note": s.RevisionNote(),
			"reviewed_by":   s.ReviewedBy(),
			"reviewed_at":   s.ReviewedAt(),
			"updated_at":    now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		return fmt.Errorf("投稿審査更新エラー: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("投稿審査が見つかりません")
	}

	return nil
}

// EnsureIndexes は submissions コレクションに必要なインデックスを作成する
func (r *SubmissionRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: 1}}},
		{Keys: bson.D{{Key: "contributor_email", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	if err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}
	return nil
}
