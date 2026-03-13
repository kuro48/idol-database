package job_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appJob "github.com/kuro48/idol-api/internal/application/job"
	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockJobRepository はdomainJob.Repositoryのモック
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Save(ctx context.Context, j *domainJob.Job) error {
	args := m.Called(ctx, j)
	if args.Error(0) == nil {
		j.SetID("test-job-id")
	}
	return args.Error(0)
}

func (m *MockJobRepository) FindByID(ctx context.Context, id string) (*domainJob.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainJob.Job), args.Error(1)
}

func (m *MockJobRepository) Update(ctx context.Context, j *domainJob.Job) error {
	args := m.Called(ctx, j)
	return args.Error(0)
}

func (m *MockJobRepository) FindByStatus(ctx context.Context, status domainJob.JobStatus, limit int) ([]*domainJob.Job, error) {
	args := m.Called(ctx, status, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainJob.Job), args.Error(1)
}

// newRunningJob はテスト用の実行中ジョブを作成する
func newRunningJob(id string) *domainJob.Job {
	j := domainJob.NewJob(domainJob.JobTypeBulkImport, []byte(`{"items":[{"name":"テスト"}]}`), "user1")
	j.SetID(id)
	_ = j.Start()
	return j
}

// newPendingJob はテスト用の保留中ジョブを作成する
func newPendingJob(id string) *domainJob.Job {
	j := domainJob.NewJob(domainJob.JobTypeBulkImport, []byte(`{"items":[{"name":"テスト"}]}`), "user1")
	j.SetID(id)
	return j
}

func TestApplicationService_EnqueueBulkImport(t *testing.T) {
	t.Run("ジョブを正常に作成してIDを返す", func(t *testing.T) {
		repo := new(MockJobRepository)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil)
		// executeBulkImport 内のFindByID/Updateも許可
		runningJob := newRunningJob("test-job-id")
		repo.On("FindByID", mock.Anything, "test-job-id").Return(runningJob, nil).Maybe()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()

		svc := appJob.NewApplicationService(repo)
		payload := []byte(`{"items":[{"name":"アイドル1"}]}`)
		j, err := svc.EnqueueBulkImport(context.Background(), payload)

		require.NoError(t, err)
		assert.NotNil(t, j)
		assert.Equal(t, "test-job-id", j.ID())
		assert.Equal(t, domainJob.JobStatusPending, j.Status())
	})

	t.Run("Save失敗時はエラーを返す", func(t *testing.T) {
		repo := new(MockJobRepository)
		repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("DB接続エラー"))

		svc := appJob.NewApplicationService(repo)
		j, err := svc.EnqueueBulkImport(context.Background(), []byte(`{"items":[]}`))

		assert.Error(t, err)
		assert.Nil(t, j)
	})
}

func TestApplicationService_GetJobStatus(t *testing.T) {
	t.Run("存在するジョブのステータスを返す", func(t *testing.T) {
		repo := new(MockJobRepository)
		j := newPendingJob("job-123")
		repo.On("FindByID", mock.Anything, "job-123").Return(j, nil)

		svc := appJob.NewApplicationService(repo)
		dto, err := svc.GetJobStatus(context.Background(), "job-123")

		require.NoError(t, err)
		require.NotNil(t, dto)
		assert.Equal(t, "job-123", dto.ID())
		assert.Equal(t, domainJob.JobStatusPending, dto.Status())
		assert.Equal(t, domainJob.JobTypeBulkImport, dto.JobType())
	})

	t.Run("存在しないジョブはエラーを返す", func(t *testing.T) {
		repo := new(MockJobRepository)
		repo.On("FindByID", mock.Anything, "not-found").Return(nil, errors.New("ジョブが見つかりません"))

		svc := appJob.NewApplicationService(repo)
		dto, err := svc.GetJobStatus(context.Background(), "not-found")

		assert.Error(t, err)
		assert.Nil(t, dto)
	})

	t.Run("completed状態のジョブはResultを含む", func(t *testing.T) {
		repo := new(MockJobRepository)
		j := newPendingJob("job-456")
		_ = j.Start()
		_ = j.Complete([]byte(`{"processed":5,"success":5}`))
		repo.On("FindByID", mock.Anything, "job-456").Return(j, nil)

		svc := appJob.NewApplicationService(repo)
		dto, err := svc.GetJobStatus(context.Background(), "job-456")

		require.NoError(t, err)
		require.NotNil(t, dto)
		assert.Equal(t, domainJob.JobStatusCompleted, dto.Status())
		assert.NotEmpty(t, dto.Result())
		assert.Contains(t, string(dto.Result()), "processed")
		assert.NotNil(t, dto.StartedAt())
		assert.NotNil(t, dto.CompletedAt())
	})
}

func TestApplicationService_RetryJob(t *testing.T) {
	t.Run("失敗したジョブをリトライできる", func(t *testing.T) {
		repo := new(MockJobRepository)
		failedJob := newPendingJob("job-789")
		_ = failedJob.Start()
		_ = failedJob.Fail("エラーが発生しました")

		repo.On("FindByID", mock.Anything, "job-789").Return(failedJob, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		// executeBulkImport 内での呼び出し（非同期）も許可
		repo.On("FindByID", mock.Anything, "job-789").Return(newRunningJob("job-789"), nil).Maybe()

		svc := appJob.NewApplicationService(repo)
		j, err := svc.RetryJob(context.Background(), "job-789")

		require.NoError(t, err)
		assert.NotNil(t, j)
		// ResetToPending後はpendingになっている
		assert.Equal(t, domainJob.JobStatusPending, j.Status())

		// goroutineが非同期で動くので少し待つ
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("pending状態のジョブはリトライできない", func(t *testing.T) {
		repo := new(MockJobRepository)
		pendingJob := newPendingJob("job-abc")
		repo.On("FindByID", mock.Anything, "job-abc").Return(pendingJob, nil)

		svc := appJob.NewApplicationService(repo)
		j, err := svc.RetryJob(context.Background(), "job-abc")

		assert.Error(t, err)
		assert.Nil(t, j)
	})

	t.Run("ジョブが見つからない場合はエラーを返す", func(t *testing.T) {
		repo := new(MockJobRepository)
		repo.On("FindByID", mock.Anything, "not-found").Return(nil, errors.New("ジョブが見つかりません"))

		svc := appJob.NewApplicationService(repo)
		j, err := svc.RetryJob(context.Background(), "not-found")

		assert.Error(t, err)
		assert.Nil(t, j)
	})
}
