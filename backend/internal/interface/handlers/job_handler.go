package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// JobService はジョブアプリケーションサービスのインターフェース
type JobService interface {
	EnqueueBulkImport(ctx context.Context, payload []byte) (*domainJob.Job, error)
	GetJobStatus(ctx context.Context, id string) (*domainJob.Job, error)
	RetryJob(ctx context.Context, id string) (*domainJob.Job, error)
}

// JobStatusDTO はジョブステータスのレスポンス
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

// JobHandler は非同期ジョブハンドラー
type JobHandler struct {
	svc JobService
}

// NewJobHandler はJobHandlerを作成する
func NewJobHandler(svc JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

// BulkImportItem はバルクインポートの1件分のデータ
type BulkImportItem struct {
	Name      string   `json:"name" binding:"required"`
	Birthdate string   `json:"birthdate,omitempty"` // YYYY-MM-DD
	AgencyID  string   `json:"agency_id,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	TagIDs    []string `json:"tag_ids,omitempty"`
}

// BulkImportRequest はバルクインポートリクエスト
type BulkImportRequest struct {
	Items []BulkImportItem `json:"items" binding:"required,min=1,max=1000,dive"`
}

// EnqueueBulkImport はバルクインポートジョブをエンキューする
// @Summary      バルクインポートジョブ作成
// @Description  バルクインポートを非同期で実行するジョブをキューに追加する（管理者専用）
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body BulkImportRequest true "インポートデータ"
// @Success      202 {object} map[string]interface{}
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/jobs/bulk-import [post]
func (h *JobHandler) EnqueueBulkImport(c *gin.Context) {
	var req BulkImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("ペイロードの変換に失敗しました"))
		return
	}

	job, err := h.svc.EnqueueBulkImport(middleware.AuditContextFor(c), payload)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "ジョブ", Message: "ジョブの作成に失敗しました"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  job.ID(),
		"status":  string(job.Status()),
		"message": "ジョブをキューに追加しました",
	})
}

// GetJobStatus はジョブのステータスを返す
// @Summary      ジョブステータス取得
// @Description  指定したジョブのステータスと結果を返す（管理者専用）
// @Tags         admin
// @Produce      json
// @Param        id path string true "ジョブID"
// @Success      200 {object} JobStatusDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/jobs/{id} [get]
func (h *JobHandler) GetJobStatus(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	job, err := h.svc.GetJobStatus(c.Request.Context(), id)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "ジョブ"})
		return
	}

	c.JSON(http.StatusOK, toJobStatusDTO(job))
}

// RetryJob は失敗したジョブをリトライする
// @Summary      ジョブリトライ
// @Description  失敗したジョブを再実行する（管理者専用）
// @Tags         admin
// @Produce      json
// @Param        id path string true "ジョブID"
// @Success      202 {object} map[string]interface{}
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/jobs/{id}/retry [post]
func (h *JobHandler) RetryJob(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	job, err := h.svc.RetryJob(middleware.AuditContextFor(c), id)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "ジョブ", Message: "ジョブのリトライに失敗しました"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  job.ID(),
		"status":  string(job.Status()),
		"message": "ジョブのリトライをキューに追加しました",
	})
}
