package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	ucEditHistory "github.com/kuro48/idol-api/internal/usecase/edithistory"
)

// EditHistoryHandler は編集履歴ハンドラー
type EditHistoryHandler struct {
	usecase *ucEditHistory.Usecase
}

// NewEditHistoryHandler は編集履歴ハンドラーを作成する
func NewEditHistoryHandler(usecase *ucEditHistory.Usecase) *EditHistoryHandler {
	return &EditHistoryHandler{usecase: usecase}
}

// GetEditHistory は編集履歴を取得する
func (h *EditHistoryHandler) GetEditHistory(c *gin.Context) {
	id, ok := getPathID(c)
	if !ok {
		return
	}

	dto, err := h.usecase.GetEditHistory(c.Request.Context(), ucEditHistory.GetEditHistoryQuery{ID: id})
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Resource: "編集履歴"})
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListEditHistory は編集履歴一覧を取得する
func (h *EditHistoryHandler) ListEditHistory(c *gin.Context) {
	var query ucEditHistory.ListEditHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです: "+err.Error()))
		return
	}

	query.ApplyDefaults()

	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.usecase.ListEditHistory(c.Request.Context(), query)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "編集履歴の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, result)
}
