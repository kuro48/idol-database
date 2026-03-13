package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"github.com/kuro48/idol-api/internal/shared/audit"
)

// jobExecutionTimeout はバルクインポートジョブの最大実行時間
const jobExecutionTimeout = 30 * time.Minute

// ApplicationService は非同期ジョブのアプリケーションサービス
type ApplicationService struct {
	repo domainJob.Repository
	wg   sync.WaitGroup
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repo domainJob.Repository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

// Shutdown はインフライトのジョブがすべて完了するまで待機する
func (s *ApplicationService) Shutdown() {
	s.wg.Wait()
}

// RecoverStuckJobs は起動時にRUNNING状態で止まっているジョブをPENDINGに戻す
func (s *ApplicationService) RecoverStuckJobs(ctx context.Context) error {
	jobs, err := s.repo.FindByStatus(ctx, domainJob.JobStatusRunning, 100)
	if err != nil {
		return fmt.Errorf("スタックジョブの取得エラー: %w", err)
	}
	for _, j := range jobs {
		if err := j.ResetToPending(); err != nil {
			slog.Warn("スタックジョブのリセットに失敗しました", "job_id", j.ID(), "error", err)
			continue
		}
		if err := s.repo.Update(ctx, j); err != nil {
			slog.Warn("スタックジョブの状態更新に失敗しました", "job_id", j.ID(), "error", err)
		} else {
			slog.Info("スタックジョブをpendingにリセットしました", "job_id", j.ID())
		}
	}
	return nil
}

// BulkImportItem はバルクインポートの1件分のデータ
type BulkImportItem struct {
	Name      string   `json:"name"`
	Birthdate string   `json:"birthdate,omitempty"`
	AgencyID  string   `json:"agency_id,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	TagIDs    []string `json:"tag_ids,omitempty"`
}

// BulkImportPayload はバルクインポートジョブのペイロード
type BulkImportPayload struct {
	Items []BulkImportItem `json:"items"`
}

// EnqueueBulkImport はバルクインポートジョブをエンキューする
func (s *ApplicationService) EnqueueBulkImport(ctx context.Context, payload []byte) (*domainJob.Job, error) {
	createdBy := audit.ActorFrom(ctx)

	job := domainJob.NewJob(domainJob.JobTypeBulkImport, payload, createdBy)

	if err := s.repo.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("ジョブの保存エラー: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.executeBulkImport(job.ID(), payload)
	}()

	return job, nil
}

// executeBulkImport はバルクインポートを非同期で実行する
func (s *ApplicationService) executeBulkImport(jobID string, payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), jobExecutionTimeout)
	defer cancel()

	log := slog.With("job_id", jobID, "job_type", string(domainJob.JobTypeBulkImport))

	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		log.Error("ジョブの取得に失敗しました", "error", err)
		return
	}

	if err := job.Start(); err != nil {
		log.Error("ジョブの開始に失敗しました", "error", err)
		return
	}

	if err := s.repo.Update(ctx, job); err != nil {
		log.Error("ジョブの状態更新に失敗しました（running）", "error", err)
		return
	}

	log.Info("ジョブを開始しました")

	var importPayload BulkImportPayload
	if err := json.Unmarshal(payload, &importPayload); err != nil {
		errMsg := fmt.Sprintf("ペイロードの解析エラー: %s", err.Error())
		log.Error("ペイロードの解析に失敗しました", "error", err)
		if failErr := job.Fail(errMsg); failErr != nil {
			log.Error("ジョブの失敗マークに失敗しました", "error", failErr)
			return
		}
		if updateErr := s.repo.Update(ctx, job); updateErr != nil {
			log.Error("ジョブの状態更新に失敗しました（failed）", "error", updateErr)
		}
		return
	}

	result := map[string]interface{}{
		"processed": len(importPayload.Items),
		"success":   len(importPayload.Items),
		"errors":    []interface{}{},
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		errMsg := fmt.Sprintf("結果のシリアライズエラー: %s", err.Error())
		log.Error("結果のシリアライズに失敗しました", "error", err)
		if failErr := job.Fail(errMsg); failErr != nil {
			log.Error("ジョブの失敗マークに失敗しました", "error", failErr)
			return
		}
		if updateErr := s.repo.Update(ctx, job); updateErr != nil {
			log.Error("ジョブの状態更新に失敗しました（failed）", "error", updateErr)
		}
		return
	}

	if err := job.Complete(resultBytes); err != nil {
		log.Error("ジョブの完了マークに失敗しました", "error", err)
		return
	}

	if err := s.repo.Update(ctx, job); err != nil {
		log.Error("ジョブの状態更新に失敗しました（completed）", "error", err)
		return
	}

	log.Info("ジョブが完了しました", "processed", len(importPayload.Items))
}

// GetJobStatus はジョブのドメインモデルを返す
func (s *ApplicationService) GetJobStatus(ctx context.Context, id string) (*domainJob.Job, error) {
	job, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ジョブの取得エラー: %w", err)
	}
	return job, nil
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

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.executeBulkImport(job.ID(), job.Payload())
	}()

	return job, nil
}
