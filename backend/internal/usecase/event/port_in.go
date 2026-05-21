package event

import "context"

// EventUseCase はイベントのユースケース Input Port
type EventUseCase interface {
	CreateEvent(ctx context.Context, cmd CreateEventCommand) (*EventDTO, error)
	GetEvent(ctx context.Context, query GetEventQuery) (*EventDTO, error)
	SearchEvents(ctx context.Context, query ListEventsQuery) (*SearchResult, error)
	UpdateEvent(ctx context.Context, cmd UpdateEventCommand) error
	DeleteEvent(ctx context.Context, cmd DeleteEventCommand) error
	AddPerformer(ctx context.Context, cmd AddPerformerCommand) error
	RemovePerformer(ctx context.Context, cmd RemovePerformerCommand) error
	FindUpcoming(ctx context.Context, limit int) ([]*EventDTO, error)
}
