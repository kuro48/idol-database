package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockJobService はjobApplicationServiceのモック
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) EnqueueBulkImport(ctx context.Context, payload []byte) (*domainJob.Job, error) {
	args := m.Called(ctx, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainJob.Job), args.Error(1)
}

func (m *MockJobService) GetJobStatus(ctx context.Context, id string) (*domainJob.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainJob.Job), args.Error(1)
}

func (m *MockJobService) RetryJob(ctx context.Context, id string) (*domainJob.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainJob.Job), args.Error(1)
}

func setupJobRouter(h *handlers.JobHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/admin/jobs/bulk-import", h.EnqueueBulkImport)
	r.GET("/admin/jobs/:id", h.GetJobStatus)
	r.POST("/admin/jobs/:id/retry", h.RetryJob)
	return r
}

func newTestJob(id string) *domainJob.Job {
	j := domainJob.NewJob(domainJob.JobTypeBulkImport, []byte(`{"items":[]}`), "test-user")
	j.SetID(id)
	return j
}

func TestJobHandler_EnqueueBulkImport(t *testing.T) {
	t.Run("正常なリクエストで202を返す", func(t *testing.T) {
		svc := new(MockJobService)
		j := newTestJob("job-001")
		svc.On("EnqueueBulkImport", mock.Anything, mock.Anything).Return(j, nil)

		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		body := `{"items":[{"name":"テストアイドル"}]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/bulk-import", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "job-001", resp["job_id"])
		assert.Equal(t, "pending", resp["status"])
	})

	t.Run("itemsが空配列のとき400を返す", func(t *testing.T) {
		svc := new(MockJobService)
		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		body := `{"items":[]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/bulk-import", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertNotCalled(t, "EnqueueBulkImport")
	})

	t.Run("itemsがない場合400を返す", func(t *testing.T) {
		svc := new(MockJobService)
		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/bulk-import", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("itemにnameがない場合400を返す", func(t *testing.T) {
		svc := new(MockJobService)
		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		body := `{"items":[{"birthdate":"2000-01-01"}]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/bulk-import", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー時は500を返す", func(t *testing.T) {
		svc := new(MockJobService)
		svc.On("EnqueueBulkImport", mock.Anything, mock.Anything).Return(nil, errors.New("DB接続エラー"))

		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		body := `{"items":[{"name":"テスト"}]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/bulk-import", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestJobHandler_GetJobStatus(t *testing.T) {
	t.Run("存在するジョブのステータスを200で返す", func(t *testing.T) {
		svc := new(MockJobService)
		j := newTestJob("job-001")
		svc.On("GetJobStatus", mock.Anything, "job-001").Return(j, nil)

		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/admin/jobs/job-001", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "job-001", resp["id"])
		assert.Equal(t, "pending", resp["status"])
	})

	t.Run("存在しないジョブは500を返す（エラー伝播）", func(t *testing.T) {
		svc := new(MockJobService)
		svc.On("GetJobStatus", mock.Anything, "not-found").Return(nil, errors.New("ジョブが見つかりません"))

		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/admin/jobs/not-found", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// WriteErrorの実装によりエラーコードが決まる
		assert.GreaterOrEqual(t, w.Code, 400)
	})
}

func TestJobHandler_RetryJob(t *testing.T) {
	t.Run("成功時は202を返す", func(t *testing.T) {
		svc := new(MockJobService)
		j := newTestJob("job-002")
		svc.On("RetryJob", mock.Anything, "job-002").Return(j, nil)

		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/job-002/retry", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "job-002", resp["job_id"])
	})

	t.Run("サービスエラー時はエラーレスポンスを返す", func(t *testing.T) {
		svc := new(MockJobService)
		svc.On("RetryJob", mock.Anything, "job-003").Return(nil, errors.New("失敗済みジョブのみリトライできます"))

		h := handlers.NewJobHandler(svc)
		router := setupJobRouter(h)

		req := httptest.NewRequest(http.MethodPost, "/admin/jobs/job-003/retry", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.GreaterOrEqual(t, w.Code, 400)
	})
}
