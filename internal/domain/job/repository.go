package job

import "context"

// Repository はジョブのリポジトリインターフェース
type Repository interface {
	Save(ctx context.Context, job *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	Update(ctx context.Context, job *Job) error
	FindByStatus(ctx context.Context, status JobStatus, limit int) ([]*Job, error)
}
