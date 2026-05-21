package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	domainAnalytics "github.com/kuro48/idol-api/internal/domain/analytics"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// AnalyticsService はanalytics applicationサービスのインターフェース
type AnalyticsService interface {
	GetUsageSummary(ctx context.Context, days int) ([]*domainAnalytics.KeyUsageSummary, error)
}

// AnalyticsHandler はAPI利用分析ハンドラー
type AnalyticsHandler struct {
	svc AnalyticsService
}

// NewAnalyticsHandler はAnalyticsHandlerを作成する
func NewAnalyticsHandler(svc AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// KeyUsageSummaryResponse はAPIキー単位の利用サマリーレスポンス
type KeyUsageSummaryResponse struct {
	MaskedKey     string  `json:"masked_key"`
	TotalRequests int64   `json:"total_requests"`
	SuccessCount  int64   `json:"success_count"`
	ErrorCount    int64   `json:"error_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	LastUsedAt    string  `json:"last_used_at"`
}

// GetUsageSummary はAPIキー単位の利用サマリーを返す
// @Summary      API利用サマリー取得
// @Description  APIキー単位のAPI利用統計を返す（管理者専用）
// @Tags         admin
// @Produce      json
// @Param        days query int false "集計期間（日数、デフォルト7、最大90）" default(7)
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/analytics/usage [get]
func (h *AnalyticsHandler) GetUsageSummary(c *gin.Context) {
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		parsed, err := strconv.Atoi(daysStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("days パラメータは正の整数である必要があります"))
			return
		}
		if parsed > 90 {
			c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("days パラメータは最大90です"))
			return
		}
		days = parsed
	}

	summaries, err := h.svc.GetUsageSummary(c.Request.Context(), days)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "API利用統計", Message: "API利用統計の取得に失敗しました"})
		return
	}

	responses := make([]KeyUsageSummaryResponse, 0, len(summaries))
	for _, s := range summaries {
		responses = append(responses, KeyUsageSummaryResponse{
			MaskedKey:     s.MaskedKey,
			TotalRequests: s.TotalRequests,
			SuccessCount:  s.SuccessCount,
			ErrorCount:    s.ErrorCount,
			AvgLatencyMs:  s.AvgLatencyMs,
			LastUsedAt:    s.LastUsedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"total": len(responses),
		"days":  days,
	})
}
