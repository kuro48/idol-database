package job_test

import (
	"testing"

	"github.com/kuro48/idol-api/internal/domain/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJob(t *testing.T) {
	payload := []byte(`{"items":[]}`)
	j := job.NewJob(job.JobTypeBulkImport, payload, "user1")

	assert.Equal(t, job.JobTypeBulkImport, j.JobType())
	assert.Equal(t, job.JobStatusPending, j.Status())
	assert.Equal(t, payload, j.Payload())
	assert.Equal(t, "user1", j.CreatedBy())
	assert.Empty(t, j.ID()) // IDは永続化後に設定
	assert.Nil(t, j.StartedAt())
	assert.Nil(t, j.CompletedAt())
}

func TestJob_Start(t *testing.T) {
	t.Run("pending→running に遷移できる", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())

		assert.Equal(t, job.JobStatusRunning, j.Status())
		assert.NotNil(t, j.StartedAt())
	})

	t.Run("running状態からは開始できない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())

		err := j.Start()
		assert.Error(t, err)
		assert.Equal(t, job.JobStatusRunning, j.Status())
	})

	t.Run("completed状態からは開始できない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())
		require.NoError(t, j.Complete([]byte(`{}`)))

		err := j.Start()
		assert.Error(t, err)
	})
}

func TestJob_Complete(t *testing.T) {
	t.Run("running→completed に遷移できる", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())

		result := []byte(`{"processed":3}`)
		require.NoError(t, j.Complete(result))

		assert.Equal(t, job.JobStatusCompleted, j.Status())
		assert.Equal(t, result, j.Result())
		assert.NotNil(t, j.CompletedAt())
	})

	t.Run("pending状態からは完了できない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		err := j.Complete([]byte(`{}`))
		assert.Error(t, err)
		assert.Equal(t, job.JobStatusPending, j.Status())
	})
}

func TestJob_Fail(t *testing.T) {
	t.Run("running→failed に遷移できる", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())

		require.NoError(t, j.Fail("何かエラーが発生しました"))

		assert.Equal(t, job.JobStatusFailed, j.Status())
		assert.Equal(t, "何かエラーが発生しました", j.ErrorMsg())
		assert.NotNil(t, j.CompletedAt())
	})

	t.Run("pending→failed に遷移できる", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Fail("起動エラー"))

		assert.Equal(t, job.JobStatusFailed, j.Status())
	})

	t.Run("completed状態からは失敗にできない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())
		require.NoError(t, j.Complete([]byte(`{}`)))

		err := j.Fail("遅延エラー")
		assert.Error(t, err)
		assert.Equal(t, job.JobStatusCompleted, j.Status())
	})
}

func TestJob_ResetToPending(t *testing.T) {
	t.Run("failed→pending にリセットできる", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, []byte(`{"items":[]}`), "user1")
		require.NoError(t, j.Start())
		require.NoError(t, j.Fail("エラー"))

		require.NoError(t, j.ResetToPending())

		assert.Equal(t, job.JobStatusPending, j.Status())
		assert.Empty(t, j.ErrorMsg())
		assert.Nil(t, j.Result())
		assert.Nil(t, j.StartedAt())
		assert.Nil(t, j.CompletedAt())
	})

	t.Run("pending状態からはリトライできない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		err := j.ResetToPending()
		assert.Error(t, err)
	})

	t.Run("running状態からはリトライできない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())

		err := j.ResetToPending()
		assert.Error(t, err)
	})

	t.Run("completed状態からはリトライできない", func(t *testing.T) {
		j := job.NewJob(job.JobTypeBulkImport, nil, "user1")
		require.NoError(t, j.Start())
		require.NoError(t, j.Complete([]byte(`{}`)))

		err := j.ResetToPending()
		assert.Error(t, err)
	})
}
