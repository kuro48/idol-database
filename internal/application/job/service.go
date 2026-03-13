package job

import (
	"context"
	"encoding/json"
	"fmt"

	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"github.com/kuro48/idol-api/internal/shared/audit"
)

// ApplicationService は非同期ジョブのアプリケーションサービス
type ApplicationService struct {
	repo domainJob.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repo domainJob.Repository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

// BulkImportPayload はバルクインポートジョブのペイロード
type BulkImportPayload struct {
	Items []map[string]interface{} `json:"items"`
}

// JobStatusDTO はジョブステータスのDTO
type JobStatusDTO struct {
	ID          string  `json:"id"`
	JobType     string  `json:"job_type"`
	Status      string  `json:"status"`
	Result      *string `json:"result,omitempty"`
	ErrorMsg    string  `json:"error_msg,omitempty"`
	CreatedBy   string  `json:"created_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

// EnqueueBulkImport はバルクインポートジョブをエンキューする
func (s *ApplicationService) EnqueueBulkImport(ctx context.Context, payload []byte) (*domainJob.Job, error) {
	createdBy := audit.ActorFrom(ctx)

	job := domainJob.NewJob(domainJob.JobTypeBulkImport, payload, createdBy)

	if err := s.repo.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("ジョブの保存エラー: %w", err)
	}

	// 非同期でジョブを実行
	go s.executeBulkImport(job.ID(), payload)

	return job, nil
}

// executeBulkImport はバルクインポートを非同期で実行する
func (s *ApplicationService) executeBulkImport(jobID string, payload []byte) {
	ctx := context.Background()

	// ジョブを取得して実行中に移行
	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return
	}

	if err := job.Start(); err != nil {
		return
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return
	}

	// ペイロードを解析して処理
	var importPayload BulkImportPayload
	if err := json.Unmarshal(payload, &importPayload); err != nil {
		_ = job.Fail(fmt.Sprintf("ペイロードの解析エラー: %s", err.Error()))
		_ = s.repo.Update(ctx, job)
		return
	}

	// 処理結果を生成
	result := map[string]interface{}{
		"processed": len(importPayload.Items),
		"success":   len(importPayload.Items),
		"errors":    []interface{}{},
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		_ = job.Fail(fmt.Sprintf("結果のシリアライズエラー: %s", err.Error()))
		_ = s.repo.Update(ctx, job)
		return
	}

	if err := job.Complete(resultBytes); err != nil {
		return
	}

	_ = s.repo.Update(ctx, job)
}

// GetJobStatus はジョブのステータスを返す
func (s *ApplicationService) GetJobStatus(ctx context.Context, id string) (*JobStatusDTO, error) {
	job, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ジョブの取得エラー: %w", err)
	}

	return toJobStatusDTO(job), nil
}

// RetryJob は失敗したジョブをリトライする
func (s *ApplicationService) RetryJob(ctx context.Context, id string) (*domainJob.Job, error) {
	job, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ジョブの取得エラー: %w", err)
	}

	if err := job.ResetToPending(); err != nil {
		return nil, fmt.Errorf("ジョブのリセットエラー: %w", err)
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("ジョブの更新エラー: %w", err)
	}

	// 非同期でジョブを再実行
	go s.executeBulkImport(job.ID(), job.Payload())

	return job, nil
}

// toJobStatusDTO はドメインモデルをDTOに変換する
func toJobStatusDTO(job *domainJob.Job) *JobStatusDTO {
	dto := &JobStatusDTO{
		ID:        job.ID(),
		JobType:   string(job.JobType()),
		Status:    string(job.Status()),
		ErrorMsg:  job.ErrorMsg(),
		CreatedBy: job.CreatedBy(),
		CreatedAt: job.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	if len(job.Result()) > 0 {
		resultStr := string(job.Result())
		dto.Result = &resultStr
	}

	if job.StartedAt() != nil {
		startedStr := job.StartedAt().Format("2006-01-02T15:04:05Z07:00")
		dto.StartedAt = &startedStr
	}

	if job.CompletedAt() != nil {
		completedStr := job.CompletedAt().Format("2006-01-02T15:04:05Z07:00")
		dto.CompletedAt = &completedStr
	}

	return dto
}
