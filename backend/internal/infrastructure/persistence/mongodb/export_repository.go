package mongodb

import (
	"context"
	"time"

	domainExport "github.com/kuro48/idol-api/internal/domain/export"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ExportLogRepository はMongoDBを使用したエクスポートログリポジトリ
type ExportLogRepository struct {
	collection *mongo.Collection
}

// NewExportLogRepository はリポジトリを作成する
func NewExportLogRepository(db *mongo.Database) *ExportLogRepository {
	return &ExportLogRepository{collection: db.Collection("export_logs")}
}

type exportLogDocument struct {
	ID          string    `bson:"_id"`
	Resource    string    `bson:"resource"`
	Format      string    `bson:"format"`
	RecordCount int       `bson:"record_count"`
	ExecutedBy  string    `bson:"executed_by"`
	Status      string    `bson:"status"`
	ErrorMsg    string    `bson:"error_msg,omitempty"`
	ExecutedAt  time.Time `bson:"executed_at"`
}

func (r *ExportLogRepository) Save(ctx context.Context, log *domainExport.ExportLog) error {
	doc := &exportLogDocument{
		ID:          log.ID(),
		Resource:    string(log.Resource()),
		Format:      string(log.Format()),
		RecordCount: log.RecordCount(),
		ExecutedBy:  log.ExecutedBy(),
		Status:      string(log.Status()),
		ErrorMsg:    log.ErrorMsg(),
		ExecutedAt:  log.ExecutedAt(),
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *ExportLogRepository) FindRecent(ctx context.Context, limit int) ([]*domainExport.ExportLog, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "executed_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []exportLogDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	logs := make([]*domainExport.ExportLog, 0, len(docs))
	for _, doc := range docs {
		logs = append(logs, docToExportLog(&doc))
	}
	return logs, nil
}

func (r *ExportLogRepository) FindLastByActor(ctx context.Context, actor string, since time.Time) (*domainExport.ExportLog, error) {
	var doc exportLogDocument
	err := r.collection.FindOne(ctx,
		bson.M{"executed_by": actor, "executed_at": bson.M{"$gte": since}},
		options.FindOne().SetSort(bson.D{{Key: "executed_at", Value: -1}}),
	).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return docToExportLog(&doc), nil
}

func docToExportLog(doc *exportLogDocument) *domainExport.ExportLog {
	log := domainExport.NewExportLog(doc.ID, domainExport.ExportResource(doc.Resource), domainExport.ExportFormat(doc.Format), doc.ExecutedBy)
	log.SetRecordCount(doc.RecordCount)
	if doc.Status == string(domainExport.ExportStatusFailed) {
		log.MarkFailed(doc.ErrorMsg)
	}
	return log
}

// EnsureIndexes は export_logs コレクションに必要なインデックスを作成する
func (r *ExportLogRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "executed_by", Value: 1}, {Key: "executed_at", Value: -1}}},
		{Keys: bson.D{{Key: "executed_at", Value: -1}}},
	})
	return err
}
