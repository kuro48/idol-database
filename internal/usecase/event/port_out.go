package event

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/event"
)

// EventAppPort は event.Usecase が event application サービスに要求する契約
type EventAppPort interface {
	CreateEvent(ctx context.Context, input EventCreateInput) (*domain.Event, error)
	GetEvent(ctx context.Context, id string) (*domain.Event, error)
	SearchEvents(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Event, int64, error)
	UpdateEvent(ctx context.Context, input EventUpdateInput) error
	DeleteEvent(ctx context.Context, id string) error
	AddPerformer(ctx context.Context, input EventAddPerformerInput) error
	RemovePerformer(ctx context.Context, input EventRemovePerformerInput) error
	FindUpcoming(ctx context.Context, limit int) ([]*domain.Event, error)
}

// EventCreateInput はイベント作成の入力
type EventCreateInput struct {
	Title         string
	EventType     string
	StartDateTime string
	EndDateTime   *string
	VenueID       *string
	PerformerIDs  []string
	TicketURL     *string
	OfficialURL   *string
	Description   *string
	Tags          []string
}

// EventUpdateInput はイベント更新の入力
type EventUpdateInput struct {
	ID            string
	Title         *string
	StartDateTime *string
	EndDateTime   *string
	VenueID       *string
	TicketURL     *string
	OfficialURL   *string
	Description   *string
}

// EventAddPerformerInput はパフォーマー追加の入力
type EventAddPerformerInput struct {
	EventID     string
	PerformerID string
}

// EventRemovePerformerInput はパフォーマー削除の入力
type EventRemovePerformerInput struct {
	EventID     string
	PerformerID string
}
