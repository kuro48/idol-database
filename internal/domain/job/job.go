package job

import (
	"errors"
	"time"
)

// JobStatus はジョブのステータス
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// JobType はジョブの種別
type JobType string

const (
	JobTypeBulkImport JobType = "bulk_import"
)

// Job は非同期ジョブのドメインエンティティ
type Job struct {
	id          string
	jobType     JobType
	status      JobStatus
	payload     []byte // JSON payload
	result      []byte // JSON result
	errorMsg    string
	createdBy   string
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
}

// NewJob は新しいジョブを作成する
func NewJob(jobType JobType, payload []byte, createdBy string) *Job {
	return &Job{
		jobType:   jobType,
		status:    JobStatusPending,
		payload:   payload,
		createdBy: createdBy,
		createdAt: time.Now(),
	}
}

// ReconstructJob はデータストアからジョブを再構築する（永続化層用）
func ReconstructJob(
	id string,
	jobType JobType,
	status JobStatus,
	payload []byte,
	result []byte,
	errorMsg string,
	createdBy string,
	createdAt time.Time,
	startedAt *time.Time,
	completedAt *time.Time,
) *Job {
	return &Job{
		id:          id,
		jobType:     jobType,
		status:      status,
		payload:     payload,
		result:      result,
		errorMsg:    errorMsg,
		createdBy:   createdBy,
		createdAt:   createdAt,
		startedAt:   startedAt,
		completedAt: completedAt,
	}
}

// ゲッター

func (j *Job) ID() string {
	return j.id
}

func (j *Job) JobType() JobType {
	return j.jobType
}

func (j *Job) Status() JobStatus {
	return j.status
}

func (j *Job) Payload() []byte {
	return j.payload
}

func (j *Job) Result() []byte {
	return j.result
}

func (j *Job) ErrorMsg() string {
	return j.errorMsg
}

func (j *Job) CreatedBy() string {
	return j.createdBy
}

func (j *Job) CreatedAt() time.Time {
	return j.createdAt
}

func (j *Job) StartedAt() *time.Time {
	return j.startedAt
}

func (j *Job) CompletedAt() *time.Time {
	return j.completedAt
}

// SetID はIDを設定する（永続化後に使用）
func (j *Job) SetID(id string) {
	j.id = id
}

// Start はジョブを実行中状態に移行する
func (j *Job) Start() error {
	if j.status != JobStatusPending {
		return errors.New("実行中または完了済みのジョブは開始できません")
	}
	now := time.Now()
	j.status = JobStatusRunning
	j.startedAt = &now
	return nil
}

// Complete はジョブを完了状態に移行する
func (j *Job) Complete(result []byte) error {
	if j.status != JobStatusRunning {
		return errors.New("実行中状態のジョブのみ完了できます")
	}
	now := time.Now()
	j.status = JobStatusCompleted
	j.result = result
	j.completedAt = &now
	return nil
}

// Fail はジョブを失敗状態に移行する
func (j *Job) Fail(errMsg string) error {
	if j.status != JobStatusRunning && j.status != JobStatusPending {
		return errors.New("実行中または保留中状態のジョブのみ失敗にできます")
	}
	now := time.Now()
	j.status = JobStatusFailed
	j.errorMsg = errMsg
	j.completedAt = &now
	return nil
}

// ResetToPending はジョブを保留状態にリセットする（リトライ用）
func (j *Job) ResetToPending() error {
	if j.status != JobStatusFailed {
		return errors.New("失敗済みジョブのみリトライできます")
	}
	j.status = JobStatusPending
	j.errorMsg = ""
	j.result = nil
	j.startedAt = nil
	j.completedAt = nil
	return nil
}
