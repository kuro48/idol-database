package adapters

import (
	"context"

	appEvent "github.com/kuro48/idol-api/internal/application/event"
	eventDomain "github.com/kuro48/idol-api/internal/domain/event"
	ucEvent "github.com/kuro48/idol-api/internal/usecase/event"
)

// EventAppAdapter は appEvent.ApplicationService を ucEvent.EventAppPort に適合させる
type EventAppAdapter struct {
	svc *appEvent.ApplicationService
}

// NewEventAppAdapter は EventAppAdapter を生成する
func NewEventAppAdapter(svc *appEvent.ApplicationService) ucEvent.EventAppPort {
	return &EventAppAdapter{svc: svc}
}

func (a *EventAppAdapter) CreateEvent(ctx context.Context, input ucEvent.EventCreateInput) (*eventDomain.Event, error) {
	performers := make([]appEvent.PerformerInput, 0, len(input.Performers))
	for _, p := range input.Performers {
		performers = append(performers, appEvent.PerformerInput{
			PerformerID:   p.PerformerID,
			BillingStatus: p.BillingStatus,
		})
	}
	return a.svc.CreateEvent(ctx, appEvent.CreateInput{
		Title:         input.Title,
		EventType:     input.EventType,
		StartDateTime: input.StartDateTime,
		EndDateTime:   input.EndDateTime,
		VenueID:       input.VenueID,
		Performers:    performers,
		TicketURL:     input.TicketURL,
		OfficialURL:   input.OfficialURL,
		Description:   input.Description,
		Tags:          input.Tags,
	})
}

func (a *EventAppAdapter) GetEvent(ctx context.Context, id string) (*eventDomain.Event, error) {
	return a.svc.GetEvent(ctx, id)
}

func (a *EventAppAdapter) SearchEvents(ctx context.Context, criteria eventDomain.SearchCriteria) ([]*eventDomain.Event, int64, error) {
	return a.svc.SearchEvents(ctx, criteria)
}

func (a *EventAppAdapter) UpdateEvent(ctx context.Context, input ucEvent.EventUpdateInput) error {
	return a.svc.UpdateEvent(ctx, appEvent.UpdateInput{
		ID:            input.ID,
		Title:         input.Title,
		StartDateTime: input.StartDateTime,
		EndDateTime:   input.EndDateTime,
		VenueID:       input.VenueID,
		TicketURL:     input.TicketURL,
		OfficialURL:   input.OfficialURL,
		Description:   input.Description,
	})
}

func (a *EventAppAdapter) DeleteEvent(ctx context.Context, id string) error {
	return a.svc.DeleteEvent(ctx, id)
}

func (a *EventAppAdapter) AddPerformer(ctx context.Context, input ucEvent.EventAddPerformerInput) error {
	return a.svc.AddPerformer(ctx, appEvent.AddPerformerInput{
		EventID:       input.EventID,
		PerformerID:   input.PerformerID,
		BillingStatus: input.BillingStatus,
	})
}

func (a *EventAppAdapter) RemovePerformer(ctx context.Context, input ucEvent.EventRemovePerformerInput) error {
	return a.svc.RemovePerformer(ctx, appEvent.RemovePerformerInput{
		EventID:     input.EventID,
		PerformerID: input.PerformerID,
	})
}

func (a *EventAppAdapter) FindUpcoming(ctx context.Context, limit int) ([]*eventDomain.Event, error) {
	return a.svc.FindUpcoming(ctx, limit)
}
