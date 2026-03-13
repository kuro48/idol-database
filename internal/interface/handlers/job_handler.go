package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	appJob "github.com/kuro48/idol-api/internal/application/job"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// JobHandler は非同期ジョブハンドラー
type JobHandler struct {
	svc *appJob.ApplicationService
}

// NewJobHandler はJobHandlerを作成する
func NewJobHandler(svc *appJob.ApplicationService) *JobHandler {
	return &JobHandler{svc: svc}
}

// BulkImportRequest はバルクインポートリクエスト
type BulkImportRequest struct {
	Items []map[string]interface{} `json:"items" binding:"required,min=1"`
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
// @Success      200 {object} appJob.JobStatusDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/jobs/{id} [get]
func (h *JobHandler) GetJobStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	dto, err := h.svc.GetJobStatus(c.Request.Context(), id)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "ジョブ"})
		return
	}

	c.JSON(http.StatusOK, dto)
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
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
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
