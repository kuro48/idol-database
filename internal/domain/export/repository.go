package export

import (
	"context"
	"time"
)

// LogRepository はエクスポートログのリポジトリインターフェース
type LogRepository interface {
	Save(ctx context.Context, log *ExportLog) error
	FindRecent(ctx context.Context, limit int) ([]*ExportLog, error)
	FindLastByActor(ctx context.Context, actor string, since time.Time) (*ExportLog, error)
}
