package event

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/kuro48/idol-api/internal/domain/event"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type eventRepoStub struct {
	data map[string]*domain.Event
}

func newEventRepoStub() *eventRepoStub {
	return &eventRepoStub{data: make(map[string]*domain.Event)}
}

func (r *eventRepoStub) Save(_ context.Context, event *domain.Event) error {
	r.data[event.ID().Value()] = event
	return nil
}

func (r *eventRepoStub) FindByID(_ context.Context, id domain.EventID) (*domain.Event, error) {
	event, ok := r.data[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return event, nil
}

func (r *eventRepoStub) Search(context.Context, domain.SearchCriteria) ([]*domain.Event, error) {
	return nil, nil
}

func (r *eventRepoStub) Count(context.Context, domain.SearchCriteria) (int64, error) {
	return 0, nil
}

func (r *eventRepoStub) Update(_ context.Context, event *domain.Event) error {
	r.data[event.ID().Value()] = event
	return nil
}

func (r *eventRepoStub) Delete(_ context.Context, id domain.EventID) error {
	delete(r.data, id.Value())
	return nil
}

func (r *eventRepoStub) Restore(context.Context, domain.EventID) error {
	return nil
}

func (r *eventRepoStub) FindUpcoming(context.Context, int) ([]*domain.Event, error) {
	return nil, nil
}

func (r *eventRepoStub) FindByPerformer(context.Context, string, int) ([]*domain.Event, error) {
	return nil, nil
}

type eventWebhookPublisherStub struct {
	calls []struct {
		event   domainWebhook.EventType
		payload interface{}
	}
}

func (p *eventWebhookPublisherStub) Publish(_ context.Context, event domainWebhook.EventType, payload interface{}) error {
	p.calls = append(p.calls, struct {
		event   domainWebhook.EventType
		payload interface{}
	}{event: event, payload: payload})
	return nil
}

func TestApplicationService_PublishesWebhookOnCreateUpdateDeleteAndPerformerChanges(t *testing.T) {
	t.Parallel()

	repo := newEventRepoStub()
	publisher := &eventWebhookPublisherStub{}
	svc := NewApplicationService(repo, publisher)

	created, err := svc.CreateEvent(context.Background(), CreateInput{
		Title:         "単独ライブ",
		EventType:     "live",
		StartDateTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 1)
	assert.Equal(t, domainWebhook.EventEventCreated, publisher.calls[0].event)

	payload, ok := publisher.calls[0].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "単独ライブ", payload["title"])

	newTitle := "更新後ライブ"
	err = svc.UpdateEvent(context.Background(), UpdateInput{ID: created.ID().Value(), Title: &newTitle})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 2)
	assert.Equal(t, domainWebhook.EventEventUpdated, publisher.calls[1].event)

	err = svc.AddPerformer(context.Background(), AddPerformerInput{EventID: created.ID().Value(), PerformerID: "idol-1"})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 3)
	assert.Equal(t, domainWebhook.EventEventUpdated, publisher.calls[2].event)

	err = svc.RemovePerformer(context.Background(), RemovePerformerInput{EventID: created.ID().Value(), PerformerID: "idol-1"})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 4)
	assert.Equal(t, domainWebhook.EventEventUpdated, publisher.calls[3].event)

	err = svc.DeleteEvent(context.Background(), created.ID().Value())
	require.NoError(t, err)
	require.Len(t, publisher.calls, 5)
	assert.Equal(t, domainWebhook.EventEventDeleted, publisher.calls[4].event)

	payload, ok = publisher.calls[4].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
}
