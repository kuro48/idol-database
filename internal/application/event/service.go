package event

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/kuro48/idol-api/internal/domain/event"
)

// ApplicationService はイベントアプリケーションサービス
type ApplicationService struct {
	repository event.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repository event.Repository) *ApplicationService {
	return &ApplicationService{
		repository: repository,
	}
}

// CreateEvent はイベントを作成する
func (s *ApplicationService) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*EventDTO, error) {
	// 値オブジェクトの生成
	title, err := event.NewEventTitle(cmd.Title)
	if err != nil {
		return nil, fmt.Errorf("タイトルの生成エラー: %w", err)
	}

	eventType, err := event.NewEventType(cmd.EventType)
	if err != nil {
		return nil, fmt.Errorf("イベントタイプの生成エラー: %w", err)
	}

	startDateTime, err := time.Parse(time.RFC3339, cmd.StartDateTime)
	if err != nil {
		return nil, fmt.Errorf("開始日時のパースエラー: %w", err)
	}

	// エンティティの生成
	newEvent, err := event.NewEvent(title, eventType, startDateTime)
	if err != nil {
		return nil, fmt.Errorf("イベントの生成エラー: %w", err)
	}

	// IDを生成
	id, err := event.NewEventID(generateID())
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}
	newEvent.SetID(id)

	// 終了日時の設定
	var endDateTime *time.Time
	if cmd.EndDateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *cmd.EndDateTime)
		if err == nil {
			endDateTime = &parsed
		}
	}

	// 詳細情報の設定
	newEvent.UpdateDetails(
		nil,
		nil,
		endDateTime,
		cmd.VenueID,
		cmd.TicketURL,
		cmd.OfficialURL,
		cmd.Description,
	)

	// パフォーマーの追加
	for _, performerID := range cmd.PerformerIDs {
		if err := newEvent.AddPerformer(performerID); err != nil {
			return nil, fmt.Errorf("パフォーマー追加エラー: %w", err)
		}
	}

	// タグの追加
	for _, tag := range cmd.Tags {
		if err := newEvent.AddTag(tag); err != nil {
			return nil, fmt.Errorf("タグ追加エラー: %w", err)
		}
	}

	// 保存
	if err := s.repository.Save(ctx, newEvent); err != nil {
		return nil, fmt.Errorf("イベントの保存エラー: %w", err)
	}

	return s.toDTO(newEvent), nil
}

// GetEvent はイベントを取得する
func (s *ApplicationService) GetEvent(ctx context.Context, query GetEventQuery) (*EventDTO, error) {
	id, err := event.NewEventID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("イベントの取得エラー: %w", err)
	}

	return s.toDTO(foundEvent), nil
}

// SearchEvents は条件を指定してイベントを検索する（並行処理版）
func (s *ApplicationService) SearchEvents(ctx context.Context, query ListEventsQuery) (*SearchResult, error) {
	// SearchCriteriaに変換
	criteria := s.queryToCriteria(query)

	// 並行処理: データ取得と件数取得を同時実行
	var events []*event.Event
	var total int64
	var errSearch, errCount error

	var wg sync.WaitGroup
	wg.Add(2)

	// データ取得
	go func() {
		defer wg.Done()
		events, errSearch = s.repository.Search(ctx, criteria)
	}()

	// 総件数取得
	go func() {
		defer wg.Done()
		total, errCount = s.repository.Count(ctx, criteria)
	}()

	wg.Wait()

	// エラーチェック
	if errSearch != nil {
		return nil, fmt.Errorf("検索エラー: %w", errSearch)
	}
	if errCount != nil {
		return nil, fmt.Errorf("件数取得エラー: %w", errCount)
	}

	// DTOに変換
	dtos := make([]*EventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, s.toDTO(e))
	}

	// ページネーション情報を計算
	meta := s.calculatePaginationMeta(total, *query.Page, *query.Limit)

	// ページネーションリンクを生成
	links := s.generatePaginationLinks(query, meta.TotalPages)

	return &SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: links,
	}, nil
}

// UpdateEvent はイベントを更新する
func (s *ApplicationService) UpdateEvent(ctx context.Context, cmd UpdateEventCommand) error {
	id, err := event.NewEventID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("イベントの取得エラー: %w", err)
	}

	// タイトルの更新
	var newTitle *event.EventTitle
	if cmd.Title != nil {
		title, err := event.NewEventTitle(*cmd.Title)
		if err != nil {
			return fmt.Errorf("タイトルの生成エラー: %w", err)
		}
		newTitle = &title
	}

	// 開始日時の更新
	var startDateTime *time.Time
	if cmd.StartDateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *cmd.StartDateTime)
		if err != nil {
			return fmt.Errorf("開始日時のパースエラー: %w", err)
		}
		startDateTime = &parsed
	}

	// 終了日時の更新
	var endDateTime *time.Time
	if cmd.EndDateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *cmd.EndDateTime)
		if err != nil {
			return fmt.Errorf("終了日時のパースエラー: %w", err)
		}
		endDateTime = &parsed
	}

	// 更新
	existingEvent.UpdateDetails(
		newTitle,
		startDateTime,
		endDateTime,
		cmd.VenueID,
		cmd.TicketURL,
		cmd.OfficialURL,
		cmd.Description,
	)

	// 保存
	if err := s.repository.Update(ctx, existingEvent); err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}

	return nil
}

