package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/event"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// EventHandler はイベントハンドラー
type EventHandler struct {
	appService *event.ApplicationService
}

// NewEventHandler はイベントハンドラーを作成する
func NewEventHandler(appService *event.ApplicationService) *EventHandler {
	return &EventHandler{
		appService: appService,
	}
}

// CreateEvent はイベントを作成する
// @Summary      イベント作成
// @Description  新しいイベント/ライブを作成する
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        event body event.CreateEventCommand true "イベント作成リクエスト"
// @Success      201 {object} event.EventDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var cmd event.CreateEventCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	dto, err := h.appService.CreateEvent(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("イベントの作成に失敗しました"))
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// GetEvent はイベントを取得する
// @Summary      イベント詳細取得
// @Description  IDを指定してイベント情報を取得する
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id path string true "イベントID"
// @Success      200 {object} event.EventDTO
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      404 {object} middleware.ErrorResponse
// @Router       /events/{id} [get]
func (h *EventHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	query := event.GetEventQuery{ID: id}

	dto, err := h.appService.GetEvent(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, middleware.NewNotFoundError("イベント"))
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ListEvents はイベント一覧を取得する（検索機能付き）
// @Summary      イベント一覧取得
// @Description  条件を指定してイベント一覧を取得（検索・フィルタリング・ページネーション対応）
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        event_type query string false "イベントタイプ" Enums(live, handshake, release, fan_meeting, online)
// @Param        start_date_from query string false "開始日FROM (YYYY-MM-DD)"
// @Param        start_date_to query string false "開始日TO (YYYY-MM-DD)"
// @Param        venue_id query string false "会場ID"
// @Param        performer_id query string false "パフォーマーID"
// @Param        tags query []string false "タグ（複数可）"
// @Param        sort query string false "ソート項目" Enums(start_date_time, created_at) default(start_date_time)
// @Param        order query string false "ソート順" Enums(asc, desc) default(asc)
// @Param        page query int false "ページ番号" default(1)
// @Param        limit query int false "1ページあたりの件数" default(20)
// @Success      200 {object} event.SearchResult
// @Failure      400 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /events [get]
func (h *EventHandler) ListEvents(c *gin.Context) {
	var query event.ListEventsQuery

	// クエリパラメータをバインド
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("無効なクエリパラメータです: "+err.Error()))
		return
	}

	// デフォルト値を適用
	query.ApplyDefaults()

	// バリデーション
	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError(err.Error()))
		return
	}

	// 検索実行
	result, err := h.appService.SearchEvents(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("検索に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateEvent はイベントを更新する
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	var cmd event.UpdateEventCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd.ID = id

	err := h.appService.UpdateEvent(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("イベントの更新に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "イベントが更新されました"})
}

// DeleteEvent はイベントを削除する
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return
	}

	cmd := event.DeleteEventCommand{ID: id}

	err := h.appService.DeleteEvent(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("イベントの削除に失敗しました"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddPerformer はイベントにパフォーマーを追加する
func (h *EventHandler) AddPerformer(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("イベントIDは必須です"))
		return
	}

	var req struct {
		PerformerID string `json:"performer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))
		return
	}

	cmd := event.AddPerformerCommand{
		EventID:     eventID,
		PerformerID: req.PerformerID,
	}

	err := h.appService.AddPerformer(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("パフォーマーの追加に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パフォーマーが追加されました"})
}

// RemovePerformer はイベントからパフォーマーを削除する
func (h *EventHandler) RemovePerformer(c *gin.Context) {
	eventID := c.Param("id")
	performerID := c.Param("performer_id")

	if eventID == "" || performerID == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("イベントIDとパフォーマーIDは必須です"))
		return
	}

	cmd := event.RemovePerformerCommand{
		EventID:     eventID,
		PerformerID: performerID,
	}

	err := h.appService.RemovePerformer(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("パフォーマーの削除に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パフォーマーが削除されました"})
}

// GetUpcomingEvents は今後開催されるイベントを取得する
func (h *EventHandler) GetUpcomingEvents(c *gin.Context) {
	limit := 20 // デフォルト値

	dtos, err := h.appService.FindUpcoming(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("今後のイベント取得に失敗しました"))
		return
	}

	c.JSON(http.StatusOK, dtos)
}
