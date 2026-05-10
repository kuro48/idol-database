package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthChecker は readiness 判定に必要な契約。
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler はヘルスチェックを扱うハンドラー。
type HealthHandler struct {
	checker HealthChecker
}

// HealthResponse はヘルスチェックレスポンス。
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewHealthHandler は HealthHandler を作成する。
func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// Live は liveness probe を返す。
// @Summary      Liveness check
// @Description  プロセスが起動しているかを確認する
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse
// @Router       /health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

// Ready は readiness probe を返す。
// @Summary      Readiness check
// @Description  依存先が利用可能かを確認する
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse
// @Failure      503 {object} HealthResponse
// @Router       /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	if err := h.checker.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status: "unavailable",
			Error:  "database unreachable",
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

// Health は後方互換用の総合ヘルスチェックを返す。
// @Summary      Health check
// @Description  後方互換のための総合ヘルスチェック
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Message: "Idol API is running with DDD architecture",
	})
}