// DeleteEvent はイベントを削除する
func (s *ApplicationService) DeleteEvent(ctx context.Context, cmd DeleteEventCommand) error {
	id, err := event.NewEventID(cmd.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("イベントの削除エラー: %w", err)
	}

	return nil
}

// AddPerformer はパフォーマーを追加する
func (s *ApplicationService) AddPerformer(ctx context.Context, cmd AddPerformerCommand) error {
	id, err := event.NewEventID(cmd.EventID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("イベントの取得エラー: %w", err)
	}

	if err := existingEvent.AddPerformer(cmd.PerformerID); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, existingEvent); err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}

	return nil
}

// RemovePerformer はパフォーマーを削除する
func (s *ApplicationService) RemovePerformer(ctx context.Context, cmd RemovePerformerCommand) error {
	id, err := event.NewEventID(cmd.EventID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("イベントの取得エラー: %w", err)
	}

	existingEvent.RemovePerformer(cmd.PerformerID)

	if err := s.repository.Update(ctx, existingEvent); err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}

	return nil
}

// FindUpcoming は今後開催されるイベントを取得する
func (s *ApplicationService) FindUpcoming(ctx context.Context, limit int) ([]*EventDTO, error) {
	events, err := s.repository.FindUpcoming(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("今後のイベント取得エラー: %w", err)
	}

	dtos := make([]*EventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, s.toDTO(e))
	}

	return dtos, nil
}

// queryToCriteria はListEventsQueryをSearchCriteriaに変換
func (s *ApplicationService) queryToCriteria(query ListEventsQuery) event.SearchCriteria {
	criteria := event.SearchCriteria{
		VenueID:     query.VenueID,
		PerformerID: query.PerformerID,
		Tags:        query.Tags,
		Sort:        *query.Sort,
		Order:       *query.Order,
		Offset:      (*query.Page - 1) * *query.Limit,
		Limit:       *query.Limit,
	}

	// イベントタイプの変換
	if query.EventType != nil {
		eventType, err := event.NewEventType(*query.EventType)
		if err == nil {
			criteria.EventType = &eventType
		}
	}

	// 開始日時範囲の変換
	if query.StartDateFrom != nil {
		if t, err := time.Parse("2006-01-02", *query.StartDateFrom); err == nil {
			criteria.StartDateFrom = &t
		}
	}
	if query.StartDateTo != nil {
		if t, err := time.Parse("2006-01-02", *query.StartDateTo); err == nil {
			// 終了時刻を23:59:59に設定
			endOfDay := t.Add(24*time.Hour - time.Second)
			criteria.StartDateTo = &endOfDay
		}
	}

	return criteria
}

// calculatePaginationMeta はページネーション情報を計算
func (s *ApplicationService) calculatePaginationMeta(total int64, page, perPage int) *PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationMeta{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// generatePaginationLinks はページネーションリンクを生成
func (s *ApplicationService) generatePaginationLinks(query ListEventsQuery, totalPages int) *PaginationLinks {
	baseURL := "/api/v1/events"

	// クエリパラメータを構築
	buildURL := func(page int) string {
		params := url.Values{}
		params.Set("page", strconv.Itoa(page))
		params.Set("limit", strconv.Itoa(*query.Limit))

		if query.EventType != nil {
			params.Set("event_type", *query.EventType)
		}
		if query.StartDateFrom != nil {
			params.Set("start_date_from", *query.StartDateFrom)
		}
		if query.StartDateTo != nil {
			params.Set("start_date_to", *query.StartDateTo)
		}
		if query.VenueID != nil {
			params.Set("venue_id", *query.VenueID)
		}
		if query.PerformerID != nil {
			params.Set("performer_id", *query.PerformerID)
		}
		for _, tag := range query.Tags {
			params.Add("tags", tag)
		}
		if query.Sort != nil {
			params.Set("sort", *query.Sort)
		}
		if query.Order != nil {
			params.Set("order", *query.Order)
		}

		return baseURL + "?" + params.Encode()
	}

	links := &PaginationLinks{
		First: buildURL(1),
		Last:  buildURL(totalPages),
	}

	// 次ページリンク
	if *query.Page < totalPages {
		next := buildURL(*query.Page + 1)
		links.Next = &next
	}

	// 前ページリンク
	if *query.Page > 1 {
		prev := buildURL(*query.Page - 1)
		links.Prev = &prev
	}

	return links
}

// toDTO はドメインモデルをDTOに変換する
func (s *ApplicationService) toDTO(e *event.Event) *EventDTO {
	var endDateTime *string
	if e.EndDateTime() != nil {
		str := e.EndDateTime().Format(time.RFC3339)
		endDateTime = &str
	}

	return &EventDTO{
		ID:            e.ID().Value(),
		Title:         e.Title().Value(),
		EventType:     e.EventType().Value(),
		StartDateTime: e.StartDateTime().Format(time.RFC3339),
		EndDateTime:   endDateTime,
		VenueID:       e.VenueID(),
		PerformerIDs:  e.PerformerIDs(),
		TicketURL:     e.TicketURL(),
		OfficialURL:   e.OfficialURL(),
		Description:   e.Description(),
		Tags:          e.Tags(),
		CreatedAt:     e.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     e.UpdatedAt().Format(time.RFC3339),
	}
}

// generateID はIDを生成する（簡易実装）
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
