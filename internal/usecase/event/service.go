package event

import (
	"context"
	"math"
	"net/url"
	"strconv"
	"time"

	app "github.com/kuro48/idol-api/internal/application/event"
	domain "github.com/kuro48/idol-api/internal/domain/event"
)

// Usecase はイベントのユースケース
type Usecase struct {
	appService *app.ApplicationService
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService *app.ApplicationService) *Usecase {
	return &Usecase{appService: appService}
}

// CreateEvent はイベントを作成する
func (u *Usecase) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*EventDTO, error) {
	entity, err := u.appService.CreateEvent(ctx, app.CreateInput{
		Title:         cmd.Title,
		EventType:     cmd.EventType,
		StartDateTime: cmd.StartDateTime,
		EndDateTime:   cmd.EndDateTime,
		VenueID:       cmd.VenueID,
		PerformerIDs:  cmd.PerformerIDs,
		TicketURL:     cmd.TicketURL,
		OfficialURL:   cmd.OfficialURL,
		Description:   cmd.Description,
		Tags:          cmd.Tags,
	})
	if err != nil {
		return nil, err
	}

	dto := toDTO(entity)
	return &dto, nil
}

// GetEvent はイベントを取得する
func (u *Usecase) GetEvent(ctx context.Context, query GetEventQuery) (*EventDTO, error) {
	entity, err := u.appService.GetEvent(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	dto := toDTO(entity)
	return &dto, nil
}

// SearchEvents は条件を指定してイベントを検索する
func (u *Usecase) SearchEvents(ctx context.Context, query ListEventsQuery) (*SearchResult, error) {
	criteria := u.queryToCriteria(query)

	events, total, err := u.appService.SearchEvents(ctx, criteria)
	if err != nil {
		return nil, err
	}

	dtos := make([]*EventDTO, 0, len(events))
	for _, e := range events {
		dto := toDTO(e)
		dtos = append(dtos, &dto)
	}

	meta := u.calculatePaginationMeta(total, *query.Page, *query.Limit)
	links := u.generatePaginationLinks(query, meta.TotalPages)

	return &SearchResult{
		Data:  dtos,
		Meta:  meta,
		Links: links,
	}, nil
}

// UpdateEvent はイベントを更新する
func (u *Usecase) UpdateEvent(ctx context.Context, cmd UpdateEventCommand) error {
	return u.appService.UpdateEvent(ctx, app.UpdateInput{
		ID:            cmd.ID,
		Title:         cmd.Title,
		StartDateTime: cmd.StartDateTime,
		EndDateTime:   cmd.EndDateTime,
		VenueID:       cmd.VenueID,
		TicketURL:     cmd.TicketURL,
		OfficialURL:   cmd.OfficialURL,
		Description:   cmd.Description,
	})
}

// DeleteEvent はイベントを削除する
func (u *Usecase) DeleteEvent(ctx context.Context, cmd DeleteEventCommand) error {
	return u.appService.DeleteEvent(ctx, cmd.ID)
}

// AddPerformer はパフォーマーを追加する
func (u *Usecase) AddPerformer(ctx context.Context, cmd AddPerformerCommand) error {
	return u.appService.AddPerformer(ctx, app.AddPerformerInput{
		EventID:     cmd.EventID,
		PerformerID: cmd.PerformerID,
	})
}

// RemovePerformer はパフォーマーを削除する
func (u *Usecase) RemovePerformer(ctx context.Context, cmd RemovePerformerCommand) error {
	return u.appService.RemovePerformer(ctx, app.RemovePerformerInput{
		EventID:     cmd.EventID,
		PerformerID: cmd.PerformerID,
	})
}

// FindUpcoming は今後開催されるイベントを取得する
func (u *Usecase) FindUpcoming(ctx context.Context, limit int) ([]*EventDTO, error) {
	events, err := u.appService.FindUpcoming(ctx, limit)
	if err != nil {
		return nil, err
	}

	dtos := make([]*EventDTO, 0, len(events))
	for _, e := range events {
		dto := toDTO(e)
		dtos = append(dtos, &dto)
	}

	return dtos, nil
}

// queryToCriteria はListEventsQueryをSearchCriteriaに変換
func (u *Usecase) queryToCriteria(query ListEventsQuery) domain.SearchCriteria {
	criteria := domain.SearchCriteria{
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
		eventType, err := domain.NewEventType(*query.EventType)
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
			endOfDay := t.Add(24*time.Hour - time.Second)
			criteria.StartDateTo = &endOfDay
		}
	}

	return criteria
}

// calculatePaginationMeta はページネーション情報を計算
func (u *Usecase) calculatePaginationMeta(total int64, page, perPage int) *PaginationMeta {
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
func (u *Usecase) generatePaginationLinks(query ListEventsQuery, totalPages int) *PaginationLinks {
	baseURL := "/api/v1/events"

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

	if *query.Page < totalPages {
		next := buildURL(*query.Page + 1)
		links.Next = &next
	}

	if *query.Page > 1 {
		prev := buildURL(*query.Page - 1)
		links.Prev = &prev
	}

	return links
}

// toDTO はドメインモデルをDTOに変換する
func toDTO(e *domain.Event) EventDTO {
	var endDateTime *string
	if e.EndDateTime() != nil {
		str := e.EndDateTime().Format(time.RFC3339)
		endDateTime = &str
	}

	return EventDTO{
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
